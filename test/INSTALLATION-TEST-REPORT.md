# lingtong-cli 安装验证报告

## 测试时间
2026-04-06

## 测试环境
- **操作系统**: macOS (Darwin 25.3.0)
- **Go 版本**: go1.26.1
- **安装方式**: 方式二 - 从源码安装

## 测试结果

### ✓ 所有测试通过 (10/10)

| 测试项 | 状态 | 说明 |
|--------|------|------|
| Go 环境检查 | ✓ 通过 | go1.26.1 已安装 |
| 源码目录检查 | ✓ 通过 | Makefile、main.go 存在 |
| CLI 构建 | ✓ 通过 | `make build` 成功 |
| 系统安装 | ✓ 通过 | 安装到 /usr/local/bin/lingtong-cli |
| 版本验证 | ✓ 通过 | lingtong-cli version dev |
| 帮助命令 | ✓ 通过 | --help 正常显示 |
| 配置命令 | ✓ 通过 | config init/show 正常 |
| 业务命令 | ✓ 通过 | 8/8 命令可用 |
| 快捷命令 | ✓ 通过 | 3/3 快捷命令可用 |
| Skills 目录 | ✓ 通过 | 6 个技能已就绪 |

### 可用命令列表

#### 业务命令 (8个)
- `config` - 配置管理
- `auth` - 认证管理
- `connector` - 连接器管理
- `scene` - 场景管理
- `workflow` - 工作流管理
- `table` - 表格管理
- `model` - 模型元数据
- `api` - 通用 API 调用

#### 快捷命令 (3个)
- `+connector-info` - 快捷查询连接器
- `+scene-list` - 快捷列出场景
- `+workflow-execute` - 快捷执行工作流

### Agent Skills (6个)
- `lingtong-shared` - 认证、配置、安全规则
- `lingtong-connector` - 连接器管理
- `lingtong-scene` - 场景管理
- `lingtong-workflow` - 工作流编排
- `lingtong-table` - 表格操作
- `lingtong-model` - 模型元数据

## 安装步骤验证

### 方式二：从源码安装（已验证）

```bash
# 1. 克隆源码
git clone <repository-url>
cd lingtong-cli

# 2. 构建并安装
make install

# 输出:
# go build -trimpath -ldflags "..." -o lingtong-cli .
# install -d /usr/local/bin
# install -m755 lingtong-cli /usr/local/bin/lingtong-cli
# OK: /usr/local/bin/lingtong-cli (dev)

# 3. 验证安装
lingtong-cli --version
# 输出: lingtong-cli version dev

# 4. 配置主机地址
lingtong-cli config init --host https://app1.ltpass.com

# 5. 查看配置
lingtong-cli config show
# 输出:
# Host:  https://app1.ltpass.com
# Brand: lingtong
```

## 注意事项

1. **权限警告**: 安装时可能出现 `chmod 755 /usr/local/bin: Operation not permitted`，这是因为目录权限限制，但不影响二进制文件的安装。

2. **环境要求**: 
   - Go v1.23+ (测试使用 go1.26.1)
   - Git (用于克隆源码)

3. **安装路径**: 默认安装到 `/usr/local/bin/lingtong-cli`

## 结论

**方式二（从源码安装）完全可用**，所有功能正常，可以写入文档。

## 测试脚本

完整的自动化测试脚本位于：
- `lingtong-cli/test-installation.sh`

运行方式：
```bash
cd lingtong-cli
chmod +x test-installation.sh
./test-installation.sh
```
