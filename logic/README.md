# logic 目录说明（yatori-go-core v3）

console 已迁移到 `github.com/yatori-dev/yatori-go-core/v3` 的统一 SDK。平台差异全部由 core 的 adapter 处理，console 不再按平台维护刷课代码。

## 目录结构

```text
logic/
├── Lunch.go        # 启动入口：配置检查、代理检查、Web/CLI 模式分发
├── runner/         # v3 统一执行器
│   ├── runner.go   # 登录、课程、章节、学习、作业、考试、日志、通知
│   └── runner_test.go
└── legacy/         # v2 旧平台实现，仅作迁移参考，不参与构建
    ├── cqie/
    ├── enaea/
    ├── haiqikeji/
    ├── icve/
    ├── ketangx/
    ├── qingshuxuetang/
    ├── welearn/
    ├── xuexitong/
    └── yinghua/
```

## 生效链路

```text
Lunch.brushBlock
  └─ runner.RunUser
       └─ yatori.New
            ├─ client.Login
            ├─ client.Courses
            ├─ client.Tree
            ├─ client.Study
            ├─ client.Works / DoWork
            └─ client.Exams / DoExam
```

## 平台能力在哪里扩展

新增平台或修改平台行为，应修改 core 的 `internal/driver/<platform>`，而不是在 console 里新增 `logic/<platform>`。

console 只负责策略映射：

- `videoModel`：0 不刷、1 普通、2 快速/秒刷、3 按普通处理
- `autoExam`：1 AI、2 外置题库（海旗为内置标答）、3 按 AI 处理
- `examAutoSubmit`：是否交卷
- 课程/考试 include/exclude
- 停止信号、日志、邮件、提示音

## legacy 说明

`legacy/` 下的文件都带有：

```go
//go:build ignore
```

默认不参与编译，仅用于对照 v2 行为。确认新执行器稳定后可以直接删除整个 `legacy/` 目录。
