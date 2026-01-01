# Yatori 课程管理系统

一个现代化的课程管理系统，提供账号管理、课程管理、功能配置等功能。

## 项目架构

### 技术栈

- **前端框架**: Next.js 14
- **UI组件库**: ShadCN UI
- **样式**: Tailwind CSS
- **状态管理**: React Hooks
- **动画**: Framer Motion
- **HTTP客户端**: Axios

### 目录结构

```
├── app/                    # Next.js 应用目录
│   ├── (dashboard)/        # 仪表盘布局
│   │   ├── accounts/       # 账号管理页面
│   │   ├── questions/      # 答题管理页面
│   │   └── layout.tsx      # 仪表盘布局
│   └── layout.tsx          # 根布局
├── components/             # 组件目录
│   ├── account-card.tsx    # 账号卡片组件
│   ├── account-detail.tsx  # 账号详情组件
│   ├── account-list.tsx    # 账号列表组件
│   ├── add-account-dialog.tsx  # 添加账号弹窗
│   ├── ai-config-form.tsx  # AI配置表单
│   ├── sidebar.tsx         # 侧边栏组件
│   └── ...                 # 其他组件
├── api/                    # API请求目录
│   ├── base.ts             # API基础配置
│   ├── accountApi.ts       # 账号API
│   └── index.ts            # API统一导出
├── utils/                  # 工具函数目录
│   └── platformUtils.ts    # 平台名称映射工具
├── public/                 # 静态资源目录
└── README.md               # 项目文档
```

### 主要功能

#### 1. 账号管理
- 账号列表展示
- 账号详情查看
- 添加新账号
- 删除账号
- 账号信息编辑

#### 2. 功能配置
- 刷视频模式选择（0: 不刷, 1: 普通模式, 2: 暴力模式）
- 自动考试模式选择（0: 不考, 1: AI考试, 2: 外置题库考试）
- 自动提交试卷模式选择（0: 只保存, 1: 自动提交）
- 通知邮箱列表配置（支持动态添加/删除）

#### 3. 答题管理
- AI模型配置
- 外部题库配置

### 代码规范

- 使用TypeScript进行类型检查
- 使用React Hooks进行状态管理
- 组件化开发，提高代码复用性
- API请求集中管理，便于维护

### 快速开始

#### 安装依赖

```bash
npm install
```

#### 运行开发服务器

```bash
npm run dev
```

访问 http://localhost:3000 查看应用。

#### 构建生产版本

```bash
npm run build
npm start
```

### 注意事项

- 确保已配置正确的API地址
- 首次使用需要配置AI模型和API密钥
- 外部题库功能需要正确配置API地址

## 贡献

欢迎提交Issue和Pull Request。

## 许可证

MIT
