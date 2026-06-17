#!/usr/bin/env bash
# refresh-examples.sh — re-emit every committed example against the
# current binary. Run before tagging every release so the examples/
# tree always reflects the latest version's output.
#
# The 4 deterministic (LLM-off) examples can be refreshed anywhere:
#   playwright-dev, gohugo-io, books-toscrape-com, es-wikipedia-org-madrid
#
# The 2 AI-on examples (spritecloud-com-dgx, petstore-swagger-io-dgx)
# require REVIEWQA_LLM pointed at an OpenAI-compatible endpoint. The
# spriteCloud DGX (http://100.82.34.115:11434) is reachable only via
# Netbird, so these two are skipped automatically when REVIEWQA_LLM is
# unset.
#
# Usage:
#   ./scripts/refresh-examples.sh                      # 4 deterministic only
#   REVIEWQA_LLM=http://… ./scripts/refresh-examples.sh  # all 6
set -euo pipefail

BIN=${BIN:-/tmp/reviewqa}
ROOT=$(git rev-parse --show-toplevel)
cd "$ROOT"

echo "==> Building binary"
go build -o "$BIN" ./cmd/reviewqa
"$BIN" --version || true

split_and_write() {
  local site_dir=$1 src=$2
  rm -rf "examples/$site_dir"
  mkdir -p "examples/$site_dir"
  awk -v base="examples/$site_dir" '
    /^--- .+ ---$/ {
      n = length($0)
      path = substr($0, 5, n - 8)
      if (path != "") {
        if (out) close(out)
        out = base "/" path
        sub_path = path
        sub(/[^/]*$/, "", sub_path)
        if (sub_path != "") system("mkdir -p \"" base "/" sub_path "\"")
        next
      }
    }
    { if (out) print >> out }
  ' "$src"
}

run_probe() {
  local site_dir=$1 url=$2; shift 2
  local out=/tmp/reviewqa-refresh-$site_dir.out
  echo "==> Probing $url -> examples/$site_dir/"
  env "$@" "$BIN" probe --url "$url" --dry-run > "$out"
  split_and_write "$site_dir" "$out"
}

# The 4 deterministic probes always run with LLM scrubbed from the env
# regardless of what the caller exported, so humanization doesn't fire.
run_probe playwright-dev          https://playwright.dev               -u REVIEWQA_LLM
run_probe gohugo-io               https://gohugo.io                    -u REVIEWQA_LLM
run_probe books-toscrape-com      https://books.toscrape.com           -u REVIEWQA_LLM
run_probe es-wikipedia-org-madrid https://es.wikipedia.org/wiki/Madrid -u REVIEWQA_LLM

# AI probes require an OpenAI-compatible endpoint reachable from this
# machine (spriteCloud's DGX is on Netbird; any local Ollama / vLLM
# works too). Skip cleanly when none is configured.
if [[ -n "${REVIEWQA_LLM:-}" ]]; then
  run_probe spritecloud-com-dgx     https://www.spritecloud.com    REVIEWQA_LLM="$REVIEWQA_LLM" REVIEWQA_HUMANIZE=0
  run_probe petstore-swagger-io-dgx https://petstore3.swagger.io   REVIEWQA_LLM="$REVIEWQA_LLM" REVIEWQA_HUMANIZE=0
else
  echo "==> REVIEWQA_LLM unset — skipping AI-on examples (spritecloud-com-dgx, petstore-swagger-io-dgx)."
fi

echo "==> Done. Review with: git status examples/"
