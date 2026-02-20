# Avacado Benchmark Findings

**Test Date:** February 20, 2026 (latest), February 9, 2026 (baseline)
**Test Configuration:** 100,000 requests, 50 concurrent clients
**Commands Tested:** SET, GET
**Comparison:** Avacado vs Official Redis Server

---

## Executive Summary

Avacado achieves **63-65% of Redis performance** for basic SET/GET operations, placing it solidly in the **"Acceptable"** performance range (50-80% target). The implementation demonstrates production-ready throughput (>100K ops/sec) with excellent tail latency characteristics.

---

## Performance Metrics

### SET Command

```
Metric                  Avacado         Redis           Ratio
─────────────────────────────────────────────────────────────
Throughput (req/s)      117,786         189,036         62.3%
Average Latency         0.241 ms        0.152 ms        158.6%
p50 (median)            0.239 ms        0.135 ms        177.0%
p95                     0.279 ms        0.295 ms        94.6% ✓
p99                     0.687 ms        0.639 ms        107.5%
Max Latency             1.535 ms        1.127 ms        136.2%
```

### GET Command

```
Metric                  Avacado         Redis           Ratio
─────────────────────────────────────────────────────────────
Throughput (req/s)      118,765         182,815         65.0%
Average Latency         0.233 ms        0.154 ms        151.3%
p50 (median)            0.239 ms        0.135 ms        177.0%
p95                     0.271 ms        0.271 ms        100.0% ✓
p99                     0.327 ms        0.359 ms        91.1% ✓
Max Latency             0.823 ms        0.479 ms        171.8%
```

---

## Key Findings

### ✅ Strengths

1. **Excellent Tail Latency Performance**
   - GET p95: Matches Redis exactly (0.271 ms)
   - GET p99: Actually 9% better than Redis (0.327 vs 0.359 ms)
   - SET p95: Within 5% of Redis (0.279 vs 0.295 ms)
   - Demonstrates good consistency under load

2. **Production-Ready Throughput**
   - Both operations exceed 100K ops/sec target
   - SET: 117,786 req/s
   - GET: 118,765 req/s
   - Consistent performance across operation types

3. **Meets Performance Goals**
   - ✓ Throughput > 100K ops/sec
   - ✓ 50-80% of Redis performance (achieved 63-65%)
   - ✓ Sub-millisecond p99 latency

4. **Balanced Read/Write Performance**
   - SET and GET have nearly identical throughput
   - Suggests good architectural balance
   - No obvious bias toward reads or writes

### ⚠️ Areas for Improvement

1. **Median Latency (p50) Gap**
   - 77% slower than Redis (0.239 vs 0.135 ms)
   - ~0.1ms difference affects typical user experience
   - Represents the biggest optimization opportunity

2. **Average Latency**
   - 51-59% higher than Redis
   - Indicates overhead in the common case
   - Likely due to protocol parsing or storage layer

3. **Maximum Latency Spikes**
   - SET: 1.535 ms vs Redis 1.127 ms (36% higher)
   - GET: 0.823 ms vs Redis 0.479 ms (72% higher)
   - Possible causes: GC pauses, lock contention, or scheduler delays

---

## Performance Analysis

### What's Working Well

**Tail Latency Optimization:**
The strong p95/p99 performance suggests:
- Efficient lock management under contention
- Good goroutine scheduling
- Minimal GC impact on tail latencies
- Effective concurrent request handling

**Throughput Scaling:**
Consistent 118K req/s across both operations indicates:
- No obvious bottlenecks in the storage layer
- Balanced read/write path performance
- Effective connection handling

### Bottleneck Hypotheses

Based on the performance gap at p50:

1. **Protocol Overhead (~40% of gap)**
   - RESP parsing/serialization adds latency
   - Buffer allocation/copying overhead
   - String conversions in command parsing

2. **Storage Layer (~30% of gap)**
   - sync.RWMutex overhead vs Redis custom implementation
   - Lock upgrade pattern in Get() operation
   - Lazy expiry checks on every operation

3. **Go Runtime (~20% of gap)**
   - Goroutine scheduling overhead
   - Interface dispatch costs
   - Memory allocations triggering GC

4. **Server Infrastructure (~10% of gap)**
   - Per-connection goroutine overhead
   - Logger overhead (structured logging)
   - Error handling and validation

---

## Optimization Roadmap

### Priority 1: Hot Path Optimization (Target: +20-30% throughput)

**Actions:**
1. Profile with pprof to identify hotspots
   ```bash
   go test -cpuprofile=cpu.prof -bench=BenchmarkGet
   go tool pprof cpu.prof
   ```

2. Focus on critical path functions:
   - RESP parser/serializer
   - Storage Get/Set operations
   - Command registry lookup

3. Measure allocation overhead
   ```bash
   go test -benchmem -bench=.
   ```

**Expected Impact:** Reduce p50 latency by 0.05-0.08ms

### Priority 2: Lock Optimization (Target: +10-15% throughput)

**Investigations:**
1. Analyze RWMutex contention at 50 clients
2. Consider sharded storage (multiple mutexes)
3. Evaluate lock-free data structures for hot keys
4. Profile lock upgrade pattern in Get()

**Expected Impact:** Better scaling under higher concurrency

### Priority 3: Protocol Optimization (Target: +5-10% throughput)

**Optimizations:**
1. Pool buffers for parsing/serialization
2. Reduce string allocations in RESP layer
3. Optimize CRLF scanning
4. Consider zero-copy techniques

**Expected Impact:** Reduce average latency by 0.02-0.03ms

### Priority 4: Memory Optimization (Target: Reduce GC impact)

**Actions:**
1. Profile memory allocations
2. Use sync.Pool for temporary objects
3. Reduce interface{} usage
4. Pre-allocate buffers where possible

**Expected Impact:** Lower max latency, reduce p99 spikes

---

## Next Benchmark Tests

### Immediate Tests

1. **Concurrency Scaling**
   ```bash
   for c in 1 10 50 100 500 1000; do
     redis-benchmark -h localhost -p 6380 -t get,set -n 100000 -c $c -q
   done
   ```
   Goal: Identify when performance degrades

2. **Value Size Impact**
   ```bash
   for size in 10 100 1000 10000; do
     redis-benchmark -h localhost -p 6380 -t set,get -n 100000 -d $size -q
   done
   ```
   Goal: Understand serialization overhead

3. **Pipeline Performance**
   ```bash
   redis-benchmark -h localhost -p 6380 -t set,get -n 100000 -P 16 -q
   ```
   Goal: Test batching efficiency

### Follow-up Tests

4. **Mixed Workload Patterns**
   - 90% GET / 10% SET (read-heavy cache)
   - 50% GET / 50% SET (balanced)
   - 10% GET / 90% SET (write-heavy)

5. **Multi-Key Operations**
   - DEL with 1, 10, 100, 1000 keys
   - EXISTS with varying key counts

6. **Expiry Performance**
   - Benchmark with TTL-enabled keys
   - Measure lazy expiry overhead
   - Test background cleanup impact

---

## Performance Goals vs Actuals

| Goal | Target | Actual | Status |
|------|--------|--------|--------|
| Throughput | > 100K ops/sec | 118K ops/sec | ✅ **EXCEEDED** |
| Redis Comparison | 50-80% | 63-65% | ✅ **MET** |
| p99 Latency | < 1ms | 0.327-0.687ms | ✅ **MET** |
| Concurrency | Linear to 100 clients | TBD | ⏳ **PENDING** |
| GC Pause | < 1ms p99 | TBD | ⏳ **PENDING** |

---

## Conclusions

### Overall Assessment: ✅ **Production Ready**

Avacado demonstrates solid performance characteristics suitable for production use cases where:
- 100K+ ops/sec is sufficient
- Sub-millisecond latencies are acceptable
- Consistent tail latency is critical

### Performance Tier: **Acceptable** (63-65% of Redis)

The implementation falls within the 50-80% acceptable range defined in the benchmarking guide. With targeted optimizations, reaching 70-80% is achievable.

### Recommended Use Cases

**Well-Suited For:**
- Development/testing environments
- Microservices with moderate traffic
- Cache layer for web applications
- Learning Redis protocol implementation
- Internal tools and services

**Consider Alternatives For:**
- Ultra-high throughput requirements (>500K ops/sec)
- Sub-0.1ms latency requirements
- Applications where every microsecond matters

### Next Steps

1. **Immediate:** Profile hot path with pprof (Priority 1)
2. **Short-term:** Run concurrency scaling tests (Priority 2)
3. **Medium-term:** Implement identified optimizations (Priorities 1-3)
4. **Long-term:** Re-benchmark and compare improvements

---

## Technical Environment

- **Avacado Version:** Latest (commit e3cd621)
- **Redis Version:** Official (homebrew)
- **Platform:** macOS Darwin 25.2.0
- **Go Version:** (check with `go version`)
- **CPU:** (check with `sysctl -n machdep.cpu.brand_string`)
- **Test Tool:** redis-benchmark (official)
- **Test Date:** February 9, 2026

---

## Related Documents

- Detailed comparison: `benchmarks/redis_benchmark/comparison_20260209.md`
- Benchmarking guide: `benchmarks/BENCHMARKING_GUIDE.md`
- Raw results: `benchmarks/redis_benchmark/` directory

---

## Benchmark Commands Used

### Avacado Server
```bash
./main --port 6380
```

### Redis Server
```bash
redis-server  # Default port 6379
```

### Benchmark Execution
```bash
# Quick comparison
redis-benchmark -h localhost -p [PORT] -t set,get -n 100000 -c 50 -q

# Detailed CSV output
redis-benchmark -h localhost -p [PORT] -t set,get -n 100000 -c 50 --csv
```

---

*Last Updated: February 20, 2026*

---

## Benchmark History

### February 20, 2026 (Full Command Suite — Automated Benchmark)

**Configuration:** 100K requests, 50 clients, 3-byte data, all benchmarkable commands
**Platform:** Darwin arm64 (Apple M4 Pro), Go 1.26.0
**Script:** `benchmarks/run_benchmark.sh` (auto-discovers commands from source)

| Command | Avacado (req/s) | Redis (req/s) | Ratio |
|---------|----------------:|---------------:|------:|
| SET     | 121,065         | 173,010        | 70.0% |
| GET     | 121,655         | 194,932        | 62.4% |
| INCR    | 121,065         | 193,050        | 62.7% |
| LPUSH   | 118,064         | 194,175        | 60.8% |
| RPUSH   | 121,655         | 194,553        | 62.5% |
| LPOP    | 122,549         | 204,082        | 60.0% |
| RPOP    | 122,850         | 195,695        | 62.8% |

**Key findings:**
- All 7 benchmarkable commands exceed **118K req/s** — well above the 100K target
- Overall range: **60–70% of Redis throughput** (within acceptable 50–80% band)
- SET is the strongest at **70.0%** of Redis
- LPOP/RPOP have **better p99 than Redis** (0.327ms vs 0.383ms for RPOP)
- INCR p99 (0.303ms) beats Redis (0.399ms) by 24%
- List commands (LPUSH/RPUSH/LPOP/RPOP) perform on par with KV commands — no overhead from list storage layer
- p50 latency gap (~0.231ms vs 0.135ms) remains consistent across all commands

**Notes:** First benchmark covering the full command suite using the automated script. List commands introduced with no performance regression on KV operations.

Detailed report: `benchmarks/redis_benchmark/comparison_20260220_111503.md`

### February 20, 2026 (Post List Commands Refactor)

**Configuration:** 100K requests, 50 clients, 3-byte data, SET/GET
**Platform:** Darwin arm64 (Apple M4 Pro, 48GB RAM), Go 1.26.0

| Command | Avacado (req/s) | Redis (req/s) | Ratio |
|---------|-----------------|----------------|-------|
| SET | 119,617 | 193,424 | 61.8% |
| GET | 124,533 | 191,571 | 65.0% |

**Changes since last benchmark:**
- SET throughput: 117,786 -> 119,617 (+1.6%)
- GET throughput: 118,765 -> 124,533 (+4.9%)
- GET p95 improved: 0.271ms -> 0.255ms
- GET p99 improved: 0.327ms -> 0.295ms

**Notes:** Performance is stable after LPOP/RPOP refactor into unified Pop command. No regression detected. Slight improvements likely due to Go 1.26 and system differences.

Detailed report: `benchmarks/redis_benchmark/comparison_20260220_101307.md`

### February 9, 2026 (Initial Baseline)

| Command | Avacado (req/s) | Redis (req/s) | Ratio |
|---------|-----------------|----------------|-------|
| SET | 117,786 | 189,036 | 62.3% |
| GET | 118,765 | 182,815 | 65.0% |

Detailed report: `benchmarks/redis_benchmark/comparison_20260209.md`
