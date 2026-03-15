# Table-Driven Tests, Subtests, and Parallel Tests

Detailed reference for structuring table-driven tests and subtests in Go.
Sources: Google Go Style Guide, Uber Go Style Guide.

---

## Basic Structure

```go
func TestCompare(t *testing.T) {
    tests := []struct {
        a, b string
        want int
    }{
        {"", "", 0},
        {"a", "", 1},
        {"", "a", -1},
        {"abc", "abc", 0},
    }
    for _, tt := range tests {
        got := Compare(tt.a, tt.b)
        if got != tt.want {
            t.Errorf("Compare(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
        }
    }
}
```

---

## Best Practices

**Use field names** when test cases span many lines or have adjacent fields of
the same type:

```go
tests := []struct {
    name  string
    input string
    want  int
}{
    {name: "empty", input: "", want: 0},
    {name: "single", input: "a", want: 1},
}
```

**Don't identify rows by index** — include inputs in failure messages instead of
`Case #%d failed`.

---

## Avoid Complexity in Table Tests

When test cases need complex setup, conditional mocking, or multiple branches,
prefer separate test functions over table tests.

```go
// Bad: Too many conditional fields make tests hard to understand
tests := []struct {
    give          string
    want          string
    wantErr       error
    shouldCallX   bool
    shouldCallY   bool
    giveXResponse string
    giveXErr      error
    giveYResponse string
    giveYErr      error
}{...}

for _, tt := range tests {
    t.Run(tt.give, func(t *testing.T) {
        if tt.shouldCallX {
            xMock.EXPECT().Call().Return(tt.giveXResponse, tt.giveXErr)
        }
        if tt.shouldCallY {
            yMock.EXPECT().Call().Return(tt.giveYResponse, tt.giveYErr)
        }
        // ...
    })
}

// Good: Separate focused tests are clearer
func TestShouldCallX(t *testing.T) {
    xMock.EXPECT().Call().Return("XResponse", nil)
    got, err := DoComplexThing("inputX", xMock, yMock)
    // assert...
}

func TestShouldCallYAndFail(t *testing.T) {
    yMock.EXPECT().Call().Return("YResponse", nil)
    _, err := DoComplexThing("inputY", xMock, yMock)
    // assert error...
}
```

**Table tests work best when:**

- All cases run identical logic (no conditional assertions)
- Setup is the same for all cases
- No conditional mocking based on test case fields
- All table fields are used in all tests

A single `shouldErr` field for success/failure is acceptable if the test body is
short and straightforward.

---

## Subtests

Use `t.Run` for better organization, filtering, and parallel execution.

### Subtest Names

- Use clear, concise names: `t.Run("empty_input", ...)`, `t.Run("hu_to_en", ...)`
- Avoid wordy descriptions or slashes (slashes break test filtering)
- Subtests must be independent — no shared state or execution order dependencies

### Table Tests with Subtests

```go
func TestTranslate(t *testing.T) {
    tests := []struct {
        name, srcLang, dstLang, input, want string
    }{
        {"hu_en_basic", "hu", "en", "köszönöm", "thank you"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Translate(tt.srcLang, tt.dstLang, tt.input); got != tt.want {
                t.Errorf("Translate(%q, %q, %q) = %q, want %q",
                    tt.srcLang, tt.dstLang, tt.input, got, tt.want)
            }
        })
    }
}
```

---

## Parallel Tests

When using `t.Parallel()` in table tests, be aware of loop variable capture:

```go
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // Go 1.22+: tt is correctly captured per iteration
        // Go 1.21-: add "tt := tt" here to capture the variable
        got := Process(tt.give)
        if got != tt.want {
            t.Errorf("Process(%q) = %q, want %q", tt.give, got, tt.want)
        }
    })
}
```
