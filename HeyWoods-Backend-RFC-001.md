# 项目结构议案

> Actor-Backend: RFC-001《项目结构议案》


```markdown
├── Project
│   ├── config [配置文件解析]
│   ├── cmd [命令行]
│   │   ├── api.go [文件名对应职能 在这里完成初始化]
│   │   ├── grpc.go
│   │   └── exec.go
│   ├── internal [内部实现 替换前app目录]
│   │   ├── controller [api实现]
│   │   ├── services [grpc实现]
│   │   ├── logic [api和grpc可复用的逻辑]
│   │   ├── crontab [cron定时任务]
│   │   ├── consumer [消费者]
│   │   ├── producer [生产者]
│   │   └── router [路由 将初始化过程和路由解耦]
│   ├── infra [基础设施 项目数据层交互用客户端]
│   │   ├── dao [driver客户端]
│   │   ├── middleware [路由中间件]
│   │   ├── mq [队列]
│   │   └── alarm [报警]
│   ├── common [全项目可复用部份.不可依赖本项目其他目录]
│   ├── model [模型层 只定义结构体 包括api/grpc的req/resp]
│   ├── adapters [适配层 实现model的CRUD]
│   ├── docs [文档说明]
│   └── project.go [项目入口]
```