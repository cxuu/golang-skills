# Goroutine Lifecycle Patterns

Detailed patterns for managing goroutine lifetimes — ensuring every goroutine
has a clear start/stop mechanism and preventing resource leaks.

---

## Making Lifetimes Obvious

> The WaitGroup example and scoping rules are in the parent skill (SKILL.md §
> Goroutine Lifetimes, Core Rules). This reference covers: stop/done channel
> patterns, waiting strategies, init() lifecycle examples, and synchronous API
> design.

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

> The `sync.WaitGroup` pattern for multiple goroutines is in the parent skill
> (SKILL.md § Goroutine Lifetimes). Below is the done-channel alternative for a
> single goroutine.

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

> The core rule is in the parent skill (SKILL.md § Core Rules, rule 3). Below
> are expanded examples showing lifecycle management.

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

> The rationale and benefit table are in the parent skill (SKILL.md §
> Synchronous Functions). Below is a concrete code example.

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
