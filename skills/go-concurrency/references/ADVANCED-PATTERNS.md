# Advanced Concurrency Patterns

Detailed reference for advanced concurrency patterns from Effective Go. These
patterns are situational — use when you need request/response multiplexing or
CPU-bound parallelization.

---

## Channels of Channels

> **Source**: Effective Go

A channel is a first-class value that can be allocated and passed around like
any other. A powerful pattern is embedding a **reply channel** inside a request
struct, letting each client provide its own path for the answer:

```go
type Request struct {
    args       []int
    f          func([]int) int
    resultChan chan int
}
```

The client sends a request with a function, its arguments, and a channel on
which to receive the result:

```go
request := &Request{[]int{3, 4, 5}, sum, make(chan int)}
clientRequests <- request
fmt.Printf("answer: %d\n", <-request.resultChan)
```

The server handler reads from the queue and sends results back on each
request's reply channel:

```go
func handle(queue chan *Request) {
    for req := range queue {
        req.resultChan <- req.f(req.args)
    }
}
```

This pattern forms the basis for a rate-limited, parallel, non-blocking RPC
system without a mutex in sight.

---

## CPU-Bound Parallelization

> **Source**: Effective Go

When a computation can be broken into independent pieces, parallelize it across
CPU cores using a channel to signal completion:

```go
type Vector []float64

func (v Vector) DoSome(i, n int, u Vector, c chan int) {
    for ; i < n; i++ {
        v[i] += u.Op(v[i])
    }
    c <- 1
}

func (v Vector) DoAll(u Vector) {
    c := make(chan int, runtime.NumCPU())
    for i := 0; i < runtime.NumCPU(); i++ {
        go v.DoSome(i*len(v)/runtime.NumCPU(), (i+1)*len(v)/runtime.NumCPU(), u, c)
    }
    for i := 0; i < runtime.NumCPU(); i++ {
        <-c
    }
}
```

Use `runtime.NumCPU()` for hardware cores or `runtime.GOMAXPROCS(0)` to honor
the user's resource configuration.

> **Important**: Don't confuse concurrency (structuring a program as
> independently executing components) with parallelism (executing calculations
> simultaneously on multiple CPUs). Go is a concurrent language; not all
> parallelization problems fit its model.
