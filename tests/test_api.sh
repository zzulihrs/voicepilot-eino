#!/bin/bash

# VoicePilot-Eino API 测试脚本
# 测试所有 API 端点的功能

echo "🚀 VoicePilot-Eino API 测试"
echo "=========================="
echo ""

BASE_URL="http://localhost:8080"

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试计数器
TOTAL=0
PASSED=0
FAILED=0

# 测试函数
test_api() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4

    TOTAL=$((TOTAL + 1))
    echo -n "测试 $TOTAL: $name ... "

    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi

    status_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')

    if [ "$status_code" = "200" ]; then
        echo -e "${GREEN}✓ 通过${NC} (状态码: $status_code)"
        PASSED=$((PASSED + 1))
        if [ ! -z "$body" ]; then
            echo "   响应: $(echo $body | python3 -m json.tool 2>/dev/null | head -n 5)"
        fi
    else
        echo -e "${RED}✗ 失败${NC} (状态码: $status_code)"
        FAILED=$((FAILED + 1))
        echo "   错误: $body"
    fi
    echo ""
}

# 1. 健康检查
test_api "健康检查" "GET" "/api/health" ""

# 2. 文本交互 - 简单问候
test_api "文本交互 - 简单问候" "POST" "/api/text" '{"text": "你好"}'

# 3. 文本交互 - 打开应用
test_api "文本交互 - 打开应用" "POST" "/api/text" '{"text": "打开音乐"}'

# 4. 文本交互 - 播放音乐
test_api "文本交互 - 播放音乐" "POST" "/api/text" '{"text": "播放歌曲稻香"}'

# 5. 文本交互 - 生成文本
test_api "文本交互 - 生成文本" "POST" "/api/text" '{"text": "帮我写一篇关于人工智能的简短介绍"}'

# 打印总结
echo "=========================="
echo "测试总结"
echo "=========================="
echo -e "总计: $TOTAL"
echo -e "${GREEN}通过: $PASSED${NC}"
echo -e "${RED}失败: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ 所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}✗ 有 $FAILED 个测试失败${NC}"
    exit 1
fi
