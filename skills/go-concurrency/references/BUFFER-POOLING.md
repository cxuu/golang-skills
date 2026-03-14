# Buffer Pooling with Channels

Use a buffered channel as a free list to reuse allocated buffers, avoiding
repeated allocations. This "leaky buffer" pattern uses `select` with `default`
for non-blocking operations.

> **Source**: Effective Go

```go
var freeList = make(chan *Buffer, 100) // Buffered channel as free list

// Client: Get buffer from free list or allocate new one
func getBuffer() *Buffer {
    select {
    case b := <-freeList:
        return b // Reuse existing buffer
    default:
        return new(Buffer) // Free list empty; allocate new buffer
    }
}

// Server: Return buffer to free list if room, otherwise drop it
func putBuffer(b *Buffer) {
    b.Reset() // Prepare for reuse
    select {
    case freeList <- b:
        // Buffer returned to free list
    default:
        // Free list full; drop buffer (GC will reclaim)
    }
}
```

## How It Works

1. **Non-blocking receive**: Client tries to grab a buffer from `freeList`. If
   empty, `default` runs and allocates a new buffer.
2. **Non-blocking send**: Server tries to return the buffer. If `freeList` is
   full, `default` runs and the buffer is dropped for garbage collection.
3. **Bounded memory**: The channel capacity (100) limits pooled buffers,
   preventing unbounded growth.

This pattern is useful when allocation is expensive and buffer reuse is
beneficial, but you don't want blocking behavior when the pool is empty or full.

## When to Use

- High-frequency allocations of similar-sized objects
- Performance-critical code paths where allocation overhead matters
- Scenarios where you want bounded memory usage

## Production Alternative

For production code, consider `sync.Pool` which provides similar functionality
with better integration into the garbage collector:

```go
var bufferPool = sync.Pool{
    New: func() any {
        return new(Buffer)
    },
}

func getBuffer() *Buffer {
    return bufferPool.Get().(*Buffer)
}

func putBuffer(b *Buffer) {
    b.Reset()
    bufferPool.Put(b)
}
```

`sync.Pool` advantages:
- Automatic cleanup during garbage collection
- No need to manage pool size
- Thread-safe by design
- Better performance under high concurrency

The channel-based approach is still valuable for understanding Go's concurrency
primitives and for cases where you need more control over pool behavior.
