# Nlip - 轻量级网络剪贴板

[English](./README.md) | 简体中文

一个**由Cursor实现的**支持跨平台文本和文件共享的轻量级网络剪贴板系统。

<p align="center">
  <img src="docs/images/logo.png" alt="Nlip Logo" width="200"/>
  <br>
  <a href="https://github.com/yourusername/nlip/releases">
    <img src="https://img.shields.io/github/v/release/yourusername/nlip" alt="GitHub release">
  </a>
  <a href="https://github.com/yourusername/nlip/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/yourusername/nlip" alt="License">
  </a>
  <a href="https://github.com/yourusername/nlip/actions">
    <img src="https://github.com/yourusername/nlip/workflows/CI/badge.svg" alt="CI Status">
  </a>
</p>

## 目录

- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [快速开始](#快速开始)
- [项目结构](#项目结构)
- [开发规范](#开发规范)
- [部署指南](#部署指南)
- [参与贡献](#参与贡献)
- [安全性](#安全性)
- [许可证](#许可证)
- [致谢](#致谢)

## 功能特性

### 核心功能
- 多平台支持（Web、浏览器插件、移动应用）
- 文本和文件共享
- 基于空间的内容组织和管理
- 实时同步
- 空间权限管理

### 安全特性
- 用户认证和授权
- 访问频率限制
- 文件类型过滤
- 内容过期管理
- 空间级别的访问控制

### 其他特性
- 响应式设计
- 离线支持
- 多语言支持
- 深色模式

## 技术栈

### 后端
- Go 1.20+ 
- Gin Web 框架
- SQLite 数据库
- JWT 认证
- WebSocket 实时通信
- 访问频率限制
- 文件存储系统

### 前端
- Web：React 18 + TypeScript 5 + Ant Design 5
- 浏览器插件：Chrome Extension API + React
- 移动应用：Flutter 3
- 状态管理：Redux Toolkit + Redux Persist
- HTTP 客户端：Axios
- 构建工具：Vite

### API 文档

详细的 API 文档请查看 [API文档](docs/api/api_zh.md)

#### API 特性
- RESTful API 设计
- JWT 认证
- 实时WebSocket通知
- 请求频率限制
- 详细的错误处理
- 版本控制
- 支持调试模式

#### API 示例

```typescript
// 登录示例
const response = await fetch('/api/v1/nlip/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    username: 'test',
    password: '123456'
  })
});

const data = await response.json();
// 使用返回的token
const token = data.token;
```

更多示例请参考 [API文档](docs/api/api_zh.md)

## 快速开始

### 环境要求
- Go 1.20+
- Node.js 18+
- Flutter 3.0+
- SQLite 3

### 安装步骤

1. 克隆仓库
```bash
git clone https://github.com/yourusername/nlip.git
cd nlip
```

2. 安装后端依赖
```bash
cd src/backend
go mod download
```

3. 安装前端依赖
```bash
cd src/frontend/web
npm install
```

4. 配置环境变量
```bash
# 后端
cp src/backend/config.example.json src/backend/config.dev.json
# 编辑 config.dev.json 配置文件

# 前端
cp src/frontend/web/.env.example src/frontend/web/.env.local
# 编辑 .env.local 配置文件
```

5. 启动开发服务器
```bash
# 后端
cd src/backend
go run main.go

# 前端
cd src/frontend/web
npm run dev
```

### 开发环境访问
- Web 应用：http://localhost:5173
- API 服务：http://localhost:3000
- API 文档：http://localhost:3000/swagger/index.html

## 项目结构

```
nlip/
├── src/
│   ├── backend/           # Go 后端
│   │   ├── config/       # 配置
│   │   ├── handlers/     # 请求处理器
│   │   ├── middleware/   # 中间件
│   │   ├── models/       # 数据模型
│   │   ├── routes/       # 路由定义
│   │   ├── utils/        # 工具函数
│   │   └── main.go       # 入口文件
│   └── frontend/
│       ├── web/          # React Web 应用
│       │   ├── src/
│       │   │   ├── api/        # API 客户端
│       │   │   ├── components/ # React 组件
│       │   │   ├── hooks/      # 自定义 Hooks
│       │   │   ├── pages/      # 页面组件
│       │   │   ├── store/      # Redux 存储
│       │   │   └── utils/      # 工具函数
│       ├── extension/    # Chrome 插件
│       └── app/          # Flutter 移动应用
├── docs/                 # 文档
└── scripts/             # 构建脚本
```

## 开发规范

### 代码风格
- Go：遵循官方 Go 代码规范
- TypeScript：ESLint + Prettier
- SCSS：Stylelint
- 提交信息：Conventional Commits

### 分支策略
- main：生产就绪代码
- develop：开发分支
- feature/*：新功能
- bugfix/*：错误修复
- release/*：发布准备

## 部署指南

### Docker 部署
```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d
```

### 手动部署
详见 [部署文档](docs/deployment.md)

## 参与贡献

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: 添加某个特性'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

详细指南请参考 [贡献指南](CONTRIBUTING.md)

## 安全性

- 基于 JWT 的身份认证
- 访问频率限制
- 输入验证和净化
- 文件类型过滤
- 内容过期机制
- 生产环境强制 HTTPS

发现安全问题请参考 [安全政策](SECURITY.md)

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 致谢

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [React](https://reactjs.org/)
- [Ant Design](https://ant.design/)
- [Redux Toolkit](https://redux-toolkit.js.org/)

## 支持

如果这个项目对你有帮助，请考虑给它一个 star ⭐️