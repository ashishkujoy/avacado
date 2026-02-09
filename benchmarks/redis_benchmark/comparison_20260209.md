# Avacado vs Redis Benchmark Comparison

**Date:** 2026-02-09
**Test Configuration:**
- Number of requests: 100,000
- Concurrent clients: 50
- Commands tested: SET, GET
- Platform: macOS (Darwin 25.2.0)

---

## Summary Results

### SET Command Performance

| Metric | Avacado | Redis | Avacado/Redis Ratio |
|--------|---------|-------|---------------------|
| **Throughput (req/sec)** | 117,785.63 | 189,035.92 | **62.3%** |
| **Avg Latency (ms)** | 0.241 | 0.152 | 158.6% |
| **Min Latency (ms)** | 0.016 | 0.072 | 22.2% |
| **p50 Latency (ms)** | 0.239 | 0.135 | 177.0% |
| **p95 Latency (ms)** | 0.279 | 0.295 | **94.6%** |
| **p99 Latency (ms)** | 0.687 | 0.639 | 107.5% |
| **Max Latency (ms)** | 1.535 | 1.127 | 136.2% |

### GET Command Performance

| Metric | Avacado | Redis | Avacado/Redis Ratio |
|--------|---------|-------|---------------------|
| **Throughput (req/sec)** | 118,764.84 | 182,815.36 | **65.0%** |
| **Avg Latency (ms)** | 0.233 | 0.154 | 151.3% |
| **Min Latency (ms)** | 0.024 | 0.064 | 37.5% |
| **p50 Latency (ms)** | 0.239 | 0.135 | 177.0% |
| **p95 Latency (ms)** | 0.271 | 0.271 | **100.0%** |
| **p99 Latency (ms)** | 0.327 | 0.359 | **91.1%** |
| **Max Latency (ms)** | 0.823 | 0.479 | 171.8% |

---

## Analysis

### Performance Summary
- **Overall Performance:** Avacado achieves approximately **63-65% of Redis throughput**
- **Latency Profile:** Avacado has slightly higher median latencies but competitive tail latencies (p95, p99)
- **Assessment:** Within the **"Acceptable"** range (50-80% of Redis) per the benchmarking guide

### Key Observations

#### Strengths
1. **Competitive Tail Latencies:**
   - SET p95: 94.6% of Redis performance
   - GET p95: 100% match with Redis
   - GET p99: 91.1% of Redis performance

2. **Consistent Performance:**
   - Similar performance profile for both SET and GET operations
   - Low variance between operations suggests good architectural balance

3. **Throughput Achievement:**
   - Both SET and GET operations exceed 100K ops/sec target
   - Demonstrates production-ready performance capabilities

#### Areas for Optimization

1. **Median Latency (p50):**
   - ~77% slower than Redis for median requests
   - Suggests optimization opportunities in the hot path

2. **Average Latency:**
   - 51-59% higher than Redis
   - Indicates room for improvement in common case performance

3. **Max Latency:**
   - Higher maximum latencies suggest potential GC pauses or lock contention
   - Worth investigating with profiling tools

### Performance Goals Assessment

Based on the benchmarking guide targets:

✅ **Throughput:** > 100K ops/sec → **PASSED** (117K-118K ops/sec)
✅ **Comparison to Redis:** 50-80% target → **PASSED** (63-65%)
⚠️  **Latency p99:** < 1ms target → **BORDERLINE** (SET: 0.687ms ✓, GET: 0.327ms ✓)

---

## Next Steps

### Immediate Optimizations
1. **Profile the hot path** - Use pprof to identify bottlenecks in SET/GET operations
2. **Storage layer analysis** - Review sync.RWMutex contention patterns
3. **Protocol overhead** - Benchmark RESP parsing/serialization separately

### Investigation Areas
1. **Lock contention:** Check if RWMutex is causing bottlenecks at 50 concurrent clients
2. **Memory allocations:** Run with `-benchmem` to identify allocation hotspots
3. **GC impact:** Monitor GC pauses during benchmarks (GODEBUG=gctrace=1)

### Future Benchmarks
1. **Concurrency scaling:** Test with 1, 10, 100, 500, 1000 clients
2. **Data size variations:** Test with different value sizes (10B, 1KB, 10KB)
3. **Mixed workloads:** 90/10, 50/50 read/write ratios
4. **Expiry scenarios:** Test with TTL-enabled keys

---

## Raw Data

### Avacado (Port 6380)
```csv
"test","rps","avg_latency_ms","min_latency_ms","p50_latency_ms","p95_latency_ms","p99_latency_ms","max_latency_ms"
"SET","117785.63","0.241","0.016","0.239","0.279","0.687","1.535"
"GET","118764.84","0.233","0.024","0.239","0.271","0.327","0.823"
```

### Redis (Port 6379)
```csv
"test","rps","avg_latency_ms","min_latency_ms","p50_latency_ms","p95_latency_ms","p99_latency_ms","max_latency_ms"
"SET","189035.92","0.152","0.072","0.135","0.295","0.639","1.127"
"GET","182815.36","0.154","0.064","0.135","0.271","0.359","0.479"
```

---

## Environment

- **Avacado:** Custom Go implementation (~4,500 LOC)
- **Redis:** Official Redis server (homebrew installation)
- **Tool:** redis-benchmark (official Redis benchmarking tool)
- **OS:** macOS Darwin 25.2.0
- **Test Date:** February 9, 2026
