# Yatori 课程管理系统（Frontend）

基于 Next.js 16 的 console Web 管理界面。账号、课程配置、AI 配置都通过 console 的 `/api/v1/*` 接口读写，账号数据保存在本地 SQLite（`yatori.db`），不再依赖 `config.yaml` 维护账号。

## 技术栈

- Next.js 16（静态导出 `output: 'export'`）
- React 19 + TypeScript
- ShadCN UI + Tailwind CSS
- Axios

## 目录结构

```text
frontend/
├── app/                    # Next.js App Router
│   ├── (dashboard)/
│   │   ├── accounts/       # 账号管理
│   │   ├── questions/      # AI 配置
│   │   └── layout.tsx
│   └── layout.tsx
├── api/
│   ├── base.ts             # API 基础配置
│   ├── accountApi.ts       # 账号 / 任务 / 登录检测
│   ├── courseApi.ts        # 课程列表
│   └── settingApi.ts       # 全局设置
├── components/             # 账号卡片、详情、添加弹窗、AI 配置等
├── utils/platformUtils.ts  # 平台、视频模式、答题模式映射
└── public/
```

## 开发

```bash
npm install
# console 默认在 8080 提供 API；Next dev 默认在 3000
NEXT_PUBLIC_API_BASE=http://localhost:8080/api npm run dev
```

访问 `http://localhost:3000/web/accounts`。

Windows PowerShell：

```powershell
$env:NEXT_PUBLIC_API_BASE="http://localhost:8080/api"
npm run dev
```

## 构建

```bash
npm run build
```

静态产物在 `frontend/out`。console 会优先从该目录提供页面；发布时也可以把 `out` 复制到 console 的 `assets/web`。

启动 console Web 模式：

```bash
./yatori-go-console --web
```

访问 `http://localhost:8080/web/`。

## 功能

### 账号管理

- 添加 / 删除账号，账号配置写入 SQLite
- 平台：学习通、英华、海旗、智慧职教、CQIE、学习公社、WeLearn、青书、码上研训
- 视频模式、答题模式、交卷方式
- 备注名、IP 代理、通知邮箱
- 只刷/排除课程
- 登录检测
- 启动 / 停止任务，可按课程选择本次执行范围

### AI 配置

- AI 提供商、模型、API Key
- 外部题库 URL
- 设置保存在 SQLite，保存后对之后启动的任务生效

### 答题模式

- 0：不答题
- 1：AI 大模型答题（海旗优先使用平台标答，缺失才调用 AI）
- 2：外置题库答题；海旗为「内置标答答题」，不调用 AI
- 3：学习通内置 AI（当前按普通 AI 处理）

## 说明

- Next.js 的 `basePath` 为 `/web`，由 console 的 Gin 服务统一托管。
- 生产环境不需要单独启动 Next 服务。
- 账号配置不再写入 `config.yaml`；`config.yaml` 仅保留基础设置、AI、邮箱等全局项（后续会迁移到数据库）。
