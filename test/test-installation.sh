#!/bin/bash
# lingtong-cli 安装验证脚本
# 测试方式二：从源码安装的完整流程

set -e

echo "=========================================="
echo "lingtong-cli 安装验证测试"
echo "=========================================="
echo ""

# 测试 1: 检查 Go 环境
echo "[测试 1] 检查 Go 环境..."
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo "✓ Go 已安装: $GO_VERSION"
else
    echo "✗ Go 未安装"
    exit 1
fi
echo ""

# 测试 2: 检查源码目录
echo "[测试 2] 检查源码目录..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/Makefile" ]; then
    echo "✓ Makefile 存在"
else
    echo "✗ Makefile 不存在"
    exit 1
fi

if [ -f "$SCRIPT_DIR/main.go" ]; then
    echo "✓ main.go 存在"
else
    echo "✗ main.go 不存在"
    exit 1
fi
echo ""

# 测试 3: 构建 CLI
echo "[测试 3] 构建 CLI..."
cd "$SCRIPT_DIR"
make build
if [ -f "$SCRIPT_DIR/lingtong-cli" ]; then
    echo "✓ 构建成功"
else
    echo "✗ 构建失败"
    exit 1
fi
echo ""

# 测试 4: 安装到系统
echo "[测试 4] 安装到系统..."
make install
if command -v lingtong-cli &> /dev/null; then
    INSTALL_PATH=$(which lingtong-cli)
    echo "✓ 安装成功: $INSTALL_PATH"
else
    echo "✗ 安装失败"
    exit 1
fi
echo ""

# 测试 5: 验证版本
echo "[测试 5] 验证版本..."
VERSION=$(lingtong-cli --version)
echo "✓ 版本: $VERSION"
echo ""

# 测试 6: 验证帮助命令
echo "[测试 6] 验证帮助命令..."
if lingtong-cli --help &> /dev/null; then
    echo "✓ 帮助命令正常"
else
    echo "✗ 帮助命令失败"
    exit 1
fi
echo ""

# 测试 7: 验证配置命令
echo "[测试 7] 验证配置命令..."
lingtong-cli config init --host https://app1.ltpass.com
if lingtong-cli config show &> /dev/null; then
    echo "✓ 配置命令正常"
    lingtong-cli config show
else
    echo "✗ 配置命令失败"
    exit 1
fi
echo ""

# 测试 8: 验证命令列表
echo "[测试 8] 验证可用命令..."
COMMANDS=("config" "auth" "connector" "scene" "workflow" "table" "model" "api")
ALL_PASSED=true
for cmd in "${COMMANDS[@]}"; do
    if lingtong-cli $cmd --help &> /dev/null; then
        echo "  ✓ $cmd 命令可用"
    else
        echo "  ✗ $cmd 命令不可用"
        ALL_PASSED=false
    fi
done

if [ "$ALL_PASSED" = true ]; then
    echo "✓ 所有命令可用"
else
    echo "✗ 部分命令不可用"
    exit 1
fi
echo ""

# 测试 9: 验证快捷命令
echo "[测试 9] 验证快捷命令..."
SHORTCUTS=("+connector-info" "+scene-list" "+workflow-execute")
ALL_PASSED=true
for shortcut in "${SHORTCUTS[@]}"; do
    if lingtong-cli $shortcut --help &> /dev/null; then
        echo "  ✓ $shortcut 可用"
    else
        echo "  ✗ $shortcut 不可用"
        ALL_PASSED=false
    fi
done

if [ "$ALL_PASSED" = true ]; then
    echo "✓ 所有快捷命令可用"
else
    echo "✗ 部分快捷命令不可用"
    exit 1
fi
echo ""

# 测试 10: 验证 Skills 目录
echo "[测试 10] 验证 Skills 目录..."
if [ -d "$SCRIPT_DIR/skills" ]; then
    SKILL_COUNT=$(ls -1 "$SCRIPT_DIR/skills" | wc -l | tr -d ' ')
    echo "✓ Skills 目录存在，包含 $SKILL_COUNT 个技能"
    ls -1 "$SCRIPT_DIR/skills" | while read skill; do
        echo "  - $skill"
    done
else
    echo "✗ Skills 目录不存在"
    exit 1
fi
echo ""

echo "=========================================="
echo "所有测试通过！✓"
echo "=========================================="
echo ""
echo "安装总结："
echo "  - Go 环境: 正常"
echo "  - 源码构建: 成功"
echo "  - 系统安装: 成功"
echo "  - 命令验证: 全部通过"
echo "  - Skills: 已就绪"
echo ""
echo "下一步："
echo "  1. 运行 'lingtong-cli auth login' 进行认证"
echo "  2. 运行 'lingtong-cli connector info --connector kmerp' 测试 API 调用"
echo ""
