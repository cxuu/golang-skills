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

type missingDoc struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type parseError struct {
	File    string `json:"file"`
	Message string `json:"message"`
}

type parsedFile struct {
	path string
	file *ast.File
}

type packageInfo struct {
	name      string
	firstFile string
	firstLine int
	hasDoc    bool
}

func usage() {
	fmt.Fprintf(os.Stdout, `check-docs.sh v%s - Check for missing doc comments on exported Go symbols

USAGE
    bash check-docs.sh [options] [path]

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --strict         Also check unexported types/functions
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
		fmt.Printf("check-docs.sh v%s\n", version)
		return
	}

	files, err := findGoFiles(opts.target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
	if len(files) == 0 {
		if opts.jsonOutput {
			fmt.Println(`{"missing":[],"total":0,"truncated":false,"status":"no_go_files"}`)
		} else {
			fmt.Printf("No Go files found in: %s\n", opts.target)
		}
		return
	}

	fset := token.NewFileSet()
	parsed := make([]parsedFile, 0, len(files))
	parseErrors := []parseError{}
	for _, path := range files {
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			parseErrors = append(parseErrors, parseError{File: path, Message: err.Error()})
			continue
		}
		parsed = append(parsed, parsedFile{path: path, file: file})
	}

	packages := map[string]*packageInfo{}
	for _, pf := range parsed {
		key := packageKey(pf.path, pf.file.Name.Name)
		line := fset.Position(pf.file.Package).Line
		info, ok := packages[key]
		if !ok {
			info = &packageInfo{name: pf.file.Name.Name, firstFile: pf.path, firstLine: line}
			packages[key] = info
		}
		if pf.path < info.firstFile {
			info.firstFile = pf.path
			info.firstLine = line
		}
		if pf.file.Doc != nil {
			info.hasDoc = true
		}
	}

	missing := []missingDoc{}
	reportedPackage := map[string]bool{}
	for _, pf := range parsed {
		key := packageKey(pf.path, pf.file.Name.Name)
		info := packages[key]
		if !info.hasDoc && !reportedPackage[key] {
			missing = append(missing, missingDoc{
				File: info.firstFile,
				Line: info.firstLine,
				Kind: "package",
				Name: info.name,
			})
			reportedPackage[key] = true
		}
		missing = append(missing, findMissingDeclDocs(fset, pf, opts.strict)...)
	}

	sort.Slice(missing, func(i, j int) bool {
		if missing[i].File == missing[j].File {
			return missing[i].Line < missing[j].Line
		}
		return missing[i].File < missing[j].File
	})

	total := len(missing)
	truncated := false
	if opts.limit > 0 && total > opts.limit {
		missing = missing[:opts.limit]
		truncated = true
	}

	if opts.jsonOutput {
		out := struct {
			Missing     []missingDoc `json:"missing"`
			Total       int          `json:"total"`
			Truncated   bool         `json:"truncated"`
			Status      string       `json:"status,omitempty"`
			ParseErrors []parseError `json:"parse_errors,omitempty"`
		}{Missing: missing, Total: total, Truncated: truncated}
		if len(parseErrors) > 0 {
			out.Status = "parse_error"
			out.ParseErrors = parseErrors
		}
		data, err := json.Marshal(out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: marshal JSON: %v\n", err)
			os.Exit(2)
		}
		fmt.Println(string(data))
	} else {
		if len(parseErrors) > 0 {
			fmt.Println("Malformed Go files:")
			fmt.Println()
			for _, item := range parseErrors {
				fmt.Printf("  %s  %s\n", item.File, item.Message)
			}
			fmt.Println()
		}
		if total == 0 {
			if len(parseErrors) == 0 {
				fmt.Println("All exported symbols are documented.")
				return
			}
		} else {
			fmt.Println("Undocumented exported symbols:")
			fmt.Println()
			for _, item := range missing {
				fmt.Printf("  %s:%d  [%s] %s\n", item.File, item.Line, item.Kind, item.Name)
			}
			if truncated {
				fmt.Printf("  ... and %d more (use --limit to adjust)\n", total-opts.limit)
			}
			fmt.Println()
			fmt.Printf("Total: %d undocumented symbol(s)\n", total)
		}
	}

	if len(parseErrors) > 0 {
		os.Exit(2)
	}
	if total > 0 {
		os.Exit(1)
	}
}

type options struct {
	jsonOutput bool
	strict     bool
	limit      int
	target     string
	help       bool
	version    bool
}

func parseArgs(args []string) (options, error) {
	opts := options{target: "./..."}
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
		case arg == "--strict":
			opts.strict = true
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

func packageKey(path, name string) string {
	return filepath.Dir(path) + "|" + name
}

func findMissingDeclDocs(fset *token.FileSet, pf parsedFile, strict bool) []missingDoc {
	var missing []missingDoc
	for _, decl := range pf.file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if ast.IsExported(d.Name.Name) || strict {
				if d.Doc == nil {
					kind := "function"
					if d.Recv != nil {
						kind = "method"
					}
					missing = append(missing, missingDoc{
						File: pf.path,
						Line: fset.Position(d.Pos()).Line,
						Kind: kind,
						Name: d.Name.Name,
					})
				}
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if ast.IsExported(s.Name.Name) || strict {
						if s.Doc == nil && d.Doc == nil {
							missing = append(missing, missingDoc{
								File: pf.path,
								Line: fset.Position(s.Pos()).Line,
								Kind: "type",
								Name: s.Name.Name,
							})
						}
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if !ast.IsExported(name.Name) && !strict {
							continue
						}
						if s.Doc == nil && d.Doc == nil {
							missing = append(missing, missingDoc{
								File: pf.path,
								Line: fset.Position(name.Pos()).Line,
								Kind: strings.ToLower(d.Tok.String()),
								Name: name.Name,
							})
						}
					}
				}
			}
		}
	}
	return missing
}
