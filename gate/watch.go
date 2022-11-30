package gate

import (
	"encoding/json"
	"strings"

	"github.com/gnh123/ktuo/model"
	"github.com/gnh123/ktuo/utils"
	"github.com/gorilla/websocket"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func (r *Gate) watchLocalRunq(runtimeName string, conn *websocket.Conn) {
	// 生成本地队列的前缀
	localPath := model.WatchLocalRuntimePrefix(runtimeName)
	// watch本地队列的任务
	localTask := defautlClient.Watch(r.ctx, localPath, clientv3.WithPrefix())

	r.Debug().Msgf(">>> watch local:%s\n", localPath)
	for ersp := range localTask {
		for _, ev := range ersp.Events {
			r.Debug().Msgf("watchLocalRunq create(%t) modify(%t) delete(%t), key(%s), value(%s)\n",
				ev.IsCreate(), ev.IsModify(), ev.Type == clientv3.EventTypeDelete, ev.Kv.Key, ev.Kv.Value)

			// 本地队列全名
			localKey := string(ev.Kv.Key)
			// 提取task名
			taskName := model.TaskName(localKey)
			// 生成全局队列名
			globalKey := model.ToGlobalTask(localKey)
			// 获取全局队列里面的task配置信息
			rsp, err := defaultKVC.Get(r.ctx, globalKey)
			if err != nil {
				r.Warn().Msgf("gate.watchLocalRunq: get param %s\n", err)
				continue
			}

			if len(rsp.Kvs) == 0 {
				continue
			}

			value := rsp.Kvs[0].Value

			var param model.Param
			err = json.Unmarshal(value, &param)
			if err != nil {
				r.Warn().Msgf("gate.watchLocalRunq:%s\n", err)
				continue
			}

			switch {
			case ev.IsCreate(), ev.IsModify():
				// 如果是新建或者被修改过的，直接推送到客户端
				// 成功的状态是model.Succeeded, 失败的状态是model.Failed
				if err := utils.WriteMessageTimeout(conn, value, r.WriteTime); err != nil {
					r.Warn().Msgf("gate.watchLocalRunq, WriteMessageTimeout :%s, runtimeName:%s bye bye, taskName(%s)\n", err, runtimeName, taskName)
					// 更新全局状态, 修改为失败标志
					defaultStore.LockUnlock(r.ctx, taskName, func() error {
						err := defaultStore.UpdateCallStateFailed(r.ctx, taskName)
						if err != nil {
							r.Error().Msgf("gate.watchLocalRunq, write failed ack fail %s, runtimeName:%s bye bye, taskName(%s)\n", err, runtimeName, taskName)
							return err
						}
						r.delRuntimeNode(model.Whoami{Name: runtimeName, Lambda: strings.Contains(localKey, model.LambdaKey)})
						return nil
					})
				}

				if param.IsRemove() {
					defaultKVC.Delete(r.ctx, globalKey)
					defaultKVC.Delete(r.ctx, localKey)
					defaultKVC.Delete(r.ctx, model.FullGlobalTaskState(taskName)) //删除本地队列
				} else {
					// 更新全局状态, 修改为成功标志
					if err := defaultStore.LockUpdateCallStateSuccessed(r.ctx, taskName); err != nil {
						r.Error().Msgf("gate.watchLocalRunq, write successed ack fail %s, runtimeName:%s taskName(%s)\n", err, runtimeName, taskName)

					}
				}
			case ev.Type == clientv3.EventTypeDelete:
				r.Debug().Msgf("delete global task:%s, state:%s\n", ev.Kv.Key, ev.Kv.Value)
			}
		}
	}
}
