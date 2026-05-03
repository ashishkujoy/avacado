# Performance Benchmark Skill

Runs comprehensive performance benchmarks comparing avacado against Redis server using redis-benchmark.

## Usage

```
/benchmark
/benchmark --commands set,get,incr
/benchmark --clients 100
/benchmark --requests 1000000
```

## What This Skill Does

This skill automates the performance benchmarking process by:
1. Starting avacado server on a test port
2. Auto-detecting all commands supported by avacado from source
3. Intersecting with commands redis-benchmark can test
4. Running redis-benchmark against avacado and Redis using those commands
5. Comparing the results
6. Generating a comprehensive findings report
7. Cleaning up (stopping avacado server)

## Process

### Step 1: Environment Check

1. **Check if Redis is installed**
   ```bash
   which redis-server && which redis-benchmark
   ```
   - If not installed, provide installation instructions
   - Recommend using homebrew on macOS: `brew install redis`

2. **Check for port conflicts**
   ```bash
   lsof -i :6379 -i :6380
   ```
   - Ensure Redis is running on port 6379 (or start it)
   - Ensure port 6380 is free for avacado
   - If ports are in use, suggest alternatives

3. **Verify avacado binary exists**
   - Check if `./main` exists
   - If not, build it: `go build -o main cmd/server/main.go`

### Step 2: Parse Arguments

Parse optional arguments from the user:
- `--commands` or `-t`: Commands to benchmark. **If not provided, auto-detect (see Step 3).**
  - Examples: `set,get`, `incr,decr`, `set,get,incr,decr,del,exists`
- `--clients` or `-c`: Number of concurrent clients (default: `50`)
  - Suggested values: 1, 10, 50, 100, 500, 1000
- `--requests` or `-n`: Total number of requests (default: `100000`)
  - Suggested values: 10000, 100000, 1000000
- `--data-size` or `-d`: Data size in bytes (default: `3`)
  - Suggested values: 10, 100, 1000, 10000
- `--port`: Port for avacado server (default: `6380`)
- `--redis-port`: Port for Redis server (default: `6379`)

### Step 3: Auto-Detect Supported Commands

**Skip this step if the user passed `--commands`.**

This step determines which commands to benchmark by reading avacado's source and intersecting with what redis-benchmark can test.

#### 3a. Extract avacado's registered commands from source

Run this grep to get every command name returned by a `Name()` method in the command packages:

```bash
grep -rh 'return "[A-Z]*"' internal/command/ --include='*.go' | grep -v mock | grep -oE '"[A-Z]+"' | tr -d '"' | sort -u
```

This produces the full list of commands avacado supports (e.g., SET, GET, INCR, LPUSH, LRANGE, HSET, PING, etc.).

#### 3b. Intersect with redis-benchmark's supported tests

`redis-benchmark -t` accepts these test names (case-insensitive):

```
ping, set, get, incr, lpush, rpush, lpop, rpop, sadd, hset, spop, zadd, zpopmin, lrange, mset, xadd
```

Build the intersection: keep only commands from 3a that appear in the above list. Convert to lowercase and join with commas for use as the `-t` argument.

**Example:** If avacado supports SET, GET, INCR, LPUSH, RPUSH, LPOP, RPOP, HSET, LRANGE, PING, DEL, EXISTS, DECR, TTL, BLPOP — the intersection is:

```
ping,set,get,incr,lpush,rpush,lpop,rpop,hset,lrange
```

(DEL, EXISTS, DECR, TTL, BLPOP, etc. are not in redis-benchmark's test suite so they are excluded.)

#### 3c. Display detected commands to user

Before running benchmarks, print:
```
Detected avacado commands:  SET, GET, INCR, LPUSH, RPUSH, LPOP, RPOP, HSET, LRANGE, PING, DEL, EXISTS, ...
redis-benchmark compatible: ping, set, get, incr, lpush, rpush, lpop, rpop, hset, lrange
Benchmarking:               ping,set,get,incr,lpush,rpush,lpop,rpop,hset,lrange
```

### Step 4: Start Avacado Server

1. **Start avacado in background**
   ```bash
   ./main --port 6380 &
   ```
   - Use the port specified by user or default 6380
   - Run as background task
   - Capture the process ID

2. **Wait for server to start**
   ```bash
   sleep 2 && lsof -i :6380
   ```
   - Verify the server is listening
   - If not started, check for errors and retry

3. **Test connectivity**
   ```bash
   redis-cli -p 6380 PING
   ```
   - Ensure avacado is responding
   - If connection fails, check server logs

### Step 5: Run Benchmarks

Run benchmarks in parallel for both servers using the command list from Step 3 (or user-supplied `--commands`):

1. **Benchmark Avacado**
   ```bash
   redis-benchmark -h localhost -p 6380 -t <commands> -n <requests> -c <clients> -d <data-size> --csv 2>/dev/null
   ```

2. **Benchmark Redis**
   ```bash
   redis-benchmark -h localhost -p 6379 -t <commands> -n <requests> -c <clients> -d <data-size> --csv
   ```

**Capture metrics:**
- Throughput (requests per second)
- Latency percentiles (p50, p95, p99)
- Min/Max/Average latency
- All data from CSV output

### Step 6: Parse and Compare Results

1. **Parse CSV output** from both benchmarks
   - Extract metrics for each command
   - Store in structured format

2. **Calculate performance ratios**
   - Avacado throughput / Redis throughput (%)
   - Latency comparisons for each percentile
   - Identify strengths and weaknesses

3. **Categorize performance**
   - **Excellent:** 80-100%+ of Redis
   - **Good:** 60-80% of Redis
   - **Acceptable:** 50-60% of Redis
   - **Needs Work:** < 50% of Redis

### Step 7: Generate Report

Create a comprehensive markdown report in `benchmarks/redis_benchmark/`:

**Filename format:** `comparison_YYYYMMDD_HHMMSS.md`

**Report sections:**
1. **Test Configuration**
   - Date and time
   - Commands tested (and how they were determined: auto-detected vs user-supplied)
   - Number of requests, clients, data size
   - Platform information

2. **Summary Results**
   - Table comparing each command
   - Throughput and latency metrics
   - Performance ratios

3. **Analysis**
   - Strengths (what's working well)
   - Areas for improvement
   - Bottleneck hypotheses

4. **Detailed Metrics**
   - Per-command breakdown
   - All latency percentiles
   - Raw CSV data

5. **Recommendations**
   - Optimization priorities
   - Next benchmark tests to run
   - Performance goals

6. **Raw Data**
   - Complete CSV output from both benchmarks
   - Commands used to reproduce
   - The grep command used to detect avacado's commands

### Step 8: Update Findings Document

If `benchmarks/BENCHMARK_FINDINGS.md` exists:
- Update the "Last Updated" date
- Add a new entry to benchmark history
- Update metrics if this is a new baseline

If running follow-up benchmarks after optimizations:
- Add a comparison section showing improvement
- Update optimization roadmap status

### Step 9: Cleanup

1. **Stop avacado server**
   ```bash
   pkill -f "main --port 6380"
   ```
   - Or use the captured process ID: `kill <PID>`

2. **Verify cleanup**
   ```bash
   lsof -i :6380
   ```
   - Ensure port is free

3. **Display summary to user**
   - Show key performance metrics
   - Highlight the most important findings
   - Provide path to detailed report

## Output Format

### Console Summary

```
╔══════════════════════════════════════════════════════════════╗
║           Avacado vs Redis Benchmark Results                 ║
╚══════════════════════════════════════════════════════════════╝

Detected avacado commands:  SET, GET, INCR, LPUSH, RPUSH, LPOP, RPOP, HSET, LRANGE, PING, ...
redis-benchmark compatible: ping, set, get, incr, lpush, rpush, lpop, rpop, hset, lrange
Benchmarking:               ping,set,get,incr,lpush,rpush,lpop,rpop,hset,lrange

Configuration:
  • Commands: PING, SET, GET, INCR, LPUSH, RPUSH, LPOP, RPOP, HSET, LRANGE
  • Requests: 100,000 per command
  • Clients: 50 concurrent
  • Data Size: 3 bytes

Results:

  SET Command:
    Avacado:  85,910 req/s | p50: 0.559ms | p99: 0.791ms
    Redis:    78,740 req/s | p50: 0.591ms | p99: 1.063ms
    Ratio:    109.1% ✅

  GET Command:
    Avacado:  85,397 req/s | p50: 0.567ms | p99: 0.767ms
    Redis:    86,655 req/s | p50: 0.591ms | p99: 0.791ms
    Ratio:    98.5% ✅

  ... (one block per benchmarked command)

Overall Performance: Excellent (avg XX% of Redis)

✅ Strengths:
  • ...

⚠️  Areas for Improvement:
  • ...

📊 Detailed report: benchmarks/redis_benchmark/comparison_YYYYMMDD_HHMMSS.md
```

### Report File

Generate detailed markdown report with:
- All metrics in tables
- Visual comparisons
- Performance analysis
- Optimization recommendations
- Next steps

## Advanced Usage

### Concurrency Scaling Test

```
/benchmark --test-type concurrency-scaling
```
Runs benchmarks with varying client counts (1, 10, 50, 100, 500, 1000) and generates scaling analysis.

### Data Size Impact Test

```
/benchmark --test-type data-size
```
Runs benchmarks with varying data sizes (10B, 100B, 1KB, 10KB) to measure serialization overhead.

### Pipeline Performance Test

```
/benchmark --test-type pipeline
```
Tests with different pipeline sizes (-P flag) to measure batching efficiency.

### Custom Benchmark (override auto-detection)

```
/benchmark --commands lpush,rpush,lrange --clients 100 --requests 500000
```
Run benchmarks for a specific subset of commands, bypassing auto-detection.

## Performance Goals Reference

From `BENCHMARKING_GUIDE.md`:

- **Throughput:** > 100K ops/sec
- **Latency p99:** < 1ms (under moderate load)
- **Redis Comparison:** 50-80% acceptable, 80%+ excellent
- **Concurrency:** Linear scaling up to 100 concurrent clients

## Notes

- Always benchmark on production-like hardware
- Run benchmarks multiple times for consistency
- Document system configuration (OS, Go version, CPU, RAM)
- Keep benchmark history for regression detection
- Warm up the system before measurements (first few requests)
- The auto-detected command list will grow automatically as new commands are added to `internal/command/registry/registry.go`

## Error Handling

- If Redis server is not running, prompt user to start it
- If port is in use, suggest alternative port
- If avacado fails to start, check logs and display error
- If benchmarks fail, provide troubleshooting steps
- Handle timeout scenarios gracefully
- If grep finds no commands (e.g., source structure changed), fall back to `set,get` and warn the user

## Examples

### Basic benchmark (auto-detects all supported commands)
```
/benchmark
```
Greps source, finds intersection with redis-benchmark, runs all supported commands.

### High concurrency test
```
/benchmark --clients 500 --requests 1000000
```
Stress test with 500 concurrent clients across all auto-detected commands.

### Override with specific commands
```
/benchmark --commands set,get,incr
```
Bypasses auto-detection and benchmarks only these three commands.

### Large value test
```
/benchmark --data-size 10000
```
Tests with 10KB values across all auto-detected commands.

## Integration with Development Workflow

### After implementing new commands
```
/benchmark
```
Auto-detection picks up the new command automatically — no skill update needed.

### After optimization work
```
/benchmark --commands <optimized-commands>
```
Measure improvement on specific commands and update findings.

### Before releasing
```
/benchmark --test-type concurrency-scaling
```
Ensure scalability meets requirements across all supported commands.

## Files Modified/Created

- `benchmarks/redis_benchmark/comparison_YYYYMMDD_HHMMSS.md` - Detailed report
- `benchmarks/BENCHMARK_FINDINGS.md` - Updated with latest results (optional)
- Temporary: `/tmp/avacado_benchmark_*.output` - Server logs

## Related Skills

- `/redis-command` - Implement new Redis commands
- Future: `/profile` - Run CPU/memory profiling
- Future: `/optimize` - Automated optimization suggestions
