# Copying Slices and Maps at API Boundaries

> **Source**: Uber Style Guide

Slices and maps contain references to their underlying data. Copy them at API
boundaries to prevent callers from mutating internal state (or vice versa).

## Receiving Slices and Maps

When a function stores a slice or map passed by the caller, always make a
defensive copy. The caller retains the original reference and can modify it
after your function returns.

### Slices

**Bad**
```go
func (d *Driver) SetTrips(trips []Trip) {
  d.trips = trips  // caller can still modify d.trips
}
```

**Good**
```go
func (d *Driver) SetTrips(trips []Trip) {
  d.trips = make([]Trip, len(trips))
  copy(d.trips, trips)
}
```

### Maps

**Bad**
```go
func (s *Server) SetConfig(cfg map[string]string) {
  s.config = cfg  // caller can still modify s.config
}
```

**Good**
```go
func (s *Server) SetConfig(cfg map[string]string) {
  s.config = make(map[string]string, len(cfg))
  for k, v := range cfg {
    s.config[k] = v
  }
}
```

## Returning Slices and Maps

When returning internal slices or maps, return a copy to prevent callers from
modifying your internal state.

### Returning a Map

**Bad**
```go
func (s *Stats) Snapshot() map[string]int {
  s.mu.Lock()
  defer s.mu.Unlock()
  return s.counters  // exposes internal state!
}
```

**Good**
```go
func (s *Stats) Snapshot() map[string]int {
  s.mu.Lock()
  defer s.mu.Unlock()
  result := make(map[string]int, len(s.counters))
  for k, v := range s.counters {
    result[k] = v
  }
  return result
}
```

### Returning a Slice

**Bad**
```go
func (q *Queue) Items() []Item {
  return q.items  // caller can append, modify, or reslice
}
```

**Good**
```go
func (q *Queue) Items() []Item {
  result := make([]Item, len(q.items))
  copy(result, q.items)
  return result
}
```

## When Copies Are Not Needed

Defensive copies have a cost. Skip them when:

- The data is **immutable by convention** and clearly documented
- The slice/map is **created fresh** for the caller (not stored internally)
- Performance profiling shows the copy is a bottleneck in a hot path

When in doubt, copy. The cost is usually negligible compared to the bugs that
shared references cause.
