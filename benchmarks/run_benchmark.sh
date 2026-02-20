#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# Avacado Benchmark Script
#
# Automatically discovers supported commands by scanning the
# internal/command/ source directory, then runs redis-benchmark
# against Avacado (and optionally Redis) for comparison.
#
# Usage:
#   ./benchmarks/run_benchmark.sh
#
# Environment variables:
#   AVACADO_PORT  - Port to run Avacado on (default: 6380)
#   REDIS_PORT    - Port Redis is running on (default: 6379)
#   REQUESTS      - Number of requests per command (default: 100000)
#   CLIENTS       - Number of concurrent clients (default: 50)
#   DATA_SIZE     - Value size in bytes (default: 3)
#   SKIP_REDIS    - Set to 1 to skip Redis comparison (default: 0)
# ============================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMMANDS_DIR="$REPO_ROOT/internal/command"
AVACADO_BIN="$REPO_ROOT/main_bench_tmp"

# Configuration (overridable via environment)
AVACADO_PORT="${AVACADO_PORT:-6380}"
REDIS_PORT="${REDIS_PORT:-6379}"
REQUESTS="${REQUESTS:-100000}"
CLIENTS="${CLIENTS:-50}"
DATA_SIZE="${DATA_SIZE:-3}"
SKIP_REDIS="${SKIP_REDIS:-0}"

DATE=$(date +%Y%m%d_%H%M%S)
OUTPUT_DIR="$SCRIPT_DIR/redis_benchmark"
OUTPUT_FILE="$OUTPUT_DIR/comparison_${DATE}.md"

mkdir -p "$OUTPUT_DIR"

# ============================================================
# Step 1: Discover supported commands from source code
# ============================================================
echo "==> Scanning $COMMANDS_DIR for supported commands..."

# Find all Name() string method return values in command source files.
# Pattern: lines that are purely `return "COMMAND_NAME"` (all uppercase).
# Excludes test files and mock files to avoid false positives.
DISCOVERED_COMMANDS=$(grep -rh \
    --include="*.go" \
    --exclude="*_test.go" \
    --exclude-dir="mock" \
    '^\s*return "[A-Z][A-Z]*"\s*$' \
    "$COMMANDS_DIR" \
    | grep -oE '"[A-Z]+"' \
    | tr -d '"' \
    | sort -u)

if [ -z "$DISCOVERED_COMMANDS" ]; then
    echo "ERROR: No commands discovered in $COMMANDS_DIR"
    exit 1
fi

echo "==> Discovered commands:"
for cmd in $DISCOVERED_COMMANDS; do
    echo "    - $cmd"
done

# ============================================================
# Step 2: Map discovered commands to redis-benchmark test types
# ============================================================
# redis-benchmark supports these test types via the -t flag:
REDIS_BENCH_SUPPORTED="set get incr lpush rpush lpop rpop sadd hset spop zadd zpopmin lrange xadd ping_inline ping_mbulk"

BENCH_COMMANDS=""
SKIPPED_COMMANDS=""

for cmd in $DISCOVERED_COMMANDS; do
    lower_cmd=$(echo "$cmd" | tr '[:upper:]' '[:lower:]')
    if echo " $REDIS_BENCH_SUPPORTED " | grep -q " ${lower_cmd} "; then
        BENCH_COMMANDS="${BENCH_COMMANDS:+$BENCH_COMMANDS,}$lower_cmd"
    else
        SKIPPED_COMMANDS="${SKIPPED_COMMANDS:+$SKIPPED_COMMANDS,}$cmd"
    fi
done

echo ""
echo "==> Benchmarkable commands:  $BENCH_COMMANDS"
if [ -n "$SKIPPED_COMMANDS" ]; then
    echo "==> Skipped (no redis-benchmark support): $SKIPPED_COMMANDS"
fi
echo ""

# ============================================================
# Step 3: Prerequisite checks
# ============================================================
if ! command -v redis-benchmark &>/dev/null; then
    echo "ERROR: redis-benchmark not found. Install Redis tools first."
    echo "  macOS: brew install redis"
    exit 1
fi

if ! command -v redis-cli &>/dev/null; then
    echo "ERROR: redis-cli not found. Install Redis tools first."
    echo "  macOS: brew install redis"
    exit 1
fi

# ============================================================
# Step 4: Build Avacado
# ============================================================
echo "==> Building Avacado..."
(cd "$REPO_ROOT" && go build -o "$AVACADO_BIN" ./cmd/server/main.go)
echo "==> Build complete"

# ============================================================
# Step 5: Start Avacado server
# ============================================================
AVACADO_PID=""

cleanup() {
    if [ -n "$AVACADO_PID" ]; then
        echo ""
        echo "==> Stopping Avacado (PID $AVACADO_PID)..."
        kill "$AVACADO_PID" 2>/dev/null || true
        wait "$AVACADO_PID" 2>/dev/null || true
    fi
    rm -f "$AVACADO_BIN"
}
trap cleanup EXIT

echo "==> Starting Avacado on port $AVACADO_PORT..."
"$AVACADO_BIN" --port "$AVACADO_PORT" >/dev/null 2>&1 &
AVACADO_PID=$!

# Wait for Avacado to be ready (up to 10 seconds)
echo "==> Waiting for Avacado to be ready..."
for i in $(seq 1 20); do
    if redis-cli -p "$AVACADO_PORT" ping >/dev/null 2>&1; then
        echo "==> Avacado is ready (PID $AVACADO_PID)"
        break
    fi
    if [ "$i" -eq 20 ]; then
        echo "ERROR: Avacado failed to start within 10 seconds"
        exit 1
    fi
    sleep 0.5
done

# ============================================================
# Step 6: Run redis-benchmark against Avacado
# ============================================================
echo ""
echo "==> Benchmarking Avacado (port $AVACADO_PORT)..."
echo "    Commands: $BENCH_COMMANDS"
echo "    Requests: $REQUESTS  |  Clients: $CLIENTS  |  Data: ${DATA_SIZE}B"
echo ""

AVACADO_CSV=$(redis-benchmark \
    -h localhost \
    -p "$AVACADO_PORT" \
    -t "$BENCH_COMMANDS" \
    -n "$REQUESTS" \
    -c "$CLIENTS" \
    -d "$DATA_SIZE" \
    --csv 2>/dev/null)

echo "$AVACADO_CSV"

# ============================================================
# Step 7: Run redis-benchmark against Redis (optional)
# ============================================================
REDIS_CSV=""
REDIS_AVAILABLE=0

if [ "$SKIP_REDIS" -eq 0 ]; then
    if redis-cli -p "$REDIS_PORT" ping >/dev/null 2>&1; then
        REDIS_AVAILABLE=1
        echo ""
        echo "==> Benchmarking Redis (port $REDIS_PORT)..."
        echo ""
        REDIS_CSV=$(redis-benchmark \
            -h localhost \
            -p "$REDIS_PORT" \
            -t "$BENCH_COMMANDS" \
            -n "$REQUESTS" \
            -c "$CLIENTS" \
            -d "$DATA_SIZE" \
            --csv 2>/dev/null)
        echo "$REDIS_CSV"
    else
        echo ""
        echo "==> Redis not available on port $REDIS_PORT — skipping comparison"
        echo "    Start Redis and re-run, or set SKIP_REDIS=1 to suppress this message"
    fi
fi

# ============================================================
# Step 8: Generate markdown report
# ============================================================
echo ""
echo "==> Generating report: $OUTPUT_FILE"

GO_VERSION=$(go version 2>/dev/null | awk '{print $3}')
PLATFORM="$(uname -s) $(uname -m)"
CPU_INFO=$(sysctl -n machdep.cpu.brand_string 2>/dev/null \
    || grep -m1 'model name' /proc/cpuinfo 2>/dev/null | cut -d: -f2 | xargs \
    || echo "Unknown")

{
cat <<HEADER
# Avacado vs Redis Benchmark Comparison

**Date:** $(date '+%Y-%m-%d %H:%M:%S')
**Platform:** $PLATFORM
**Go Version:** $GO_VERSION
**CPU:** $CPU_INFO

## Discovered Commands

Commands were automatically discovered by scanning \`internal/command/\` source files
for \`Name() string\` method implementations.

| Command | Benchmarkable | Notes |
|---------|:------------:|-------|
HEADER

for cmd in $DISCOVERED_COMMANDS; do
    lower_cmd=$(echo "$cmd" | tr '[:upper:]' '[:lower:]')
    if echo " $REDIS_BENCH_SUPPORTED " | grep -q " ${lower_cmd} "; then
        echo "| \`$cmd\` | ✅ | Included in benchmark |"
    else
        echo "| \`$cmd\` | ❌ | Not a redis-benchmark test type |"
    fi
done

cat <<CONFIG

## Test Configuration

| Parameter | Value |
|-----------|-------|
| Commands benchmarked | \`$BENCH_COMMANDS\` |
| Requests per command | $REQUESTS |
| Concurrent clients | $CLIENTS |
| Data size | ${DATA_SIZE} bytes |
| Avacado port | $AVACADO_PORT |
| Redis port | $REDIS_PORT |

## Results

CONFIG

# ---- Helper: strip quotes from a CSV field ----
strip() { echo "$1" | tr -d '"'; }

if [ "$REDIS_AVAILABLE" -eq 1 ]; then
cat <<'TABLE_HEADER'
### Avacado vs Redis Summary

| Command | Avacado (req/s) | Redis (req/s) | Ratio | Avacado p50 | Redis p50 | Avacado p99 | Redis p99 |
|---------|----------------:|-------------:|------:|------------:|----------:|------------:|----------:|
TABLE_HEADER

    # Parse avacado CSV, skip header line
    while IFS=',' read -r test rps avg_lat min_lat p50 p95 p99 max_lat; do
        test=$(strip "$test")
        [ "$test" = "test" ] && continue

        a_rps=$(strip "$rps")
        a_p50=$(strip "$p50")
        a_p99=$(strip "$p99")

        # Look for matching Redis result (test name matches)
        redis_line=$(echo "$REDIS_CSV" | grep "^\"${test}\"" 2>/dev/null || echo "")
        if [ -n "$redis_line" ]; then
            r_rps=$(echo "$redis_line" | cut -d',' -f2 | tr -d '"')
            r_p50=$(echo "$redis_line" | cut -d',' -f5 | tr -d '"')
            r_p99=$(echo "$redis_line" | cut -d',' -f7 | tr -d '"')
            ratio=$(awk "BEGIN {printf \"%.1f%%\", ($a_rps / $r_rps) * 100}" 2>/dev/null || echo "N/A")
            echo "| **$test** | $a_rps | $r_rps | $ratio | ${a_p50}ms | ${r_p50}ms | ${a_p99}ms | ${r_p99}ms |"
        else
            echo "| **$test** | $a_rps | — | — | ${a_p50}ms | — | ${a_p99}ms | — |"
        fi
    done < <(echo "$AVACADO_CSV")

    echo ""
    echo "### Detailed Metrics"
    echo ""

    # Per-command detailed tables
    while IFS=',' read -r test rps avg_lat min_lat p50 p95 p99 max_lat; do
        test=$(strip "$test")
        [ "$test" = "test" ] && continue

        a_rps=$(strip "$rps")
        a_avg=$(strip "$avg_lat")
        a_min=$(strip "$min_lat")
        a_p50=$(strip "$p50")
        a_p95=$(strip "$p95")
        a_p99=$(strip "$p99")
        a_max=$(strip "$max_lat")

        redis_line=$(echo "$REDIS_CSV" | grep "^\"${test}\"" 2>/dev/null || echo "")

        echo "#### $test Command"
        echo ""
        echo "| Metric | Avacado | Redis | Ratio |"
        echo "|--------|--------:|------:|------:|"

        if [ -n "$redis_line" ]; then
            r_rps=$(echo "$redis_line" | cut -d',' -f2 | tr -d '"')
            r_avg=$(echo "$redis_line" | cut -d',' -f3 | tr -d '"')
            r_min=$(echo "$redis_line" | cut -d',' -f4 | tr -d '"')
            r_p50=$(echo "$redis_line" | cut -d',' -f5 | tr -d '"')
            r_p95=$(echo "$redis_line" | cut -d',' -f6 | tr -d '"')
            r_p99=$(echo "$redis_line" | cut -d',' -f7 | tr -d '"')
            r_max=$(echo "$redis_line" | cut -d',' -f8 | tr -d '"')

            ratio_rps=$(awk "BEGIN {printf \"%.1f%%\", ($a_rps/$r_rps)*100}" 2>/dev/null || echo "N/A")
            ratio_p50=$(awk "BEGIN {printf \"%.1f%%\", ($r_p50/$a_p50)*100}" 2>/dev/null || echo "N/A")
            ratio_p99=$(awk "BEGIN {printf \"%.1f%%\", ($r_p99/$a_p99)*100}" 2>/dev/null || echo "N/A")

            echo "| Throughput (req/s) | $a_rps | $r_rps | $ratio_rps |"
            echo "| Avg Latency (ms)   | $a_avg | $r_avg | — |"
            echo "| Min Latency (ms)   | $a_min | $r_min | — |"
            echo "| p50 Latency (ms)   | $a_p50 | $r_p50 | — |"
            echo "| p95 Latency (ms)   | $a_p95 | $r_p95 | — |"
            echo "| p99 Latency (ms)   | $a_p99 | $r_p99 | — |"
            echo "| Max Latency (ms)   | $a_max | $r_max | — |"
        else
            echo "| Throughput (req/s) | $a_rps | — | — |"
            echo "| Avg Latency (ms)   | $a_avg | — | — |"
            echo "| Min Latency (ms)   | $a_min | — | — |"
            echo "| p50 Latency (ms)   | $a_p50 | — | — |"
            echo "| p95 Latency (ms)   | $a_p95 | — | — |"
            echo "| p99 Latency (ms)   | $a_p99 | — | — |"
            echo "| Max Latency (ms)   | $a_max | — | — |"
        fi
        echo ""
    done < <(echo "$AVACADO_CSV")

else
    # No Redis — Avacado-only results table
    echo "### Avacado Results"
    echo ""
    echo "| Command | req/s | Avg (ms) | Min (ms) | p50 (ms) | p95 (ms) | p99 (ms) | Max (ms) |"
    echo "|---------|------:|---------:|---------:|---------:|---------:|---------:|---------:|"

    while IFS=',' read -r test rps avg_lat min_lat p50 p95 p99 max_lat; do
        test=$(strip "$test")
        [ "$test" = "test" ] && continue
        echo "| **$test** | $(strip $rps) | $(strip $avg_lat) | $(strip $min_lat) | $(strip $p50) | $(strip $p95) | $(strip $p99) | $(strip $max_lat) |"
    done < <(echo "$AVACADO_CSV")

    echo ""
fi

cat <<'RAW_DATA'
## Raw CSV Data

### Avacado
```csv
RAW_DATA
echo "$AVACADO_CSV"
echo '```'
echo ""

if [ "$REDIS_AVAILABLE" -eq 1 ]; then
cat <<'REDIS_RAW'
### Redis
```csv
REDIS_RAW
    echo "$REDIS_CSV"
    echo '```'
    echo ""
fi

cat <<REPRO
## Reproduction

\`\`\`bash
# Build and start Avacado
go build -o main ./cmd/server/main.go
./main --port $AVACADO_PORT &

# Run benchmark
redis-benchmark -h localhost -p $AVACADO_PORT \\
    -t $BENCH_COMMANDS \\
    -n $REQUESTS -c $CLIENTS -d $DATA_SIZE --csv
\`\`\`
REPRO

} >"$OUTPUT_FILE"

echo "==> Report saved: $OUTPUT_FILE"
echo ""
echo "==> Done!"
