#!/usr/bin/env bash
# 计算 go build 的 -tags 参数（供 .mise/tasks/ 下的构建任务共用）：
#   1. mise.toml 的 BOT_BLACKLIST（逗号/空格分隔）→ -tags=no<名>...
#   2. .env 存在 → 先由 tools/gensecrets 生成内嵌密钥文件，追加 -tags=secrets
# 输出形如 "-tags=nogal,secrets"；无标签时输出空串。
set -euo pipefail
cd "$(dirname "$0")/.."

tags=""
for b in $(printf '%s' "${BOT_BLACKLIST:-}" | tr ',' ' '); do
  tags="${tags:+$tags,}no$b"
done

if [ -f .env ]; then
  go run ./tools/gensecrets
  tags="${tags:+$tags,}secrets"
fi

if [ -n "$tags" ]; then
  echo "-tags=$tags"
fi
