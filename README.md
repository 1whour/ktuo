## scheduler
scheduler是分布式调度框架，主要功能定时，延时，lambda等功能。可以基于DAG组织任务。

## 进展
开发中。。。

## 主要特性
* 内置服务注册与发现。实现节点的水平扩张, 依赖etcd
* 支持http, shell, grpc(TODO)
* 支持lambda，基于runtime级别多语言扩展能力, 性能可观
* 大量的测试，让bug少之又少
* DAG支持(TODO)

## 架构图

## 快速开始
### 一、http任务配置
```yaml
apiVersion: v0.0.1 #api版本号
kind: oneRuntime #只在一个runtime上运行
trigger:
  cron: "* * * * * *"
executer:
  taskName: first-task
  http:
    method: post
    scheme: http
    host: 127.0.0.1
    port: 8080
    headers:
    - name: Bid
      value: xxxx
    - name: token
      value: vvvv
    body: |
      {"a":"b"}
```
### 1. 创建任务
```bash
 ./scheduler start -f ./example/http.yaml -g gate_addr
```
### 2. 删除任务
```bash
./scheduler rm -f ./example/http.yaml -g gate_addr
```
### 3. 停止任务
```bash
./scheduler stop -f ./example/http.yaml -g gate_addr
```
### 4. 更新任务
```bash
./scheduler update -f ./example/http.yaml -g gate_addr
```

### 二、shell任务配置
```yaml
apiVersion: v0.0.1
kind: oneRuntime
trigger:
  cron: "* * * * * *" #每秒触发一次
executer:
  shell:
    command: curl -X POST 127.0.0.1:8181/task #command和args的作用是等价的，唯一的区别是命令放在一个字符串或者slice里面。
    #args:
    #- echo
    #- "hello"
```

### 三、其它命令
```bash
scheduler start 配置文件. #创建新的dag任务，并且运行
scheduler stop 配置文件. #停止dag任务
scheduler rm 配置文件. #删除dag任务
scheduler run 配置文件. #运行已存在的任务，如果不存在会返回错误
scheduler status 获取任务的状态
```
