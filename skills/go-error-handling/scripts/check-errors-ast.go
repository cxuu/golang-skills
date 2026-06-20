package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const version = "1.1.0"

type finding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type options struct {
	jsonOutput      bool
	checkBareReturn bool
	limit           int
	target          string
	help            bool
	version         bool
}

func usage() {
	fmt.Fprintf(os.Stdout, `check-errors.sh v%s - Check Go code for common error handling anti-patterns

USAGE
    bash check-errors.sh [options] [path]

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --no-bare-return Skip the bare 'return err' check (high false-positive rate)
    --limit N        Show at most N results (default: all)
`, version)
}

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	if opts.help {
		usage()
		return
	}
	if opts.version {
		fmt.Printf("check-errors.sh v%s\n", version)
		return
	}

	files, err := findGoFiles(opts.target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
	if len(files) == 0 {
		if opts.jsonOutput {
			fmt.Println(`{"findings":[],"total":0,"truncated":false,"status":"no_go_files"}`)
		} else {
			fmt.Printf("No Go files found in: %s\n", opts.target)
		}
		return
	}

	findings := []finding{}
	for _, file := range files {
		fileFindings, err := analyzeFile(file, opts.checkBareReturn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: parse %s: %v\n", file, err)
			os.Exit(2)
		}
		findings = append(findings, fileFindings...)
	}

	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].File == findings[j].File {
			if findings[i].Line == findings[j].Line {
				return findings[i].Rule < findings[j].Rule
			}
			return findings[i].Line < findings[j].Line
		}
		return findings[i].File < findings[j].File
	})

	total := len(findings)
	truncated := false
	if opts.limit > 0 && total > opts.limit {
		findings = findings[:opts.limit]
		truncated = true
	}

	if opts.jsonOutput {
		out := struct {
			Findings  []finding `json:"findings"`
			Total     int       `json:"total"`
			Truncated bool      `json:"truncated"`
		}{Findings: findings, Total: total, Truncated: truncated}
		data, err := json.Marshal(out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: marshal JSON: %v\n", err)
			os.Exit(2)
		}
		fmt.Println(string(data))
	} else {
		if total == 0 {
			fmt.Println("No error handling anti-patterns found.")
			return
		}
		fmt.Println("Error handling anti-patterns found:")
		fmt.Println()
		for _, item := range findings {
			fmt.Printf("  %s:%d  [%s] %s\n", item.File, item.Line, item.Rule, item.Message)
		}
		if truncated {
			fmt.Printf("  ... and %d more (use --limit to adjust)\n", total-opts.limit)
		}
		fmt.Println()
		fmt.Printf("Total: %d finding(s)\n", total)
	}

	if total > 0 {
		os.Exit(1)
	}
}

func analyzeFile(path string, checkBareReturn bool) ([]finding, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	var findings []finding
	var logLines []int
	var errReturnLines []int

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.BinaryExpr:
			if node.Op != token.EQL && node.Op != token.NEQ {
				return true
			}
			line := fset.Position(node.Pos()).Line
			switch {
			case isErrorCall(node.X) && isStringLiteral(node.Y):
				findings = append(findings, finding{
					File:    path,
					Line:    line,
					Rule:    "string-error-compare",
					Message: "comparing err.Error() to string; use errors.Is() or errors.As() instead",
				})
			case isStringLiteral(node.X) && isErrorCall(node.Y):
				findings = append(findings, finding{
					File:    path,
					Line:    line,
					Rule:    "string-error-compare",
					Message: "comparing string to err.Error(); use errors.Is() or errors.As() instead",
				})
			}
		case *ast.CallExpr:
			line := fset.Position(node.Pos()).Line
			if isStringsContainsErrorCall(node) {
				findings = append(findings, finding{
					File:    path,
					Line:    line,
					Rule:    "string-error-compare",
					Message: "using strings.Contains on err.Error(); use errors.Is() or errors.As() instead",
				})
			}
			if isLogCallWithErr(node) {
				logLines = append(logLines, line)
			}
		case *ast.ReturnStmt:
			if returnsErr(node) {
				line := fset.Position(node.Return).Line
				errReturnLines = append(errReturnLines, line)
				if checkBareReturn {
					findings = append(findings, finding{
						File:    path,
						Line:    line,
						Rule:    "bare-return-err",
						Message: "returning err without wrapping context; consider fmt.Errorf('...: %w', err)",
					})
				}
			}
		}
		return true
	})

	for _, retLine := range errReturnLines {
		for i := len(logLines) - 1; i >= 0; i-- {
			logLine := logLines[i]
			if logLine >= retLine {
				continue
			}
			if retLine-logLine > 5 {
				break
			}
			findings = append(findings, finding{
				File:    path,
				Line:    logLine,
				Rule:    "log-and-return",
				Message: fmt.Sprintf("error is both logged (line %d) and returned (line %d); handle errors once", logLine, retLine),
			})
			break
		}
	}

	return findings, nil
}

func isErrorCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "Error"
}

func isStringLiteral(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	return ok && lit.Kind == token.STRING
}

func isStringsContainsErrorCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Contains" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "strings" || len(call.Args) == 0 {
		return false
	}
	return containsErrorCall(call.Args[0])
}

func containsErrorCall(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		if e, ok := n.(ast.Expr); ok && isErrorCall(e) {
			found = true
			return false
		}
		return true
	})
	return found
}

func isLogCallWithErr(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	switch x.Name {
	case "log", "logger", "slog":
	default:
		return false
	}
	for _, arg := range call.Args {
		if containsIdent(arg, "err") {
			return true
		}
	}
	return false
}

func containsIdent(expr ast.Expr, name string) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		ident, ok := n.(*ast.Ident)
		if ok && ident.Name == name {
			found = true
			return false
		}
		return true
	})
	return found
}

func returnsErr(stmt *ast.ReturnStmt) bool {
	if len(stmt.Results) == 0 {
		return false
	}
	ident, ok := stmt.Results[len(stmt.Results)-1].(*ast.Ident)
	return ok && ident.Name == "err"
}

func findGoFiles(target string) ([]string, error) {
	info, err := os.Stat(target)
	if err == nil {
		if !info.IsDir() {
			if strings.HasSuffix(target, ".go") && !strings.HasSuffix(target, "_test.go") {
				return []string{target}, nil
			}
			return nil, nil
		}
		return walkGoFiles(target)
	}

	if strings.HasSuffix(target, "/...") {
		dir := strings.TrimSuffix(target, "/...")
		if dir == "" {
			dir = "."
		}
		if info, statErr := os.Stat(dir); statErr == nil && info.IsDir() {
			return walkGoFiles(dir)
		}
	}

	return nil, fmt.Errorf("path not found: %s", target)
}

func walkGoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func parseArgs(args []string) (options, error) {
	opts := options{target: ".", checkBareReturn: true}
	var positionals []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			opts.help = true
		case arg == "-v" || arg == "--version":
			opts.version = true
		case arg == "--json":
			opts.jsonOutput = true
		case arg == "--no-bare-return":
			opts.checkBareReturn = false
		case arg == "--limit":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--limit requires a number")
			}
			i++
			limit, err := parseNonNegativeInt(args[i])
			if err != nil {
				return opts, fmt.Errorf("--limit must be a non-negative integer, got: %s", args[i])
			}
			opts.limit = limit
		case strings.HasPrefix(arg, "--limit="):
			value := strings.TrimPrefix(arg, "--limit=")
			limit, err := parseNonNegativeInt(value)
			if err != nil {
				return opts, fmt.Errorf("--limit must be a non-negative integer, got: %s", value)
			}
			opts.limit = limit
		case strings.HasPrefix(arg, "-"):
			return opts, fmt.Errorf("unknown option: %s", arg)
		default:
			positionals = append(positionals, arg)
		}
	}
	if len(positionals) > 1 {
		return opts, fmt.Errorf("expected at most one path")
	}
	if len(positionals) == 1 {
		opts.target = positionals[0]
	}
	return opts, nil
}

func parseNonNegativeInt(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid")
		}
		n = n*10 + int(r-'0')
	}
	return n, nil
}
