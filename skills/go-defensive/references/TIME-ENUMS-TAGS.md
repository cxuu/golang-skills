# Time, Struct Tags, and Embedding Patterns

## Use time.Time and time.Duration

Always use the `time` package. Avoid raw `int` for time values.

### Instants

**Bad**
```go
func isActive(now, start, stop int) bool {
  return start <= now && now < stop
}
```

**Good**
```go
func isActive(now, start, stop time.Time) bool {
  return (start.Before(now) || start.Equal(now)) && now.Before(stop)
}
```

### Durations

**Bad**
```go
func poll(delay int) {
  time.Sleep(time.Duration(delay) * time.Millisecond)
}
poll(10)  // seconds? milliseconds?
```

**Good**
```go
func poll(delay time.Duration) {
  time.Sleep(delay)
}
poll(10 * time.Second)
```

### JSON Fields

When `time.Duration` isn't possible, include unit in field name:

**Bad**
```go
type Config struct {
  Interval int `json:"interval"`
}
```

**Good**
```go
type Config struct {
  IntervalMillis int `json:"intervalMillis"`
}
```

## Avoid Embedding Types in Public Structs

Embedded types leak implementation details and inhibit type evolution.

**Bad**
```go
type ConcreteList struct {
  *AbstractList
}
```

**Good**
```go
type ConcreteList struct {
  list *AbstractList
}

func (l *ConcreteList) Add(e Entity) {
  l.list.Add(e)
}

func (l *ConcreteList) Remove(e Entity) {
  l.list.Remove(e)
}
```

Embedding problems:
- Adding methods to embedded interface is a breaking change
- Removing methods from embedded struct is a breaking change
- Replacing the embedded type is a breaking change

## Use Field Tags in Marshaled Structs

Always use explicit field tags for JSON, YAML, etc.

**Bad**
```go
type Stock struct {
  Price int
  Name  string
}
```

**Good**
```go
type Stock struct {
  Price int    `json:"price"`
  Name  string `json:"name"`
  // Safe to rename Name to Symbol
}
```

Tags make the serialization contract explicit and safe to refactor.
