# lingtong-cli 认证模块使用指南

## 概述

lingtong-cli 使用 API Key（格式：`apk-xxx`）进行认证，Token 安全存储在操作系统 Keychain 中。

## 认证方式

### 方式 1: 直接提供 Token（推荐）

```bash
lingtong-cli auth login --token apk-xxx
```

### 方式 2: 从环境变量读取

```bash
# 设置环境变量
export LINGTONG_API_TOKEN=apk-xxx

# 从环境变量登录
lingtong-cli auth login --from-env
```

### 方式 3: 交互式输入

```bash
lingtong-cli auth login
# 提示输入: API Token (apk-xxx):
```

## 认证管理

### 查看认证状态

```bash
lingtong-cli auth status
```

**输出示例**:
```
Authenticated: Yes
Token: apk-Gx6v...4nWF
Host: https://app1.ltpass.com

Verifying token... Valid
```

### 登出

```bash
lingtong-cli auth logout
```

**输出示例**:
```
Logged out successfully. Token removed from keychain.
```

## AI Agent 使用示例

### 自动化认证流程

```bash
# 1. 配置主机地址
lingtong-cli config init --host https://app1.ltpass.com

# 2. 使用环境变量认证（适合 AI Agent）
export LINGTONG_API_TOKEN=apk-xxx
lingtong-cli auth login --from-env

# 3. 验证认证状态
lingtong-cli auth status

# 4. 开始使用 API
lingtong-cli connector info --connector kmerp --env test
```

### 在 OpenCode 环境中使用

在 `.opencode/skills/lingtong-api/.env` 中配置 Token：

```bash
LINGTONG_API_HOST=https://app1.ltpass.com
LINGTONG_API_TOKEN=apk-xxx
```

然后 AI Agent 可以自动读取并认证：

```bash
# AI Agent 自动从 .env 读取配置
lingtong-cli config init --host $LINGTONG_API_HOST
lingtong-cli auth login --token $LINGTONG_API_TOKEN
```

## Token 安全

- Token 存储在 OS Keychain（macOS Keychain / Windows Credential Manager / Linux Secret Service）
- 不会明文存储在配置文件或终端输出中
- `auth status` 命令会脱敏显示 Token（如 `apk-Gx6v...4nWF`）
- 每次请求自动携带 `Authorization: Bearer {token}` Header

## 常见问题

### Token 格式错误

```
Warning: Token should start with 'apk-'. Proceeding anyway...
```

**解决方案**: 确保 Token 格式正确，应以 `apk-` 开头。

### Token 验证失败

```
Warning: Token verification failed: invalid token: authentication failed
Token saved but may be invalid. You can try again with a valid token.
```

**解决方案**:
1. 检查 Token 是否正确
2. 确认主机地址配置正确
3. 联系管理员确认 Token 权限

### 未认证错误

```
Error: API error (401): {"code":401,"msg":"无身份信息!","success":false}
```

**解决方案**:
```bash
# 检查认证状态
lingtong-cli auth status

# 如果未认证，重新登录
lingtong-cli auth login --token apk-xxx
```

## 完整认证测试脚本

```bash
#!/bin/bash
# 测试完整认证流程

set -e

echo "=== 认证模块测试 ==="

# 1. 配置主机地址
echo "1. 配置主机地址..."
lingtong-cli config init --host https://app1.ltpass.com

# 2. 登录
echo "2. 登录..."
lingtong-cli auth login --token apk-xxx

# 3. 验证状态
echo "3. 验证认证状态..."
lingtong-cli auth status

# 4. 测试 API 调用
echo "4. 测试 API 调用..."
lingtong-cli connector info --connector kmerp --env test

# 5. 登出
echo "5. 登出..."
lingtong-cli auth logout

# 6. 验证登出
echo "6. 验证登出..."
lingtong-cli auth status

echo "=== 测试完成 ==="
```

## API 文档

### auth login

```bash
lingtong-cli auth login [flags]

Flags:
      --token string      API token (apk-xxx)
      --from-env          Read token from LINGTONG_API_TOKEN environment variable
  -h, --help              help for login
```

### auth status

```bash
lingtong-cli auth status

显示:
- 认证状态
- Token（脱敏）
- 主机地址
- Token 有效性验证结果
```

### auth logout

```bash
lingtong-cli auth logout

操作:
- 从 OS Keychain 删除 Token
- 清除认证状态
```
