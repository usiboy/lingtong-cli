# 测试结构说明

## 目录结构

```
lingtong-cli/
├── cmd/app/
│   ├── app.go              # 业务代码
│   ├── export.go           # 业务代码
│   ├── validate.go         # 业务代码
│   ├── import.go           # 业务代码
│   ├── scaffold.go         # 业务代码
│   ├── diff.go             # 业务代码
│   ├── scene.go            # 业务代码
│   ├── export_test.go      # 单元测试 (Go 标准：与业务代码同目录)
│   ├── validate_test.go    # 单元测试
│   ├── import_test.go      # 单元测试
│   ├── scaffold_test.go    # 单元测试
│   ├── diff_test.go        # 单元测试
│   └── scene_test.go       # 单元测试
├── test/
│   ├── test-full-validation.sh   # 集成测试脚本
│   ├── test-installation.sh      # 安装测试脚本
│   └── TEST-STRUCTURE.md         # 本文件
└── doc/
    └── lingtong-cli-guide.md     # 使用指南
```

## 测试分类

| 类型 | 位置 | 说明 |
|------|------|------|
| 单元测试 | `cmd/app/*_test.go` | Go 标准，与业务代码同目录，使用 `go test` 运行 |
| 集成测试 | `test/test-full-validation.sh` | 端到端测试，验证完整 CLI 功能 |
| 安装测试 | `test/test-installation.sh` | 验证 CLI 安装和命令可用性 |
| 兼容性测试 | `test/test-opencode-compatibility.sh` | 验证与 OpenCode 的兼容性 |

## 运行测试

```bash
# 运行所有单元测试
cd lingtong-cli && go test ./... -cover

# 运行特定包的测试
cd lingtong-cli && go test ./cmd/app/... -cover -v

# 运行集成测试
cd lingtong-cli && bash test/test-full-validation.sh
```

## 为什么 Go 单元测试与业务代码同目录？

这是 Go 语言的标准约定：
- `go test` 要求测试文件与被测试的代码在同一包中
- `*_test.go` 文件后缀明确区分了测试代码和业务代码
- IDE 和工具链都基于此约定工作

业务代码和测试代码通过后缀名清晰分离，不需要物理目录隔离。
