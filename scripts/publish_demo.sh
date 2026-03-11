#!/usr/bin/env bash
set -euo pipefail

REG=${1:-http://localhost:8080}

echo "Starting registry if not running..."
echo "(Run: go run registry/server/main.go)"

AGENTS=(
  product-manager
  be-developer
  web-developer
  qa
  code-reviewer
  coding-agent
  research-agent
  seo-agent
  growth-hacker
  bug-fixer
  data-analyst
  web-scraper
)

for a in "${AGENTS[@]}"; do
  echo "Publishing $a"
  ./agent-go publish "$a" || agent publish "$a" || true
done

echo "Done. Try:"
echo "  agent search --registry $REG"
echo "  agent install coding-agent --registry $REG"
echo "  agent run coding-agent"

