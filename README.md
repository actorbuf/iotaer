### iotaer

### 安装

```shell
$ go install github.com/actorbuf/iotaer@latest
## output
go: downloading github.com/actorbuf/iotaer v0.0.0-20210902040340-7af67c79f914
```

安装完毕后执行 `iotaer -h` 验证是否安装成功

```shell
[iotaer@iotaer iotaer]$ iotaer -h
iotaer 是基于 omega 库的一个提高生产效率的工具链

Usage:
  iotaer [command]

Available Commands:
  addapi                 给路由组新增一个api
  addroute               快速添加一个路由组
  addrpc                 给service新增一个rpc
  auth                   进行gitlab授权认证, 否则部份授权操作无法进行
  create                 创建一个新项目
  dep                    更新iotaer依赖的工具链
  fmt                    格式化 proto 文件使其看的赏心悦目
  gen                    解析proto文件, 自动生成开发代码.
  gen-k8s-deployment-yml 生成k8s deployment yml文件
  gen-k8s-ingress-yml    生成k8s ingress yml文件
  gitter                 拉取git仓库代码并布署
  go                     golang 版本管理
  help                   Help about any command
  md                     给路由组的api输出markdown文档
  repo                   gitlab仓库操作
  run                    快速运行项目
  tool                   将文件构建为任意工具箱
  update                 检测并更新iotaer
  version                打印iotaer版本信息

Flags:
  -h, --help   help for iotaer

Use "iotaer [command] --help" for more information about a command.

```

### iotaer 版本更新

在安装完 `iotaer` 后, 可以使用内置的 `update` 命令进行自更新. 当然, iotaer 内置有检测是否有新版本的机制,当版本过低,将提醒你手动更新

```shell
[iotaer@iotaer iotaer]$ iotaer update
-----
更新 iotaer 插件中...

iotaer 更新完成
-----

```

### iotaer 依赖工具链更新

在安装完 `iotaer` 后, 务必使用内置的 `dep` 命令进行工具链更新.

```shell
[iotaer@iotaer iotaer]$ iotaer dep
检测 protoc 插件中...
protoc local version: v3.18.0
protoc latest version: v3.18.1
正在下载 protoc...
插件将被放置到: /home/x/go/bin/protoc
include文件夹将被放置在: /home/x/go/bin/include
protoc version: libprotoc 3.18.1
安装 protoc完毕
-----
更新 protoc-go-inject-tag 插件中...
protoc-go-inject-tag更新完成
-----
更新 protoc-gen-go 插件中...
protoc-gen-go更新完成
-----
更新 protoc-gen-go-grpc 插件中...

protoc-gen-go-grpc 更新完成
-----
-----
更新 iotaer 插件中...

iotaer 更新完成
-----
```

### 新建项目

在当前目录新建一个名为 `MyProject` 的项目

```shell
[iotaer@iotaer iotaer]$ iotaer create --name MyProject --path .
```
