#!/bin/bash
# lingtong-cli 完整验证测试脚本
# 用于每次代码调整后的全面验证
# 用法: ./test-full-validation.sh

set +e  # 不要遇到错误就退出，我们要运行所有测试

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试统计
PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0
TOTAL_COUNT=0

# 测试环境配置
TEST_HOST="https://app1.ltpass.com"
TEST_TOKEN="${LINGTONG_API_TOKEN:-apk-Gx6vDOEmALY7iJRcLZcD4nWF}"
TEST_CONNECTOR="kmerp"
TEST_ENV="test"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASS_COUNT++))
    ((TOTAL_COUNT++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAIL_COUNT++))
    ((TOTAL_COUNT++))
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    ((WARN_COUNT++))
    ((TOTAL_COUNT++))
}

section() {
    echo ""
    echo -e "${BLUE}=========================================="
    echo "$1"
    echo -e "==========================================${NC}"
    echo ""
}

# ==========================================
# 测试开始
# ==========================================

section "lingtong-cli 完整验证测试"
echo "测试时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "测试环境: $TEST_HOST"
echo ""

# ==========================================
# 阶段 1: 环境检查
# ==========================================
section "阶段 1: 环境检查"

# 1.1 检查 Go 环境
log_info "检查 Go 环境..."
if command -v go &> /dev/null; then
    GO_VERSION=$(go version 2>&1)
    log_pass "Go 环境: $GO_VERSION"
else
    log_fail "Go 未安装"
fi

# 1.2 检查源码目录
log_info "检查源码目录..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/go.mod" ]; then
    log_pass "go.mod 存在"
else
    log_fail "go.mod 不存在"
fi

if [ -f "$SCRIPT_DIR/Makefile" ]; then
    log_pass "Makefile 存在"
else
    log_fail "Makefile 不存在"
fi

# 1.3 检查关键代码文件
log_info "检查关键代码文件..."
CRITICAL_FILES=(
    "cmd/auth/auth.go"
    "internal/client/client.go"
    "internal/cmdutil/factory.go"
    "internal/auth/keychain.go"
    "internal/config/config.go"
)

for file in "${CRITICAL_FILES[@]}"; do
    if [ -f "$SCRIPT_DIR/$file" ]; then
        log_pass "$file 存在"
    else
        log_fail "$file 不存在"
    fi
done

# 1.4 检查技能文件
log_info "检查技能文件..."
SKILLS=(
    "skills/lingtong-shared/SKILL.md"
    "skills/lingtong-connector/SKILL.md"
    "skills/lingtong-workflow/SKILL.md"
    "skills/lingtong-scene/SKILL.md"
    "skills/lingtong-table/SKILL.md"
    "skills/lingtong-model/SKILL.md"
)

for skill in "${SKILLS[@]}"; do
    if [ -f "$SCRIPT_DIR/$skill" ]; then
        log_pass "$skill 存在"
    else
        log_fail "$skill 不存在"
    fi
done

# ==========================================
# 阶段 2: 构建测试
# ==========================================
section "阶段 2: 构建测试"

# 2.1 清理旧构建
log_info "清理旧构建..."
make clean > /dev/null 2>&1
log_pass "清理完成"

# 2.2 代码检查
log_info "运行 go vet..."
cd "$SCRIPT_DIR"
VET_OUTPUT=$(go vet ./... 2>&1)
if [ $? -eq 0 ]; then
    log_pass "go vet 通过"
else
    log_fail "go vet 失败: $VET_OUTPUT"
fi

# 2.3 构建二进制文件
log_info "构建 CLI..."
BUILD_OUTPUT=$(make build 2>&1)
if [ $? -eq 0 ]; then
    log_pass "构建成功"
else
    log_fail "构建失败: $BUILD_OUTPUT"
fi

# 2.4 检查二进制文件
if [ -f "$SCRIPT_DIR/lingtong-cli" ]; then
    BINARY_SIZE=$(du -h "$SCRIPT_DIR/lingtong-cli" | cut -f1)
    log_pass "二进制文件生成 ($BINARY_SIZE)"
else
    log_fail "二进制文件未生成"
fi

# 2.5 安装到系统
log_info "安装到系统..."
INSTALL_OUTPUT=$(make install 2>&1)
if echo "$INSTALL_OUTPUT" | grep -q "OK:"; then
    log_pass "安装成功"
else
    log_warn "安装可能有警告: $INSTALL_OUTPUT"
fi

# 2.6 验证安装
if command -v lingtong-cli &> /dev/null; then
    CLI_PATH=$(which lingtong-cli)
    CLI_VERSION=$(lingtong-cli --version 2>&1)
    log_pass "CLI 可用: $CLI_PATH ($CLI_VERSION)"
else
    log_fail "CLI 未在 PATH 中"
fi

# ==========================================
# 阶段 3: 配置测试
# ==========================================
section "阶段 3: 配置测试"

# 3.1 清理旧配置
log_info "清理测试配置..."
rm -rf ~/.lingtong-cli/test-config 2>/dev/null
log_pass "清理完成"

# 3.2 初始化配置
log_info "初始化配置..."
CONFIG_OUTPUT=$(lingtong-cli config init --host "$TEST_HOST" 2>&1)
if [ $? -eq 0 ]; then
    log_pass "配置初始化成功"
else
    log_fail "配置初始化失败: $CONFIG_OUTPUT"
fi

# 3.3 验证配置
log_info "验证配置..."
CONFIG_SHOW=$(lingtong-cli config show 2>&1)
if echo "$CONFIG_SHOW" | grep -q "$TEST_HOST"; then
    log_pass "配置正确: $TEST_HOST"
else
    log_fail "配置错误: $CONFIG_SHOW"
fi

# ==========================================
# 阶段 4: 认证测试
# ==========================================
section "阶段 4: 认证测试"

# 4.1 清理旧认证
log_info "清理旧认证..."
lingtong-cli auth logout > /dev/null 2>&1
log_pass "清理完成"

# 4.2 测试 auth status (未认证)
log_info "测试未认证状态..."
STATUS_OUTPUT=$(lingtong-cli auth status 2>&1)
if echo "$STATUS_OUTPUT" | grep -q "Not authenticated"; then
    log_pass "未认证状态显示正确"
else
    log_warn "未认证状态显示: $STATUS_OUTPUT"
fi

# 4.3 测试登录 - 方式 1: 直接 Token
log_info "测试登录 - 直接 Token..."
LOGIN_OUTPUT=$(lingtong-cli auth login --token "$TEST_TOKEN" 2>&1)
if echo "$LOGIN_OUTPUT" | grep -q "Login successful"; then
    log_pass "登录成功 (直接 Token)"
else
    log_fail "登录失败: $LOGIN_OUTPUT"
fi

# 4.4 测试 auth status (已认证)
log_info "测试已认证状态..."
STATUS_OUTPUT=$(lingtong-cli auth status 2>&1)
if echo "$STATUS_OUTPUT" | grep -q "Authenticated: Yes"; then
    log_pass "已认证状态显示正确"
    if echo "$STATUS_OUTPUT" | grep -q "Valid"; then
        log_pass "Token 验证通过"
    else
        log_warn "Token 验证状态未知"
    fi
else
    log_fail "已认证状态显示错误: $STATUS_OUTPUT"
fi

# 4.5 测试登出
log_info "测试登出..."
LOGOUT_OUTPUT=$(lingtong-cli auth logout 2>&1)
if echo "$LOGOUT_OUTPUT" | grep -q "Logged out successfully"; then
    log_pass "登出成功"
else
    log_fail "登出失败: $LOGOUT_OUTPUT"
fi

# 4.6 测试登录 - 方式 2: 环境变量
log_info "测试登录 - 环境变量..."
LOGIN_ENV_OUTPUT=$(LINGTONG_API_TOKEN="$TEST_TOKEN" lingtong-cli auth login --from-env 2>&1)
if echo "$LOGIN_ENV_OUTPUT" | grep -q "Login successful"; then
    log_pass "登录成功 (环境变量)"
else
    log_fail "登录失败 (环境变量): $LOGIN_ENV_OUTPUT"
fi

# ==========================================
# 阶段 5: API 调用测试
# ==========================================
section "阶段 5: API 调用测试"

# 5.1 测试连接器查询
log_info "测试连接器查询..."
CONNECTOR_OUTPUT=$(lingtong-cli connector info --connector "$TEST_CONNECTOR" --env "$TEST_ENV" 2>&1)
if echo "$CONNECTOR_OUTPUT" | grep -q '"success": true\|"code": 10000'; then
    log_pass "连接器查询成功"
else
    log_fail "连接器查询失败: $CONNECTOR_OUTPUT"
fi

# 5.2 测试场景列表
log_info "测试场景列表..."
SCENE_OUTPUT=$(lingtong-cli scene list 2>&1)
if [ $? -eq 0 ]; then
    log_pass "场景列表查询成功"
else
    log_warn "场景列表查询: $SCENE_OUTPUT"
fi

# 5.3 测试工作流列表
log_info "测试工作流列表..."
WORKFLOW_OUTPUT=$(lingtong-cli workflow list --help 2>&1)
if [ $? -eq 0 ]; then
    log_pass "工作流命令可用"
else
    log_fail "工作流命令不可用"
fi

# 5.4 测试表格命令
log_info "测试表格命令..."
TABLE_OUTPUT=$(lingtong-cli table list --help 2>&1)
if [ $? -eq 0 ]; then
    log_pass "表格命令可用"
else
    log_fail "表格命令不可用"
fi

# 5.5 测试模型命令
log_info "测试模型命令..."
MODEL_OUTPUT=$(lingtong-cli model interface list --help 2>&1)
if [ $? -eq 0 ]; then
    log_pass "模型命令可用"
else
    log_fail "模型命令不可用"
fi

# ==========================================
# 阶段 6: 命令帮助测试
# ==========================================
section "阶段 6: 命令帮助测试"

COMMANDS=(
    "config"
    "auth"
    "connector"
    "scene"
    "workflow"
    "table"
    "model"
    "api"
)

for cmd in "${COMMANDS[@]}"; do
    log_info "测试 $cmd --help..."
    HELP_OUTPUT=$(lingtong-cli $cmd --help 2>&1)
    if [ $? -eq 0 ]; then
        log_pass "$cmd 帮助正常"
    else
        log_fail "$cmd 帮助异常"
    fi
done

# ==========================================
# 阶段 7: 快捷命令测试
# ==========================================
section "阶段 7: 快捷命令测试"

SHORTCUTS=(
    "+connector-info"
    "+scene-list"
    "+workflow-execute"
)

for shortcut in "${SHORTCUTS[@]}"; do
    log_info "测试 $shortcut --help..."
    SHORTCUT_OUTPUT=$(lingtong-cli $shortcut --help 2>&1)
    if [ $? -eq 0 ]; then
        log_pass "$shortcut 可用"
    else
        log_fail "$shortcut 不可用"
    fi
done

# ==========================================
# 阶段 8: 错误处理测试
# ==========================================
section "阶段 8: 错误处理测试"

# 8.1 测试无效命令
log_info "测试无效命令..."
INVALID_OUTPUT=$(lingtong-cli invalid-command 2>&1)
if [ $? -ne 0 ]; then
    log_pass "无效命令正确返回错误"
else
    log_warn "无效命令未返回错误"
fi

# 8.2 测试缺少必需参数
log_info "测试缺少必需参数..."
MISSING_PARAM_OUTPUT=$(lingtong-cli connector info 2>&1)
if echo "$MISSING_PARAM_OUTPUT" | grep -qi "required\|error"; then
    log_pass "缺少参数检测正常"
else
    log_warn "缺少参数检测: $MISSING_PARAM_OUTPUT"
fi

# 8.3 测试无效 Token
log_info "测试无效 Token..."
INVALID_TOKEN_OUTPUT=$(lingtong-cli auth login --token "invalid-token" 2>&1)
if echo "$INVALID_TOKEN_OUTPUT" | grep -qi "Warning\|invalid\|failed"; then
    log_pass "无效 Token 检测正常"
else
    log_warn "无效 Token 检测: $INVALID_TOKEN_OUTPUT"
fi

# ==========================================
# 阶段 9: 输出格式测试
# ==========================================
section "阶段 9: 输出格式测试"

# 重新登录以确保认证状态
lingtong-cli auth login --token "$TEST_TOKEN" > /dev/null 2>&1

# 9.1 JSON 格式
log_info "测试 JSON 输出格式..."
JSON_OUTPUT=$(lingtong-cli connector info --connector "$TEST_CONNECTOR" --env "$TEST_ENV" --format json 2>&1)
if echo "$JSON_OUTPUT" | grep -q "^{"; then
    log_pass "JSON 格式正确"
else
    log_fail "JSON 格式异常"
fi

# 9.2 Pretty 格式
log_info "测试 Pretty 输出格式..."
PRETTY_OUTPUT=$(lingtong-cli connector info --connector "$TEST_CONNECTOR" --env "$TEST_ENV" --format pretty 2>&1)
if [ $? -eq 0 ]; then
    log_pass "Pretty 格式正常"
else
    log_fail "Pretty 格式异常"
fi

# 9.3 Table 格式
log_info "测试 Table 输出格式..."
TABLE_FORMAT_OUTPUT=$(lingtong-cli connector info --connector "$TEST_CONNECTOR" --env "$TEST_ENV" --format table 2>&1)
if [ $? -eq 0 ]; then
    log_pass "Table 格式正常"
else
    log_fail "Table 格式异常"
fi

# ==========================================
# 阶段 10: 技能集成测试
# ==========================================
section "阶段 10: 技能集成测试"

OPENCODE_SKILLS_DIR="/Users/liumingjian/Development/source/lingtong_source/.opencode/skills"
CLI_SKILLS=(
    "lingtong-cli-shared"
    "lingtong-cli-connector"
    "lingtong-cli-scene"
    "lingtong-cli-workflow"
    "lingtong-cli-table"
    "lingtong-cli-model"
)

for skill in "${CLI_SKILLS[@]}"; do
    if [ -L "$OPENCODE_SKILLS_DIR/$skill" ] || [ -d "$OPENCODE_SKILLS_DIR/$skill" ]; then
        if [ -f "$OPENCODE_SKILLS_DIR/$skill/SKILL.md" ]; then
            log_pass "技能 $skill 已集成"
        else
            log_warn "技能 $skill 缺少 SKILL.md"
        fi
    else
        log_fail "技能 $skill 未找到"
    fi
done

# ==========================================
# 测试总结
# ==========================================
section "测试总结"

echo "总测试数: $TOTAL_COUNT"
echo -e "${GREEN}✓ 通过:   $PASS_COUNT${NC}"
echo -e "${RED}✗ 失败:   $FAIL_COUNT${NC}"
echo -e "${YELLOW}⚠ 警告:   $WARN_COUNT${NC}"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    COMPATIBILITY_RATE=$(( (PASS_COUNT * 100) / TOTAL_COUNT ))
    echo -e "${GREEN}=========================================="
    echo "兼容性评分: ${COMPATIBILITY_RATE}%"
    echo "==========================================${NC}"
    echo ""
    
    if [ $COMPATIBILITY_RATE -ge 95 ]; then
        echo -e "${GREEN}✓ 测试通过！代码可以提交。${NC}"
        echo ""
        echo "建议的提交消息:"
        echo "  feat: implement full authentication module"
        echo "  - Add API token login with 3 methods (direct, env, interactive)"
        echo "  - Implement token storage in OS Keychain"
        echo "  - Add token verification and status check"
        echo "  - Auto-load token from Keychain in Factory"
        echo "  - Update documentation and compatibility report"
        echo ""
        exit 0
    elif [ $COMPATIBILITY_RATE -ge 80 ]; then
        echo -e "${YELLOW}⚠ 基本通过，但有警告需要处理。${NC}"
        echo ""
        exit 0
    else
        echo -e "${RED}✗ 通过率过低，需要修复后再提交。${NC}"
        echo ""
        exit 1
    fi
else
    echo -e "${RED}=========================================="
    echo "✗ 测试失败: $FAIL_COUNT 个错误"
    echo "==========================================${NC}"
    echo ""
    echo "请修复以下问题后重新测试:"
    echo "  - 检查失败的项目并修复"
    echo "  - 运行 ./test-full-validation.sh 重新验证"
    echo ""
    exit 1
fi
