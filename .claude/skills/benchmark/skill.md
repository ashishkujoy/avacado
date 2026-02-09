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
2. Running redis-benchmark against avacado
3. Running redis-benchmark against Redis server
4. Comparing the results
5. Generating a comprehensive findings report
6. Cleaning up (stopping avacado server)

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
- `--commands` or `-t`: Commands to benchmark (default: `set,get`)
  - Examples: `set,get`, `incr,decr`, `set,get,incr,decr,del,exists`
- `--clients` or `-c`: Number of concurrent clients (default: `50`)
  - Suggested values: 1, 10, 50, 100, 500, 1000
- `--requests` or `-n`: Total number of requests (default: `100000`)
  - Suggested values: 10000, 100000, 1000000
- `--data-size` or `-d`: Data size in bytes (default: `3`)
  - Suggested values: 10, 100, 1000, 10000
- `--port`: Port for avacado server (default: `6380`)
- `--redis-port`: Port for Redis server (default: `6379`)

### Step 3: Start Avacado Server

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

### Step 4: Run Benchmarks

Run benchmarks in parallel for both servers:

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

### Step 5: Parse and Compare Results

1. **Parse CSV output** from both benchmarks
   - Extract metrics for each command
   - Store in structured format

2. **Calculate performance ratios**
   - Avacado throughput / Redis throughput (%)
   - Latency comparisons for each percentile
   - Identify strengths and weaknesses

3. **Categorize performance**
   - **Excellent:** 80-100% of Redis
   - **Good:** 60-80% of Redis
   - **Acceptable:** 50-60% of Redis
   - **Needs Work:** < 50% of Redis

### Step 6: Generate Report

Create a comprehensive markdown report in `benchmarks/redis_benchmark/`:

**Filename format:** `comparison_YYYYMMDD_HHMMSS.md`

**Report sections:**
1. **Test Configuration**
   - Date and time
   - Commands tested
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

### Step 7: Update Findings Document

If `benchmarks/BENCHMARK_FINDINGS.md` exists:
- Update the "Last Updated" date
- Add a new entry to benchmark history
- Update metrics if this is a new baseline

If running follow-up benchmarks after optimizations:
- Add a comparison section showing improvement
- Update optimization roadmap status

### Step 8: Cleanup

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

Configuration:
  • Commands: SET, GET
  • Requests: 100,000 per command
  • Clients: 50 concurrent
  • Data Size: 3 bytes

Results:

  SET Command:
    Avacado:  117,786 req/s | p50: 0.239ms | p99: 0.687ms
    Redis:    189,036 req/s | p50: 0.135ms | p99: 0.639ms
    Ratio:    62.3% ⚠️

  GET Command:
    Avacado:  118,765 req/s | p50: 0.239ms | p99: 0.327ms
    Redis:    182,815 req/s | p50: 0.135ms | p99: 0.359ms
    Ratio:    65.0% ✓

Overall Performance: Acceptable (63.6% of Redis)

✅ Strengths:
  • Excellent tail latency (GET p95 matches Redis)
  • Production-ready throughput (>100K ops/sec)

⚠️  Areas for Improvement:
  • Median latency 77% slower than Redis
  • Hot path optimization needed

📊 Detailed report: benchmarks/redis_benchmark/comparison_20260209_143025.md
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

### Custom Benchmark

```
/benchmark --commands lpush,rpush,lrange --clients 100 --requests 500000
```
Run benchmarks for specific commands with custom parameters.

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

## Error Handling

- If Redis server is not running, prompt user to start it
- If port is in use, suggest alternative port
- If avacado fails to start, check logs and display error
- If benchmarks fail, provide troubleshooting steps
- Handle timeout scenarios gracefully

## Examples

### Basic benchmark (default settings)
```
/benchmark
```
Runs SET and GET with 100K requests, 50 clients

### High concurrency test
```
/benchmark --clients 500 --requests 1000000
```
Stress test with 500 concurrent clients

### Multiple commands
```
/benchmark --commands set,get,incr,decr,del,exists
```
Comprehensive command benchmark

### Large value test
```
/benchmark --data-size 10000 --commands set,get
```
Test with 10KB values

## Integration with Development Workflow

### After implementing new features
```
/benchmark
```
Validate that performance hasn't regressed

### After optimization work
```
/benchmark --commands <optimized-commands>
```
Measure improvement and update findings

### Before releasing
```
/benchmark --test-type concurrency-scaling
```
Ensure scalability meets requirements

## Files Modified/Created

- `benchmarks/redis_benchmark/comparison_YYYYMMDD_HHMMSS.md` - Detailed report
- `benchmarks/BENCHMARK_FINDINGS.md` - Updated with latest results (optional)
- Temporary: `/tmp/avacado_benchmark_*.output` - Server logs

## Related Skills

- `/redis-command` - Implement new Redis commands
- Future: `/profile` - Run CPU/memory profiling
- Future: `/optimize` - Automated optimization suggestions
