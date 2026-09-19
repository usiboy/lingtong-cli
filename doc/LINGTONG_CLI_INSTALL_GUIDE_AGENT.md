# 绫通 CLI Agent 安装指南

本文供 Cursor、Claude Code、OpenCode、Qoder、Codex 等 AI Agent 执行安装。

## 安全约束

- 仅从 npm 包 `@lingtong-cli/cli` 和配置的 HTTPS Release 地址下载。
- 不要求用户在聊天中提供 Token。
- 不输出、记录或提交认证凭证。
- 安装后先执行只读自检，不自动创建、修改或删除业务数据。

## 安装

确认本机已安装 Node.js 18 或更高版本，然后执行：

```bash
node --version
npx @lingtong-cli/cli@latest install
lingtong-cli --version
```

默认安装目录：

| 平台 | 目录 |
|------|------|
| macOS/Linux | `~/.local/bin` |
| Windows | `%LOCALAPPDATA%\Lingtong\bin` |

如果目录不在 `PATH`，根据安装命令输出添加后重新打开终端。也可以在执行安装前设置：

```bash
export LINGTONG_CLI_INSTALL_DIR="$HOME/bin"
```

GitHub Release 不可达时，可指定公司 HTTPS 制品镜像。地址中的 `{version}` 会替换为 npm 包版本：

```bash
export LINGTONG_CLI_DOWNLOAD_BASE='https://downloads.example.com/lingtong-cli/v{version}'
npx @lingtong-cli/cli@latest install
```

## 初始化与自检

安装成功后执行：

```bash
lingtong-cli config init
lingtong-cli auth status
lingtong-cli doctor
lingtong-cli skills install --dry-run
```

`config init` 需要用户确认绫通服务 Host。未登录时，提示用户自行运行：

```bash
lingtong-cli auth login
```

登录完成后再执行：

```bash
lingtong-cli doctor
lingtong-cli skills install
lingtong-cli schema list --envelope
```

## 成功标准

- `lingtong-cli --version` 正常输出版本。
- `lingtong-cli doctor` 能识别配置、认证和网络状态。
- `lingtong-cli skills install --dry-run` 能列出目标编辑器位置。
- 未经用户确认，不调用任何写入或删除业务数据的命令。

## 故障处理

| 现象 | 处理 |
|------|------|
| `command not found` | 将安装目录加入 `PATH` 并重启终端 |
| GitHub 下载失败 | 设置 `LINGTONG_CLI_DOWNLOAD_BASE` 指向 HTTPS 镜像 |
| checksum 不匹配 | 停止安装，不绕过校验；检查镜像文件和 Release 版本 |
| 平台不支持 | 使用源码构建，或提交对应平台二进制发布需求 |
| 未配置 Host | 执行 `lingtong-cli config init` |
| 未认证或 Token 过期 | 提示用户执行 `lingtong-cli auth login` |
