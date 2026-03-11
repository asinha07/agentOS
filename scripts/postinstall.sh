#!/usr/bin/env sh
set -eu

CFG="/etc/agentos/application.yml"
if [ ! -f "$CFG" ]; then
  mkdir -p /etc/agentos
  cat > "$CFG" << 'EOF'
# AgentOS default configuration (sample)
registry:
  url: ""
models:
  provider: "openai"
  model: "gpt-4.1"
# Providers keys: OPENAI_API_KEY, ANTHROPIC_API_KEY, XAI_API_KEY
EOF
  echo "[agentos] Wrote default config to $CFG"
fi

exit 0

