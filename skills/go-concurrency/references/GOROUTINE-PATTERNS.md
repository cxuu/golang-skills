# Goroutine Lifecycle Patterns

Detailed patterns for managing goroutine lifetimes — ensuring every goroutine
has a clear start/stop mechanism and preventing resource leaks.

---

## Making Lifetimes Obvious

Keep synchronization-related code constrained within function scope and factor
logic into synchronous functions.

```go
// Good: Goroutine lifetimes are clear
func (w *Worker) Run(ctx context.Context) error {
    var wg sync.WaitGroup
    for item := range w.q {
        wg.Add(1)
        go func() {
            defer wg.Done()
            process(ctx, item) // Returns when context is cancelled
        }()
    }
    wg.Wait() // Prevent spawned goroutines from outliving this function
}
```

```go
// Bad: Careless about when goroutines finish
func (w *Worker) Run() {
    for item := range w.q {
        go process(item) // When does this finish? What if it never does?
    }
}
```

---

## Stop/Done Channel Pattern

Every goroutine must have a predictable stop mechanism. Use a stop channel to
signal shutdown and a done channel to confirm exit:

```go
var (
    stop = make(chan struct{}) // tells the goroutine to stop
    done = make(chan struct{}) // tells us that the goroutine exited
)
go func() {
    defer close(done)
    ticker := time.NewTicker(delay)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            flush()
        case <-stop:
            return
        }
    }
}()

// To shut down:
close(stop)  // signal the goroutine to stop
<-done       // and wait for it to exit
```

Sending on a closed channel panics — always use `close()` to signal, never send:

```go
ch := make(chan int)
close(ch)
ch <- 13 // panic: send on closed channel
```

---

## Waiting for Goroutines

Use `sync.WaitGroup` for multiple goroutines:

```go
var wg sync.WaitGroup
for i := 0; i < N; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // work...
    }()
}
wg.Wait()
```

Use a done channel for a single goroutine:

```go
done := make(chan struct{})
go func() {
    defer close(done)
    // work...
}()
<-done // wait for goroutine to finish
```

---

## No Goroutines in init()

`init()` functions should not spawn goroutines. Expose an object that manages
the goroutine's lifetime with a method (`Close`, `Stop`, `Shutdown`):

```go
// Bad: Spawns uncontrollable background goroutine
func init() {
    go doWork()
}
```

```go
// Good: Explicit lifecycle management
type Worker struct {
    stop chan struct{}
    done chan struct{}
}

func NewWorker() *Worker {
    w := &Worker{
        stop: make(chan struct{}),
        done: make(chan struct{}),
    }
    go w.doWork()
    return w
}

func (w *Worker) Shutdown() {
    close(w.stop)
    <-w.done
}
```

---

## Prefer Synchronous Functions

Synchronous functions let the caller control concurrency:

```go
// Good: Synchronous function - caller controls concurrency
func ProcessItems(items []Item) ([]Result, error) {
    var results []Result
    for _, item := range items {
        result, err := processItem(item)
        if err != nil {
            return nil, err
        }
        results = append(results, result)
    }
    return results, nil
}

// Caller can add concurrency if needed:
go func() {
    results, err := ProcessItems(items)
    // handle results
}()
```

It is quite difficult (sometimes impossible) to remove unnecessary concurrency
at the caller side. Let the caller add concurrency when needed.
