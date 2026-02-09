# Avacado Benchmarking Guide

## Project Overview

Avacado is a Redis-compatible in-memory key-value store written in Go (~4,500 lines of code) that implements the RESP (REdis Serialization Protocol) and several Redis commands.

**Current Implementation Status:**
- RESP protocol parser and serializer
- TCP server with concurrent connection handling
- In-memory storage with expiry support
- Commands: SET, GET, INCR, DECR, DECRBY, DEL, EXISTS, TTL, PTTL, HELLO, CLIENT

---

## Key Areas to Benchmark

### 1. Protocol Layer (Critical Path)

#### RESP Protocol Parsing
**Location:** `internal/protocol/resp/parser.go`

**What to benchmark:**
- Parsing different RESP data types:
  - Simple Strings (`+OK\r\n`)
  - Bulk Strings (`$5\r\nhello\r\n`)
  - Integers (`:1000\r\n`)
  - Arrays (`*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n`)
  - Maps (`%2\r\n+key1\r\n+val1\r\n+key2\r\n+val2\r\n`)
- Buffer reading efficiency with `bufio.Reader`
- CRLF parsing overhead
- Nested data structure parsing (arrays of arrays, maps)

**Why it matters:** Every incoming request must be parsed from RESP format. This is on the critical path for every operation.

#### RESP Protocol Serialization
**Location:** `internal/protocol/resp/serializer.go`

**What to benchmark:**
- Serializing responses back to RESP format
- Buffer writing efficiency with `bytes.Buffer`
- String conversions and allocations
- Different response types serialization

**Why it matters:** Every response must be serialized back to RESP format before sending to the client.

---

### 2. Storage Operations (Core System)

**Location:** `internal/storage/kv/memory/memory.go`

#### Core Operations

**Get()**
- Read operation with expiry check
- Contains TOCTOU lock upgrade pattern (RLock → Unlock → Lock)
- Lazy expiry deletion

**Set()**
- Write operation with options:
  - `NX` - Only set if key doesn't exist
  - `XX` - Only set if key exists
  - `EX` - Set expiry time
  - `GET` - Return old value

**Incr(), Decr(), DecrBy()**
- Atomic integer operations
- String to int conversion overhead
- Int to string conversion on write

**Del()**
- Multi-key deletion
- Count of successfully deleted keys

**Exists()**
- Multi-key existence check
- Read-lock efficiency with multiple keys

#### Concurrency Performance

**What to benchmark:**
- `sync.RWMutex` contention under high concurrent load
- Read vs Write lock performance
- Lock upgrade pattern in `Get()` performance
- Concurrent read scaling (multiple readers)
- Concurrent write scaling (single writer bottleneck)
- Mixed read-write workloads

**Key concerns:**
- Lock contention at high concurrency
- Fairness between readers and writers
- Impact of lock upgrades in Get()

#### Expiry Handling

**Background Cleanup:**
- Runs every 1 second
- Scans all keys and deletes expired ones
- Impact on overall performance

**Lazy Expiry:**
- Checked on every Get/Set/Incr/Decr/Del/Exists operation
- Performance overhead of time comparisons

**What to benchmark:**
- Impact of expired keys percentage on operation performance
- Background cleanup goroutine CPU usage
- Memory reclamation efficiency

**Why it matters:** The storage layer is the core of the system. All performance bottlenecks eventually trace back here.

---

### 3. Server Request Handling

**Location:** `internal/server/server.go`

#### Connection Handling

**What to benchmark:**
- Per-connection goroutine overhead
- Connection setup/teardown costs
- Request parsing → command execution → response serialization pipeline
- Error handling overhead
- Logger overhead (structured logging)

**Why it matters:** Understanding end-to-end request latency is critical for real-world performance.

---

### 4. Command Execution

#### Individual Commands

**Simple Commands:**
- `GET` - single key read
- `SET` - single key write with options

**Numeric Operations:**
- `INCR` - increment by 1
- `DECR` - decrement by 1
- `DECRBY` - decrement by N

**Multi-key Operations:**
- `DEL` - delete multiple keys
- `EXISTS` - check existence of multiple keys

**Metadata Operations:**
- `TTL` - time to live in seconds
- `PTTL` - time to live in milliseconds

#### Command Registry

**Location:** `internal/command/registry/registry.go`

**What to benchmark:**
- Command lookup overhead
- String case conversion (`strings.ToUpper`)
- Map lookup performance

**Why it matters:** Every command goes through the registry lookup.

---

### 5. Performance Scenarios

#### Workload Patterns

1. **Read-Heavy Workload**
   - 90% GET, 10% SET
   - Simulates cache-like usage
   - Tests read lock scaling

2. **Write-Heavy Workload**
   - 10% GET, 90% SET
   - Tests write lock contention
   - GC pressure from allocations

3. **Balanced Workload**
   - 50% GET, 50% SET
   - Realistic mixed usage
   - Tests lock fairness

4. **Increment-Heavy Workload**
   - 80% INCR/DECR, 20% GET
   - Counter/metrics use case
   - Tests numeric conversion overhead

5. **Multi-Key Operations**
   - DEL with 1, 10, 100, 1000 keys
   - EXISTS with 1, 10, 100, 1000 keys
   - Tests lock hold time impact

#### Concurrency Levels

Test with varying numbers of concurrent clients:
- **1 client** - baseline performance
- **10 clients** - low concurrency
- **50 clients** - moderate concurrency
- **100 clients** - high concurrency
- **500 clients** - very high concurrency
- **1000 clients** - stress test

**Metrics to track:**
- Throughput (ops/sec) vs concurrency
- Latency (p50, p95, p99) vs concurrency
- Point where performance degrades
- CPU and memory usage at each level

#### Data Characteristics

**Value Sizes:**
- **Small** - 10 bytes (IDs, flags)
- **Medium** - 1KB (JSON objects)
- **Large** - 10KB (HTML fragments)
- **Very Large** - 100KB (documents)

**Key Distribution:**
- **Uniform** - all keys equally likely
- **Zipfian** - realistic hot-key distribution
- **Sequential** - cache-friendly access pattern

**Key Count:**
- 1K, 10K, 100K, 1M keys in store

#### Expiry Scenarios

1. **No Expiry (Baseline)**
   - Pure performance without expiry overhead

2. **Short TTL**
   - 1-10 seconds
   - High churn rate
   - Tests cleanup efficiency

3. **Long TTL**
   - Hours to days
   - Low churn rate
   - Steady-state performance

4. **Mixed TTL**
   - 25% no expiry, 50% short, 25% long
   - Realistic usage pattern

5. **High Expired Keys Percentage**
   - 50%+ keys expired
   - Tests lazy expiry overhead
   - Background cleanup impact

---

### 6. Memory and Resource Utilization

#### Memory Metrics

**What to benchmark:**
- Bytes allocated per operation
- Heap allocations per operation
- Total heap size vs number of keys
- Memory growth rate over time
- Memory reclamation after deletions

**Tools:**
- `go test -bench=. -benchmem`
- `runtime.MemStats`
- `pprof` heap profiling

#### Garbage Collection

**What to monitor:**
- GC pause times (p50, p95, p99)
- GC frequency
- GC CPU percentage
- Impact on tail latencies

**Tools:**
- `GODEBUG=gctrace=1`
- `runtime.ReadMemStats()`

#### Goroutine and CPU

**What to track:**
- Number of goroutines under load
- CPU utilization per core
- Context switch rates
- Goroutine creation/destruction overhead

---

### 7. Comparison Benchmarks

#### Compare Against:

**Redis (Official)**
- Use `redis-benchmark` tool
- Validates protocol compatibility
- Sets performance targets
- Identifies optimization opportunities

**Pure Go Implementation**
- Simple `map[string][]byte` with `sync.RWMutex`
- Measures protocol overhead
- Quantifies feature costs (expiry, options)

**Different Storage Backends** (future)
- If you add persistence, disk-backed storage, etc.
- Compare in-memory vs persistent performance

---

## Recommended Benchmark Tools

### 1. Go Built-in Benchmarking

```go
func BenchmarkGet(b *testing.B) {
    store := memory.NewKVMemoryStore()
    ctx := context.Background()
    store.Set(ctx, "key", []byte("value"), nil)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        store.Get(ctx, "key")
    }
}
```

**Usage:**
```bash
go test -bench=. -benchmem -benchtime=10s
go test -bench=BenchmarkGet -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

### 2. redis-benchmark (Official Redis Tool)

```bash
redis-benchmark -h localhost -p 6379 -t set,get -n 100000 -c 50
redis-benchmark -h localhost -p 6379 -t incr,decr -n 100000 -c 50
redis-benchmark -h localhost -p 6379 -q  # Quick mode
```

**Metrics provided:**
- Requests per second
- Latency distribution
- Percentiles (p50, p95, p99)

### 3. Custom Load Generator

Build a custom tool for specific workload patterns:
```go
// Example: Mixed workload generator
func runMixedWorkload(ratio float64, duration time.Duration) {
    // ratio = read percentage
}
```

### 4. Profiling Tools

**CPU Profiling:**
```bash
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

**Memory Profiling:**
```bash
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

**Trace Analysis:**
```bash
go test -trace=trace.out -bench=.
go tool trace trace.out
```

### 5. Load Testing Tools

**go-wrk** - HTTP-like load testing
```bash
go-wrk -c 100 -d 30s http://localhost:6379
```

**vegeta** - HTTP load testing and benchmarking
```bash
echo "GET http://localhost:6379" | vegeta attack -duration=30s -rate=1000 | vegeta report
```

---

## Priority Benchmarking Order

### Phase 1: Micro-benchmarks (Baseline)
1. **Storage layer operations** - most critical bottleneck
2. **Protocol parsing/serialization** - every request/response
3. **Command execution** - individual command performance

### Phase 2: Integration (Real-world)
4. **End-to-end server performance** - full request pipeline
5. **Concurrency scaling** - production readiness
6. **Workload patterns** - specific use cases

### Phase 3: Analysis (Optimization)
7. **Memory and GC behavior** - stability and efficiency
8. **Profiling and hotspot identification** - targeted optimization
9. **Comparison benchmarks** - validate against Redis

---

## Key Metrics to Track

### Throughput
- **Operations per second (ops/sec)**
- Sustained throughput over time
- Throughput vs concurrency curve
- Throughput vs data size

### Latency
- **p50 (median)** - typical case
- **p95** - most users' experience
- **p99** - tail latency
- **p999** - worst case scenarios
- **max** - absolute worst case

### Memory
- **Bytes per operation**
- **Allocations per operation**
- **Total heap usage**
- **GC pause times**
- **Memory growth rate**

### CPU
- **CPU utilization percentage**
- **Per-core utilization**
- **Time spent in GC**
- **Lock contention time**

### Scalability
- **Performance vs concurrency** - does it scale?
- **Performance vs data size** - algorithmic complexity
- **Performance degradation point** - where does it break?

### Lock Contention
- **Time waiting for locks**
- **Lock hold duration**
- **Read vs write lock ratio**

---

## Benchmark Results Storage

Store benchmark results in:
```
benchmarks/
├── redis_benchmark/
│   └── results/
│       ├── baseline_YYYYMMDD.txt
│       ├── optimized_v1_YYYYMMDD.txt
│       └── comparison_redis_YYYYMMDD.txt
├── go_benchmarks/
│   └── results/
│       ├── storage_YYYYMMDD.txt
│       ├── protocol_YYYYMMDD.txt
│       └── server_YYYYMMDD.txt
└── reports/
    └── analysis_YYYYMMDD.md
```

---

## Example Benchmark Commands

### Storage Layer
```bash
cd internal/storage/kv/memory
go test -bench=BenchmarkGet -benchmem -benchtime=10s
go test -bench=BenchmarkSet -benchmem -benchtime=10s
go test -bench=BenchmarkIncr -benchmem -benchtime=10s
go test -bench=. -benchmem -cpuprofile=cpu.prof
```

### Protocol Layer
```bash
cd internal/protocol/resp
go test -bench=BenchmarkParse -benchmem -benchtime=10s
go test -bench=BenchmarkSerialize -benchmem -benchtime=10s
```

### End-to-End
```bash
# Start server
go run cmd/server/main.go -port 6379

# In another terminal
redis-benchmark -h localhost -p 6379 -t set,get -n 1000000 -c 50 -d 100
redis-benchmark -h localhost -p 6379 -t incr,decr -n 1000000 -c 50
redis-benchmark -h localhost -p 6379 -q
```

### Concurrency Scaling
```bash
for c in 1 10 50 100 500 1000; do
    echo "Testing with $c clients"
    redis-benchmark -h localhost -p 6379 -t get,set -n 100000 -c $c -q
done
```

---

## Performance Goals

### Target Metrics (Baseline)
- **Throughput:** > 100K ops/sec (single command)
- **Latency p99:** < 1ms (under moderate load)
- **Memory:** < 100 bytes per key-value pair overhead
- **Concurrency:** Linear scaling up to 100 concurrent clients
- **GC pause:** < 1ms p99

### Comparison to Redis
- **Aim for:** 50-80% of Redis performance
- **Acceptable:** 30-50% of Redis performance
- **Investigate if:** < 30% of Redis performance

---

## Next Steps

1. **Create micro-benchmarks** for storage operations
2. **Set up redis-benchmark** test suite with various scenarios
3. **Establish baseline metrics** before any optimizations
4. **Profile and identify hotspots** using pprof
5. **Optimize critical paths** based on profiling data
6. **Re-benchmark** to validate improvements
7. **Document results** and optimization decisions

---

## Notes

- Always benchmark on production-like hardware
- Run benchmarks multiple times and report averages
- Warm up the system before measurements
- Monitor system resources during benchmarks
- Keep benchmark history for regression detection
- Document system configuration (OS, Go version, CPU, RAM)
