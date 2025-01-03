# GM-Note

一个基于 Wails + Vue3 开发的命令管理工具。

## 前置要求

1. 安装 Go 1.21 或更高版本
2. 安装 Node.js 和 npm
3. 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```
## 安装依赖

1. 安装 Go 依赖

```bash
go mod download
go mod tidy
```

2. 安装前端依赖

```bash
cd frontend
npm install
```
## 编译

```bash
wails build -trimpath -ldflags "-s -w" -upx
```

编译后的可执行文件将在 `build/bin` 目录下。

首次运行时会自动生成 `config.yaml` 配置文件。