#!/bin/bash
# 连接器授权认证测试脚本
# 用途：演示和测试 CLI 的连接器授权检查功能

set -e

echo "========================================="
echo "绫通 CLI - 连接器授权认证测试"
echo "========================================="
echo ""

# 检查环境变量
if [ -z "$LINGTONG_TOKEN" ]; then
    echo "⚠ 警告: LINGTONG_TOKEN 环境变量未设置"
    echo "  请设置: export LINGTONG_TOKEN=your_token"
    echo ""
    echo "继续测试将使用模拟数据演示..."
    echo ""
fi

if [ -z "$LINGTONG_HOST" ]; then
    export LINGTONG_HOST="https://app1.ltpass.com"
    echo "使用默认主机: $LINGTONG_HOST"
    echo ""
fi

# 测试 1: 列出连接器账户
echo "【测试 1】列出快麦ERP连接器账户"
echo "命令: lingtong-cli connector account list --connector kmerp"
echo "-----------------------------------------"
go run . connector account list --connector kmerp --format pretty 2>&1 || echo "(需要有效token)"
echo ""

# 测试 2: 检查特定连接器的授权状态
echo "【测试 2】检查快麦ERP授权状态"
echo "命令: lingtong-cli connector check-auth --connector kmerp"
echo "-----------------------------------------"
go run . connector check-auth --connector kmerp --format pretty 2>&1 || echo "(需要有效token)"
echo ""

# 测试 3: 检查工作流的连接器授权
echo "【测试 3】检查工作流947的连接器授权"
echo "命令: lingtong-cli connector check-auth --workflow-id 947"
echo "-----------------------------------------"
go run . connector check-auth --workflow-id 947 --format pretty 2>&1 || echo "(需要有效token)"
echo ""

# 测试 4: 创建连接器账户（演示）
echo "【测试 4】创建连接器账户（演示命令）"
echo "命令: lingtong-cli connector account create --connector kmerp --name '测试账号' --env test --data '{\"appKey\":\"xxx\",\"appSecret\":\"yyy\"}'"
echo "-----------------------------------------"
echo "注意: 此命令需要有效的token和真实的认证信息"
echo "实际使用时请替换 appKey 和 appSecret"
echo ""

# 测试 5: 验证连接器账户
echo "【测试 5】验证连接器账户（演示命令）"
echo "命令: lingtong-cli connector account verify --connector kmerp --account-id 123"
echo "-----------------------------------------"
echo "注意: 此命令需要有效的 account-id"
echo ""

echo "========================================="
echo "测试完成"
echo "========================================="
echo ""
echo "功能总结:"
echo "✓ connector account list - 列出连接器账户"
echo "✓ connector account create - 创建连接器账户"
echo "✓ connector account verify - 验证连接器账户"
echo "✓ connector check-auth - 检查工作流/场景的连接器授权状态"
echo ""
echo "使用提示:"
echo "1. 在执行工作流前，先检查连接器授权状态"
echo "2. 如果缺少授权，使用 account create 创建账户"
echo "3. 创建后使用 account verify 验证连接"
echo "4. 使用 check-auth 确认所有连接器已授权"
