package evals_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse --show-toplevel: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func findAllScripts(t *testing.T) []string {
	t.Helper()
	root := repoRoot(t)
	matches, err := filepath.Glob(filepath.Join(root, "skills", "*", "scripts", "*.sh"))
	if err != nil {
		t.Fatalf("glob scripts: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("no scripts found")
	}
	return matches
}

func findSkillDirs(t *testing.T) []string {
	t.Helper()
	root := repoRoot(t)
	entries, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		t.Fatalf("read skills dir: %v", err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(root, "skills", e.Name()))
		}
	}
	if len(dirs) == 0 {
		t.Fatal("no skill directories found")
	}
	return dirs
}

func readLines(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	return lines
}

func runCommand(t *testing.T, wantExit int, name string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	gotExit := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			gotExit = exitErr.ExitCode()
		} else {
			t.Fatalf("%s %s failed to start: %v\n%s", name, strings.Join(args, " "), err, out)
		}
	}
	if gotExit != wantExit {
		t.Fatalf("%s %s exit = %d, want %d\n%s", name, strings.Join(args, " "), gotExit, wantExit, out)
	}
	return out
}

func runCommandStdout(t *testing.T, wantExit int, name string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	gotExit := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			gotExit = exitErr.ExitCode()
		} else {
			t.Fatalf("%s %s failed to start: %v\nstderr:\n%s", name, strings.Join(args, " "), err, stderr.String())
		}
	}
	if gotExit != wantExit {
		t.Fatalf("%s %s exit = %d, want %d\nstdout:\n%s\nstderr:\n%s", name, strings.Join(args, " "), gotExit, wantExit, stdout.String(), stderr.String())
	}
	return stdout.Bytes()
}

func runCommandInDir(t *testing.T, dir string, wantExit int, name string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	gotExit := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			gotExit = exitErr.ExitCode()
		} else {
			t.Fatalf("%s %s failed to start in %s: %v\n%s", name, strings.Join(args, " "), dir, err, out)
		}
	}
	if gotExit != wantExit {
		t.Fatalf("%s %s exit = %d, want %d in %s\n%s", name, strings.Join(args, " "), gotExit, wantExit, dir, out)
	}
	return out
}

func pathHasSuffix(got, wantSuffix string) bool {
	got = filepath.ToSlash(got)
	wantSuffix = filepath.ToSlash(wantSuffix)
	return got == wantSuffix || strings.HasSuffix(got, "/"+wantSuffix)
}

type jsonFinding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Rule    string `json:"rule"`
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

func requireFinding(t *testing.T, findings []jsonFinding, fileSuffix string, line int, ruleOrKind, messagePart string) {
	t.Helper()
	if messagePart == "" {
		t.Fatal("requireFinding messagePart must be non-empty; use a schema-specific helper for findings without messages")
	}
	for _, f := range findings {
		gotRule := f.Rule
		if gotRule == "" {
			gotRule = f.Kind
		}
		if pathHasSuffix(f.File, fileSuffix) && f.Line == line && gotRule == ruleOrKind && strings.Contains(f.Message, messagePart) {
			return
		}
	}
	t.Fatalf("missing finding %s:%d %s containing %q in %#v", fileSuffix, line, ruleOrKind, messagePart, findings)
}

func requireDocMissing(t *testing.T, missing []jsonFinding, fileSuffix string, line int, kind, name string) {
	t.Helper()
	for _, f := range missing {
		if pathHasSuffix(f.File, fileSuffix) && f.Line == line && f.Kind == kind && f.Name == name {
			return
		}
	}
	t.Fatalf("missing doc finding %s:%d %s %s in %#v", fileSuffix, line, kind, name, missing)
}

func requireInterfaceMissing(t *testing.T, missing []jsonFinding, fileSuffix string, line int, name string) {
	t.Helper()
	for _, f := range missing {
		if pathHasSuffix(f.File, fileSuffix) && f.Line == line && f.Name == name {
			return
		}
	}
	t.Fatalf("missing interface finding %s:%d %s in %#v", fileSuffix, line, name, missing)
}

func splitFrontmatter(content []byte) (fm, body string, ok bool) {
	s := string(content)
	lines := strings.Split(s, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", s, false
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return "", s, false
	}

	return strings.Join(lines[1:end], "\n"), strings.Join(lines[end+1:], "\n"), true
}

// parseFrontmatter extracts name, description, and body from SKILL.md content.
func parseFrontmatter(content []byte) (name, desc, body string) {
	fm, body, ok := splitFrontmatter(content)
	if !ok {
		return "", "", body
	}

	for _, line := range strings.Split(fm, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		}
		if strings.HasPrefix(line, "description:") {
			rest := strings.TrimPrefix(line, "description:")
			rest = strings.TrimSpace(rest)
			rest = strings.Trim(rest, `"'>`)
			desc = rest
		}
	}

	// Handle multi-line description (YAML folded/literal blocks)
	if desc == "" || desc == "|" || desc == ">" {
		inDesc := false
		var descLines []string
		for _, line := range strings.Split(fm, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "description:") {
				inDesc = true
				rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
				rest = strings.Trim(rest, `"'>|`)
				if rest != "" {
					descLines = append(descLines, rest)
				}
				continue
			}
			if inDesc {
				if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
					break
				}
				descLines = append(descLines, strings.TrimSpace(line))
			}
		}
		desc = strings.Join(descLines, " ")
	}

	return name, desc, body
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// TestPortability - scan scripts for macOS-incompatible patterns
// ---------------------------------------------------------------------------

func TestPortability(t *testing.T) {
	t.Parallel()
	scripts := findAllScripts(t)

	reDeclareA := regexp.MustCompile(`declare\s+-A`)
	reGNUSed := regexp.MustCompile(`sed\s+(-[^E\s]*\s+)?'s/[^']*\\[+]`) // sed without -E using \+

	t.Run("NoDeclareA", func(t *testing.T) {
		t.Parallel()
		for _, script := range scripts {
			lines := readLines(t, script)
			for i, line := range lines {
				if reDeclareA.MatchString(line) {
					t.Errorf("%s:%d: uses 'declare -A' (requires bash 4+, macOS ships 3.2)", filepath.Base(script), i+1)
				}
			}
		}
	})

	t.Run("NoGNUSedSyntax", func(t *testing.T) {
		t.Parallel()
		for _, script := range scripts {
			lines := readLines(t, script)
			for i, line := range lines {
				trimmed := strings.TrimSpace(line)
				if !strings.Contains(trimmed, "sed") {
					continue
				}
				if strings.Contains(trimmed, "sed -E") || strings.Contains(trimmed, "sed -i") {
					continue
				}
				if reGNUSed.MatchString(line) {
					t.Errorf("%s:%d: uses GNU sed \\+ syntax (fails on macOS BSD sed)", filepath.Base(script), i+1)
				}
			}
		}
	})

	t.Run("NoHardcodedGofmt", func(t *testing.T) {
		t.Parallel()
		reGofmtDot := regexp.MustCompile(`gofmt\s+-l\s+\.[)\s]|gofmt\s+-l\s+\.$`)
		for _, script := range scripts {
			lines := readLines(t, script)
			for i, line := range lines {
				if reGofmtDot.MatchString(line) {
					t.Errorf("%s:%d: hardcodes 'gofmt -l .' instead of using $TARGET", filepath.Base(script), i+1)
				}
			}
		}
	})
}

// ---------------------------------------------------------------------------
// TestScriptSmoke - verify all scripts accept --help and --version
// ---------------------------------------------------------------------------

func TestScriptSmoke(t *testing.T) {
	t.Parallel()
	scripts := findAllScripts(t)

	for _, script := range scripts {
		base := filepath.Base(script)
		script := script

		t.Run(base+"/help", func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command("bash", script, "--help")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("--help failed (exit %v):\n%s", err, out)
			}
		})

		t.Run(base+"/version", func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command("bash", script, "--version")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("--version failed (exit %v):\n%s", err, out)
			}
		})
	}
}

func TestScriptSyntax(t *testing.T) {
	t.Parallel()
	for _, script := range findAllScripts(t) {
		script := script
		t.Run(filepath.Base(script), func(t *testing.T) {
			t.Parallel()
			runCommand(t, 0, "bash", "-n", script)
			info, err := os.Stat(script)
			if err != nil {
				t.Fatalf("stat script: %v", err)
			}
			if info.Mode()&0111 == 0 {
				t.Fatalf("%s is not executable", script)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestScriptFunctional - run scripts against fixture files
// ---------------------------------------------------------------------------

func TestScriptFunctional(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	fixturesDir := filepath.Join(root, "evals", "fixtures")

	scriptPath := func(skill, name string) string {
		return filepath.Join(root, "skills", skill, "scripts", name)
	}

	t.Run("CheckNaming", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-naming", "check-naming.sh")
		fixture := filepath.Join(fixturesDir, "naming", "violations.go")
		out := runCommand(t, 1, "bash", script, "--json", fixture)

		if !json.Valid(out) {
			t.Fatalf("--json output is not valid JSON:\n%s", out)
		}

		var result struct {
			Violations []jsonFinding `json:"violations"`
			Total      int           `json:"total"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatalf("parse JSON: %v\n%s", err, out)
		}
		if result.Total != 5 {
			t.Fatalf("expected exactly 5 naming violations, got %d\n%s", result.Total, out)
		}
		requireFinding(t, result.Violations, "evals/fixtures/naming/violations.go", 1, "bad-package-name", "too generic")
		requireFinding(t, result.Violations, "evals/fixtures/naming/violations.go", 3, "screaming-const", "BAD_NAME")
		requireFinding(t, result.Violations, "evals/fixtures/naming/violations.go", 6, "screaming-const", "ALSO_BAD")
		requireFinding(t, result.Violations, "evals/fixtures/naming/violations.go", 12, "get-prefix", "GetName")
		requireFinding(t, result.Violations, "evals/fixtures/naming/violations.go", 12, "bad-receiver", "this")

		clean := filepath.Join(fixturesDir, "naming", "clean")
		out = runCommand(t, 0, "bash", script, "--json", clean)
		var cleanResult struct {
			Total int `json:"total"`
		}
		if err := json.Unmarshal(out, &cleanResult); err != nil {
			t.Fatalf("parse clean JSON: %v\n%s", err, out)
		}
		if cleanResult.Total != 0 {
			t.Fatalf("clean naming fixture produced %d violations\n%s", cleanResult.Total, out)
		}

		out = runCommand(t, 0, "bash", script, "--json", filepath.Join(fixturesDir, "no_go_files"))
		var emptyResult struct {
			Total     int    `json:"total"`
			Truncated bool   `json:"truncated"`
			Status    string `json:"status"`
		}
		if err := json.Unmarshal(out, &emptyResult); err != nil {
			t.Fatalf("parse no-Go naming JSON: %v\n%s", err, out)
		}
		if emptyResult.Total != 0 || emptyResult.Truncated || emptyResult.Status != "no_go_files" {
			t.Fatalf("unexpected no-Go naming result: %#v\n%s", emptyResult, out)
		}
	})

	t.Run("CheckDocs", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-documentation", "check-docs.sh")
		fixture := filepath.Join(fixturesDir, "docs", "violations.go")
		out := runCommand(t, 1, "bash", script, "--json", fixture)

		if !json.Valid(out) {
			t.Fatalf("--json output is not valid JSON:\n%s", out)
		}

		var result struct {
			Missing []jsonFinding `json:"missing"`
			Total   int           `json:"total"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatalf("parse JSON: %v\n%s", err, out)
		}
		if result.Total != 4 {
			t.Fatalf("expected exactly 4 undocumented symbols, got %d\n%s", result.Total, out)
		}
		requireDocMissing(t, result.Missing, "evals/fixtures/docs/violations.go", 1, "package", "docsbad")
		requireDocMissing(t, result.Missing, "evals/fixtures/docs/violations.go", 3, "type", "Cache")
		requireDocMissing(t, result.Missing, "evals/fixtures/docs/violations.go", 5, "function", "Map")
		requireDocMissing(t, result.Missing, "evals/fixtures/docs/violations.go", 9, "type", "Loader")

		out = runCommand(t, 1, "bash", script, fixture, "--json")
		if !json.Valid(out) {
			t.Fatalf("path-before-flag --json output is not valid JSON:\n%s", out)
		}

		documented := filepath.Join(fixturesDir, "docs", "documented")
		out = runCommand(t, 0, "bash", script, "--json", documented)
		var cleanResult struct {
			Missing []jsonFinding `json:"missing"`
			Total   int           `json:"total"`
		}
		if err := json.Unmarshal(out, &cleanResult); err != nil {
			t.Fatalf("parse documented JSON: %v\n%s", err, out)
		}
		if cleanResult.Total != 0 {
			t.Fatalf("documented package fixture produced %d findings\n%s", cleanResult.Total, out)
		}
		if cleanResult.Missing == nil || len(cleanResult.Missing) != 0 {
			t.Fatalf("documented package should emit an empty missing array, got %#v\n%s", cleanResult.Missing, out)
		}

		out = runCommand(t, 0, "bash", script, "--json", filepath.Join(fixturesDir, "no_go_files"))
		var emptyResult struct {
			Total     int    `json:"total"`
			Truncated bool   `json:"truncated"`
			Status    string `json:"status"`
		}
		if err := json.Unmarshal(out, &emptyResult); err != nil {
			t.Fatalf("parse no-Go docs JSON: %v\n%s", err, out)
		}
		if emptyResult.Total != 0 || emptyResult.Truncated || emptyResult.Status != "no_go_files" {
			t.Fatalf("unexpected no-Go docs result: %#v\n%s", emptyResult, out)
		}

		out = runCommand(t, 1, "bash", script, "--json", filepath.Join(fixturesDir, "docs", "detached"))
		var detachedResult struct {
			Missing []jsonFinding `json:"missing"`
			Total   int           `json:"total"`
		}
		if err := json.Unmarshal(out, &detachedResult); err != nil {
			t.Fatalf("parse detached docs JSON: %v\n%s", err, out)
		}
		if detachedResult.Total != 1 {
			t.Fatalf("detached package comment should produce one finding, got %d\n%s", detachedResult.Total, out)
		}
		requireDocMissing(t, detachedResult.Missing, "evals/fixtures/docs/detached/detached.go", 3, "package", "detached")

		malformedDir := t.TempDir()
		malformed := filepath.Join(malformedDir, "bad.go")
		if err := os.WriteFile(malformed, []byte("package malformed\n\nfunc Broken( {\n"), 0644); err != nil {
			t.Fatalf("write malformed fixture: %v", err)
		}
		parseable := filepath.Join(malformedDir, "good.go")
		if err := os.WriteFile(parseable, []byte("package malformed\n\ntype Exported struct{}\n"), 0644); err != nil {
			t.Fatalf("write parseable malformed-dir fixture: %v", err)
		}
		out = runCommand(t, 2, "bash", script, "--json", malformedDir)
		if !json.Valid(out) {
			t.Fatalf("malformed docs JSON output is not valid JSON:\n%s", out)
		}
		var malformedResult struct {
			Total       int    `json:"total"`
			Truncated   bool   `json:"truncated"`
			Status      string `json:"status"`
			ParseErrors []struct {
				File string `json:"file"`
			} `json:"parse_errors"`
		}
		if err := json.Unmarshal(out, &malformedResult); err != nil {
			t.Fatalf("parse malformed JSON: %v\n%s", err, out)
		}
		if malformedResult.Total != 2 || malformedResult.Truncated || malformedResult.Status != "parse_error" || len(malformedResult.ParseErrors) != 1 {
			t.Fatalf("unexpected malformed docs result: %#v\n%s", malformedResult, out)
		}
	})

	t.Run("CheckErrors", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-error-handling", "check-errors.sh")
		fixture := filepath.Join(fixturesDir, "errors", "violations.go")
		out := runCommand(t, 1, "bash", script, "--json", fixture)

		if !json.Valid(out) {
			t.Fatalf("--json output is not valid JSON:\n%s", out)
		}

		var result struct {
			Findings []jsonFinding `json:"findings"`
			Total    int           `json:"total"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatalf("parse JSON: %v\n%s", err, out)
		}
		if result.Total != 6 {
			t.Fatalf("expected exactly 6 error findings, got %d\n%s", result.Total, out)
		}
		requireFinding(t, result.Findings, "evals/fixtures/errors/violations.go", 11, "bare-return-err", "wrapping context")
		requireFinding(t, result.Findings, "evals/fixtures/errors/violations.go", 19, "string-error-compare", "errors.Is")
		requireFinding(t, result.Findings, "evals/fixtures/errors/violations.go", 22, "bare-return-err", "wrapping context")
		requireFinding(t, result.Findings, "evals/fixtures/errors/violations.go", 30, "log-and-return", "logged")
		requireFinding(t, result.Findings, "evals/fixtures/errors/violations.go", 31, "bare-return-err", "wrapping context")
		requireFinding(t, result.Findings, "evals/fixtures/errors/violations.go", 39, "bare-return-err", "wrapping context")

		out = runCommand(t, 1, "bash", script, "--json", "--no-bare-return", fixture)
		var noBareResult struct {
			Findings []jsonFinding `json:"findings"`
			Total    int           `json:"total"`
		}
		if err := json.Unmarshal(out, &noBareResult); err != nil {
			t.Fatalf("parse no-bare errors JSON: %v\n%s", err, out)
		}
		if noBareResult.Total != 2 {
			t.Fatalf("--no-bare-return should suppress 4 bare-return findings and leave 2, got %d\n%s", noBareResult.Total, out)
		}
		for _, finding := range noBareResult.Findings {
			if finding.Rule == "bare-return-err" {
				t.Fatalf("--no-bare-return emitted bare-return finding: %#v\n%s", finding, out)
			}
		}

		tupleFixture := filepath.Join(fixturesDir, "errors", "tuple", "tuple.go")
		out = runCommand(t, 1, "bash", script, "--json", tupleFixture)
		var tupleResult struct {
			Findings []jsonFinding `json:"findings"`
			Total    int           `json:"total"`
		}
		if err := json.Unmarshal(out, &tupleResult); err != nil {
			t.Fatalf("parse tuple errors JSON: %v\n%s", err, out)
		}
		if tupleResult.Total != 1 {
			t.Fatalf("3-result tuple return should produce one finding, got %d\n%s", tupleResult.Total, out)
		}
		requireFinding(t, tupleResult.Findings, "evals/fixtures/errors/tuple/tuple.go", 6, "bare-return-err", "wrapping context")

		multilineFixture := filepath.Join(fixturesDir, "errors", "multiline", "multiline.go")
		out = runCommand(t, 1, "bash", script, "--json", multilineFixture)
		var multilineResult struct {
			Findings []jsonFinding `json:"findings"`
			Total    int           `json:"total"`
		}
		if err := json.Unmarshal(out, &multilineResult); err != nil {
			t.Fatalf("parse multiline errors JSON: %v\n%s", err, out)
		}
		if multilineResult.Total != 1 {
			t.Fatalf("multiline 3-result tuple return should produce one finding, got %d\n%s", multilineResult.Total, out)
		}
		requireFinding(t, multilineResult.Findings, "evals/fixtures/errors/multiline/multiline.go", 6, "bare-return-err", "wrapping context")

		out = runCommand(t, 0, "bash", script, "--json", filepath.Join(fixturesDir, "errors", "clean"))
		var cleanResult struct {
			Total int `json:"total"`
		}
		if err := json.Unmarshal(out, &cleanResult); err != nil {
			t.Fatalf("parse clean errors JSON: %v\n%s", err, out)
		}
		if cleanResult.Total != 0 {
			t.Fatalf("clean error fixture produced %d findings\n%s", cleanResult.Total, out)
		}

		out = runCommand(t, 0, "bash", script, "--json", "--no-bare-return", filepath.Join(fixturesDir, "errors", "clean"))
		if err := json.Unmarshal(out, &cleanResult); err != nil {
			t.Fatalf("parse clean errors no-bare JSON: %v\n%s", err, out)
		}
		if cleanResult.Total != 0 {
			t.Fatalf("clean error fixture with --no-bare-return produced %d findings\n%s", cleanResult.Total, out)
		}

		out = runCommand(t, 0, "bash", script, "--json", filepath.Join(fixturesDir, "no_go_files"))
		var emptyResult struct {
			Total     int    `json:"total"`
			Truncated bool   `json:"truncated"`
			Status    string `json:"status"`
		}
		if err := json.Unmarshal(out, &emptyResult); err != nil {
			t.Fatalf("parse no-Go errors JSON: %v\n%s", err, out)
		}
		if emptyResult.Total != 0 || emptyResult.Truncated || emptyResult.Status != "no_go_files" {
			t.Fatalf("unexpected no-Go errors result: %#v\n%s", emptyResult, out)
		}
	})

	t.Run("CheckInterfaceCompliance", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-interfaces", "check-interface-compliance.sh")
		missingDir := filepath.Join(fixturesDir, "interfaces", "missing")
		out := runCommand(t, 1, "bash", script, "--json", missingDir)

		if !json.Valid(out) {
			t.Fatalf("--json output is not valid JSON:\n%s", out)
		}

		var result struct {
			Missing      []jsonFinding `json:"missing"`
			CountMissing int           `json:"count_missing"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatalf("parse JSON: %v\n%s", err, out)
		}
		if result.CountMissing != 1 {
			t.Fatalf("expected exactly 1 missing interface check, got %d\n%s", result.CountMissing, out)
		}
		requireInterfaceMissing(t, result.Missing, "evals/fixtures/interfaces/missing/missing.go", 4, "Runner")

		goodDir := filepath.Join(fixturesDir, "interfaces", "good")
		out = runCommand(t, 0, "bash", script, "--json", goodDir)
		var goodResult struct {
			CountMissing int `json:"count_missing"`
		}
		if err := json.Unmarshal(out, &goodResult); err != nil {
			t.Fatalf("parse good JSON: %v\n%s", err, out)
		}
		if goodResult.CountMissing != 0 {
			t.Fatalf("all-good interface fixture produced %d missing checks\n%s", goodResult.CountMissing, out)
		}

		consumerDir := filepath.Join(fixturesDir, "interfaces", "consumer")
		out = runCommand(t, 0, "bash", script, "--json", consumerDir)
		var consumerResult struct {
			CountInterfaces int `json:"count_interfaces"`
			CountMissing    int `json:"count_missing"`
		}
		if err := json.Unmarshal(out, &consumerResult); err != nil {
			t.Fatalf("parse consumer JSON: %v\n%s", err, out)
		}
		if consumerResult.CountInterfaces != 1 || consumerResult.CountMissing != 0 {
			t.Fatalf("consumer-owned interface should not require assertion, got %#v\n%s", consumerResult, out)
		}

		collisionDir := filepath.Join(fixturesDir, "interfaces", "collision")
		out = runCommand(t, 0, "bash", script, "--json", collisionDir)
		var collisionResult struct {
			CountInterfaces int `json:"count_interfaces"`
			CountMissing    int `json:"count_missing"`
		}
		if err := json.Unmarshal(out, &collisionResult); err != nil {
			t.Fatalf("parse collision JSON: %v\n%s", err, out)
		}
		if collisionResult.CountInterfaces != 1 || collisionResult.CountMissing != 0 {
			t.Fatalf("method-name collision should not count as implementation, got %#v\n%s", collisionResult, out)
		}

		dupDir := filepath.Join(fixturesDir, "interfaces", "duplicates")
		out = runCommand(t, 1, "bash", script, "--json", dupDir)
		var dupResult struct {
			Missing      []jsonFinding `json:"missing"`
			CountMissing int           `json:"count_missing"`
		}
		if err := json.Unmarshal(out, &dupResult); err != nil {
			t.Fatalf("parse duplicate JSON: %v\n%s", err, out)
		}
		if dupResult.CountMissing != 1 {
			t.Fatalf("duplicate interface fixture should have exactly one missing check, got %d\n%s", dupResult.CountMissing, out)
		}
		requireInterfaceMissing(t, dupResult.Missing, "evals/fixtures/interfaces/duplicates/b/runner.go", 4, "Runner")

		out = runCommand(t, 0, "bash", script, "--json", filepath.Join(fixturesDir, "no_go_files"))
		var emptyResult struct {
			CountInterfaces int    `json:"count_interfaces"`
			CountMissing    int    `json:"count_missing"`
			Truncated       bool   `json:"truncated"`
			Status          string `json:"status"`
		}
		if err := json.Unmarshal(out, &emptyResult); err != nil {
			t.Fatalf("parse no-Go interface JSON: %v\n%s", err, out)
		}
		if emptyResult.CountInterfaces != 0 || emptyResult.CountMissing != 0 || emptyResult.Truncated || emptyResult.Status != "no_go_files" {
			t.Fatalf("unexpected no-Go interface result: %#v\n%s", emptyResult, out)
		}

		noExportedDir := t.TempDir()
		noExportedFile := filepath.Join(noExportedDir, "local.go")
		if err := os.WriteFile(noExportedFile, []byte("package local\n\ntype runner interface{ Run() error }\n"), 0644); err != nil {
			t.Fatalf("write no-exported interface fixture: %v", err)
		}
		out = runCommand(t, 0, "bash", script, "--json", noExportedDir)
		var noExportedResult struct {
			CountInterfaces int    `json:"count_interfaces"`
			CountMissing    int    `json:"count_missing"`
			Truncated       bool   `json:"truncated"`
			Status          string `json:"status"`
		}
		if err := json.Unmarshal(out, &noExportedResult); err != nil {
			t.Fatalf("parse no-exported JSON: %v\n%s", err, out)
		}
		if noExportedResult.CountInterfaces != 0 || noExportedResult.CountMissing != 0 || noExportedResult.Truncated || noExportedResult.Status != "no_exported_interfaces" {
			t.Fatalf("unexpected no-exported interface result: %#v\n%s", noExportedResult, out)
		}
	})

	t.Run("GenTableTest", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-testing", "gen-table-test.sh")
		out := runCommand(t, 0, "bash", script, "--parallel", "ParseDuration", "parser")

		if _, fmtErr := format.Source(out); fmtErr != nil {
			t.Errorf("generated Go code is not valid:\n%v\n%s", fmtErr, out)
		}

		if !strings.Contains(string(out), "t.Parallel()") {
			t.Error("--parallel flag did not produce t.Parallel() in output")
		}
		out = runCommandStdout(t, 0, "bash", script, "--json", "--parallel", "ParseDuration", "parser")
		if !json.Valid(out) {
			t.Fatalf("gen-table --json stdout is not valid JSON:\n%s", out)
		}
		var jsonResult struct {
			Func       string `json:"func"`
			Package    string `json:"package"`
			OutputFile string `json:"output_file"`
			Parallel   bool   `json:"parallel"`
			Written    bool   `json:"written"`
		}
		if err := json.Unmarshal(out, &jsonResult); err != nil {
			t.Fatalf("parse gen-table JSON: %v\n%s", err, out)
		}
		if jsonResult.Func != "ParseDuration" || jsonResult.Package != "parser" || jsonResult.OutputFile != "" || !jsonResult.Parallel || jsonResult.Written {
			t.Fatalf("unexpected gen-table JSON metadata: %#v\n%s", jsonResult, out)
		}
		runCommand(t, 2, "bash", script, "parseDuration", "parser")
		runCommand(t, 2, "bash", script, "ParseDuration", "123bad")
		runCommand(t, 2, "bash", script, "ParseDuration", "func")
	})

	t.Run("SetupLintDryRun", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-linting", "setup-lint.sh")
		out := runCommand(t, 0, "bash", script, "--dry-run", "github.com/acme/project")
		s := string(out)
		if !strings.Contains(s, `version: "2"`) {
			t.Error("dry-run output missing golangci-lint v2 version")
		}
		if !strings.Contains(s, "linters:") {
			t.Error("dry-run output missing 'linters:' section")
		}
		if !strings.Contains(s, "formatters:") {
			t.Error("dry-run output missing 'formatters:' section")
		}
		golangciLint, err := exec.LookPath("golangci-lint")
		if err != nil {
			t.Fatal("golangci-lint not installed; config verification is required")
		}
		tmp := filepath.Join(t.TempDir(), ".golangci.yml")
		if err := os.WriteFile(tmp, out, 0644); err != nil {
			t.Fatalf("write temp golangci config: %v", err)
		}
		runCommand(t, 0, golangciLint, "config", "verify", "--config", tmp)
		runCommand(t, 0, golangciLint, "config", "verify", "--config", scriptPath("go-linting", "../assets/golangci.yml"))

		jsonDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(jsonDir, "go.mod"), []byte("module setuplintjson\n\ngo 1.22\n"), 0644); err != nil {
			t.Fatalf("write setup-lint temp go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(jsonDir, "main.go"), []byte("package setuplintjson\n\n// Good returns a stable value.\nfunc Good() int { return 1 }\n"), 0644); err != nil {
			t.Fatalf("write setup-lint temp source: %v", err)
		}
		out = runCommandInDir(t, jsonDir, 0, "bash", script, "--json", "github.com/acme/project")
		if !json.Valid(out) {
			t.Fatalf("setup-lint --json output is not valid JSON:\n%s", out)
		}
		var setupJSON struct {
			ConfigPath  string `json:"config_path"`
			LocalPrefix string `json:"local_prefix"`
			Created     bool   `json:"created"`
			LintIssues  bool   `json:"lint_issues"`
			LintOutput  string `json:"lint_output"`
		}
		if err := json.Unmarshal(out, &setupJSON); err != nil {
			t.Fatalf("parse setup-lint JSON: %v\n%s", err, out)
		}
		if setupJSON.ConfigPath != ".golangci.yml" || setupJSON.LocalPrefix != "github.com/acme/project" || !setupJSON.Created || setupJSON.LintIssues {
			t.Fatalf("unexpected setup-lint JSON metadata: %#v\n%s", setupJSON, out)
		}
	})

	t.Run("PreReview", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-code-review", "pre-review.sh")
		out := runCommand(t, 1, "bash", script, "--json", "--force", filepath.Join(fixturesDir, "pre_review_missing"))
		if !json.Valid(out) {
			t.Fatalf("--json output is not valid JSON:\n%s", out)
		}
		var failedResult struct {
			Gofmt struct {
				Status string   `json:"status"`
				Files  []string `json:"files"`
			} `json:"gofmt"`
			Passed bool `json:"passed"`
		}
		if err := json.Unmarshal(out, &failedResult); err != nil {
			t.Fatalf("parse pre-review failure JSON: %v\n%s", err, out)
		}
		if failedResult.Gofmt.Status != "fail" || len(failedResult.Gofmt.Files) == 0 || failedResult.Passed {
			t.Fatalf("pre-review missing fixture should fail gofmt and passed=false, got %#v\n%s", failedResult, out)
		}

		cleanDir := filepath.Join(fixturesDir, "pre_review_clean")
		out = runCommandInDir(t, cleanDir, 0, "bash", script, "--json", "--force", ".")
		var cleanResult struct {
			Gofmt struct {
				Status string `json:"status"`
			} `json:"gofmt"`
			Govet struct {
				Status string `json:"status"`
			} `json:"govet"`
			Passed bool `json:"passed"`
		}
		if err := json.Unmarshal(out, &cleanResult); err != nil {
			t.Fatalf("parse pre-review clean JSON: %v\n%s", err, out)
		}
		if cleanResult.Gofmt.Status != "pass" || cleanResult.Govet.Status != "pass" || !cleanResult.Passed {
			t.Fatalf("pre-review clean fixture should pass gofmt/go vet, got %#v\n%s", cleanResult, out)
		}
	})

	t.Run("CheckDocsStrict", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-documentation", "check-docs.sh")
		fixture := filepath.Join(fixturesDir, "docs", "strict.go")
		out := runCommand(t, 1, "bash", script, "--json", "--strict", fixture)
		if !json.Valid(out) {
			t.Fatalf("--strict --json output is not valid JSON:\n%s", out)
		}
		var result struct {
			Missing []jsonFinding `json:"missing"`
			Total   int           `json:"total"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatalf("parse JSON: %v\n%s", err, out)
		}
		if result.Total != 2 {
			t.Fatalf("--strict should find exactly package + unexported function, got %d\n%s", result.Total, out)
		}
		requireDocMissing(t, result.Missing, "evals/fixtures/docs/strict.go", 1, "package", "docsbad")
		requireDocMissing(t, result.Missing, "evals/fixtures/docs/strict.go", 3, "function", "helperThatNeedsDocs")
	})

	t.Run("BenchCompare", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-performance", "bench-compare.sh")
		out := runCommandStdout(t, 0, "bash", script, "--json", "--limit", "1", "-n", "1", filepath.Join(fixturesDir, "bench"))
		if !json.Valid(out) {
			t.Fatalf("bench JSON output is not valid JSON:\n%s", out)
		}
		var result struct {
			BenchmarksFound int `json:"benchmarks_found"`
			Count           int `json:"count"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatalf("parse bench JSON: %v\n%s", err, out)
		}
		if result.BenchmarksFound != 1 || result.Count != 1 {
			t.Fatalf("unexpected bench metadata: %#v\n%s", result, out)
		}
		runCommand(t, 1, "bash", script, "-n", "1", filepath.Join(fixturesDir, "bench", "nobench"))
		runCommand(t, 2, "bash", script, "-n", "nope", filepath.Join(fixturesDir, "bench"))
	})

	t.Run("InvalidLimitFlags", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			skill string
			name  string
			args  []string
		}{
			{"go-code-review", "pre-review.sh", []string{"--force"}},
			{"go-documentation", "check-docs.sh", nil},
			{"go-error-handling", "check-errors.sh", nil},
			{"go-interfaces", "check-interface-compliance.sh", nil},
			{"go-linting", "setup-lint.sh", []string{"--dry-run"}},
			{"go-naming", "check-naming.sh", nil},
			{"go-performance", "bench-compare.sh", nil},
			{"go-testing", "gen-table-test.sh", []string{"ParseDuration", "parser"}},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				script := scriptPath(tc.skill, tc.name)
				args := append([]string{script, "--limit", "nope"}, tc.args...)
				runCommand(t, 2, "bash", args...)
			})
		}
	})
}

// ---------------------------------------------------------------------------
// TestStructure - validate SKILL.md frontmatter across all skills
// ---------------------------------------------------------------------------

func TestStructure(t *testing.T) {
	t.Parallel()
	skillDirs := findSkillDirs(t)

	for _, dir := range skillDirs {
		dirName := filepath.Base(dir)
		dir := dir
		t.Run(dirName, func(t *testing.T) {
			t.Parallel()
			skillFile := filepath.Join(dir, "SKILL.md")
			content, err := os.ReadFile(skillFile)
			if err != nil {
				t.Fatalf("read SKILL.md: %v", err)
			}

			name, desc, body := parseFrontmatter(content)

			if name != dirName {
				t.Errorf("frontmatter name %q does not match directory %q", name, dirName)
			}
			if desc == "" {
				t.Error("description is empty")
			}
			if !strings.HasPrefix(desc, "Use when ") {
				t.Errorf("description must preserve trigger-oriented 'Use when ...' shape, got %q", desc)
			}
			if len(desc) > 1024 {
				t.Errorf("description is %d chars (max 1024)", len(desc))
			}
			versionSensitive := map[string]bool{
				"go-code-review":    true,
				"go-concurrency":    true,
				"go-context":        true,
				"go-declarations":   true,
				"go-defensive":      true,
				"go-error-handling": true,
				"go-generics":       true,
				"go-logging":        true,
				"go-testing":        true,
			}
			if versionSensitive[dirName] && !strings.Contains(body, "> Compatibility:") {
				t.Error("version-sensitive skill must preserve compatibility metadata in the body")
			}
			frontmatterKeys := map[string]bool{}
			if fm, _, ok := splitFrontmatter(content); ok {
				for _, line := range strings.Split(fm, "\n") {
					line = strings.TrimSpace(line)
					if line == "" || strings.HasPrefix(line, "#") {
						continue
					}
					key := strings.TrimSpace(strings.SplitN(line, ":", 2)[0])
					frontmatterKeys[key] = true
					if key != "name" && key != "description" && key != "allowed-tools" {
						t.Errorf("non-portable frontmatter key %q; keep only name, description, and runtime-required allowed-tools", key)
					}
				}
			}

			bodyLines := strings.Count(body, "\n")
			if bodyLines >= 500 {
				t.Errorf("body is %d lines (spec recommends < 500)", bodyLines)
			}

			// Check shebang on all .sh files
			scriptDir := filepath.Join(dir, "scripts")
			if entries, err := os.ReadDir(scriptDir); err == nil {
				hasShellScript := false
				for _, e := range entries {
					if !strings.HasSuffix(e.Name(), ".sh") {
						continue
					}
					hasShellScript = true
					shContent, err := os.ReadFile(filepath.Join(scriptDir, e.Name()))
					if err != nil {
						t.Errorf("read %s: %v", e.Name(), err)
						continue
					}
					if !strings.HasPrefix(string(shContent), "#!/usr/bin/env bash") {
						t.Errorf("%s: missing #!/usr/bin/env bash shebang", e.Name())
					}
				}
				if hasShellScript && !frontmatterKeys["allowed-tools"] {
					t.Errorf("script-backed skill is missing allowed-tools frontmatter grant")
				}
			} else if frontmatterKeys["allowed-tools"] {
				t.Errorf("allowed-tools frontmatter is only permitted for script-backed skills")
			}
		})
	}
}

func TestFrontmatterDescriptionsInvariant(t *testing.T) {
	t.Parallel()
	want := map[string]string{
		"go-code-review":        "Use when reviewing Go code or checking code against community style standards. Also use proactively before submitting a Go PR or when reviewing any Go code changes, even if the user doesn't explicitly request a style review. Does not cover language-specific syntax — delegates to specialized skills.",
		"go-concurrency":        "Use when writing concurrent Go code — goroutines, channels, mutexes, or thread-safety guarantees. Also use when parallelizing work, fixing data races, or protecting shared state, even if the user doesn't explicitly mention concurrency primitives. Does not cover context.Context patterns (see go-context).",
		"go-context":            "Use when working with context.Context in Go — placement in signatures, propagating cancellation and deadlines, and storing values in context vs parameters. Also use when cancelling long-running operations, setting timeouts, or passing request-scoped data, even if they don't mention context.Context directly. Does not cover goroutine lifecycle or sync primitives (see go-concurrency).",
		"go-control-flow":       "Use when writing conditionals, loops, or switch statements in Go — including if with initialization, early returns, for loop forms, range, switch, type switches, and blank identifier patterns. Also use when writing a simple if/else or for loop, even if the user doesn't mention guard clauses or variable scoping. Does not cover error flow patterns (see go-error-handling).",
		"go-data-structures":    "Use when working with Go slices, maps, or arrays — choosing between new and make, using append, declaring empty slices (nil vs literal for JSON), implementing sets with maps, and copying data at boundaries. Also use when building or manipulating collections, even if the user doesn't ask about allocation idioms. Does not cover concurrent data structure safety (see go-concurrency).",
		"go-declarations":       "Use when declaring or initializing Go variables, constants, structs, or maps — including var vs :=, reducing scope with if-init, formatting composite literals, designing iota enums, and using any instead of interface{}. Also use when writing a new struct or const block, even if the user doesn't ask about declaration style. Does not cover naming conventions (see go-naming).",
		"go-defensive":          "Use when hardening Go code at API boundaries — copying slices/maps, verifying interface compliance, using defer for cleanup, time.Time/time.Duration, or avoiding mutable globals. Also use when reviewing for robustness concerns like missing cleanup or unsafe crypto usage, even if the user doesn't mention \"defensive programming.\" Does not cover error handling strategy (see go-error-handling).",
		"go-documentation":      "Use when writing or reviewing documentation for Go packages, types, functions, or methods. Also use proactively when creating new exported types, functions, or packages, even if the user doesn't explicitly ask about documentation. Does not cover code comments for non-exported symbols (see go-style-core).",
		"go-error-handling":     "Use when writing Go code that returns, wraps, or handles errors — choosing between sentinel errors, custom types, and fmt.Errorf (%w vs %v), structuring error flow, or deciding whether to log or return. Also use when propagating errors across package boundaries or using errors.Is/As, even if the user doesn't ask about error strategy. Does not cover panic/recover patterns (see go-defensive).",
		"go-functional-options": "Use when designing a Go constructor or factory function with optional configuration — especially with 3+ optional parameters or extensible APIs. Also use when building a New* function that takes many settings, even if they don't mention \"functional options\" by name. Does not cover general function design (see go-functions).",
		"go-functions":          "Use when organizing functions within a Go file, formatting function signatures, designing return values, or following Printf-style naming conventions. Also use when a user is adding or refactoring any Go function, even if they don't mention function design or signature formatting. Does not cover functional options constructors (see go-functional-options).",
		"go-generics":           "Use when deciding whether to use Go generics, writing generic functions or types, choosing constraints, or picking between type aliases and type definitions. Also use when a user is writing a utility function that could work with multiple types, even if they don't mention generics explicitly. Does not cover interface design without generics (see go-interfaces).",
		"go-interfaces":         "Use when defining or implementing Go interfaces, designing abstractions, creating mockable boundaries for testing, or composing types through embedding. Also use when deciding whether to accept an interface or return a concrete type, or using type assertions or type switches, even if the user doesn't explicitly mention interfaces. Does not cover generics-based polymorphism (see go-generics).",
		"go-linting":            "Use when setting up linting for a Go project, configuring golangci-lint, or adding Go checks to a CI/CD pipeline. Also use when starting a new Go project and deciding which linters to enable, even if the user only asks about \"code quality\" or \"static analysis\" without mentioning specific linter names. Does not cover code review process (see go-code-review).",
		"go-logging":            "Use when choosing a logging approach, configuring slog, writing structured log statements, or deciding log levels in Go. Also use when setting up production logging, adding request-scoped context to logs, or migrating from log to slog, even if the user doesn't explicitly mention logging. Does not cover error handling strategy (see go-error-handling).",
		"go-naming":             "Use when naming any Go identifier — packages, types, functions, methods, variables, constants, or receivers — to ensure idiomatic, clear names. Also use when a user is creating new types, packages, or exported APIs, even if they don't explicitly ask about naming conventions. Does not cover package organization (see go-packages).",
		"go-packages":           "Use when creating Go packages, organizing imports, managing dependencies, or deciding how to structure Go code into packages. Also use when starting a new Go project or splitting a growing codebase into packages, even if the user doesn't explicitly ask about package organization. Does not cover naming individual identifiers (see go-naming).",
		"go-performance":        "Use when optimizing Go code, investigating slow performance, or writing performance-critical sections. Also use when a user mentions slow Go code, string concatenation in loops, or asks about benchmarking, even if the user doesn't explicitly mention performance patterns. Does not cover concurrent performance patterns (see go-concurrency).",
		"go-style-core":         "Use when working with Go formatting, line length, nesting, naked returns, semicolons, or core style principles. Also use when a style question isn't covered by a more specific skill, even if the user doesn't reference a specific style rule. Does not cover domain-specific patterns like error handling, naming, or testing (see specialized skills). Acts as fallback when no more specific style skill applies.",
		"go-testing":            "Use when writing, reviewing, or improving Go test code — including table-driven tests, subtests, parallel tests, test helpers, test doubles, and assertions with cmp.Diff. Also use when a user asks to write a test for a Go function, even if they don't mention specific patterns like table-driven tests or subtests. Does not cover benchmark performance testing (see go-performance).",
	}
	for _, dir := range findSkillDirs(t) {
		dir := dir
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
			if err != nil {
				t.Fatalf("read SKILL.md: %v", err)
			}
			name, desc, _ := parseFrontmatter(content)
			if got, ok := want[name]; !ok {
				t.Fatalf("missing description golden for %s", name)
			} else if desc != got {
				t.Fatalf("description changed for %s\n got: %q\nwant: %q", name, desc, got)
			}
		})
	}
	if len(want) != len(findSkillDirs(t)) {
		t.Fatalf("description golden count = %d, skill count = %d", len(want), len(findSkillDirs(t)))
	}
}

func TestSkillArchitecture(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)

	for _, rel := range []string{
		"docs/SKILL_AUTHORING_TEMPLATE.md",
		"docs/RULE_OWNERSHIP.md",
		"docs/SCRIPT_JSON_CONTRACTS.md",
	} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("missing architecture doc %s: %v", rel, err)
		}
	}

	for _, dir := range findSkillDirs(t) {
		dir := dir
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Parallel()
			contentBytes, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
			if err != nil {
				t.Fatalf("read SKILL.md: %v", err)
			}
			content := string(contentBytes)
			if count := strings.Count(content, "\n## Resource Routing\n"); count != 1 {
				t.Fatalf("Resource Routing section count = %d, want 1", count)
			}
			if !strings.Contains(content, "\n## Related Skills\n") {
				t.Fatal("missing ## Related Skills section")
			}
			if strings.Contains(content, "\n## Available Scripts\n") {
				t.Fatal("scripts must be listed in Resource Routing, not a second Available Scripts section")
			}
			if strings.Contains(content, "> Read [references/") {
				t.Fatal("reference routing must live in Resource Routing, not inline blockquotes")
			}
			danglingRouting := regexp.MustCompile(`(?m)^> (implementing|writing|reviewing|choosing|working|designing|TestMain|custom|helpers|when writing|test helpers)\b`)
			if match := danglingRouting.FindString(content); match != "" {
				t.Fatalf("dangling reference-routing fragment remains: %q", match)
			}
			if lines := strings.Count(content, "\n") + 1; lines > 225 {
				t.Fatalf("SKILL.md has %d lines, want <= 225; move bulky examples into references", lines)
			}
			maxBlock := maxFencedBlockLines(content)
			if maxBlock > 40 {
				t.Fatalf("largest fenced code block has %d lines, want <= 40; move long examples into references", maxBlock)
			}
			idx := strings.Index(content, "\n## Resource Routing\n")
			if idx < 0 {
				t.Fatal("missing ## Resource Routing section")
			}
			rest := content[idx+len("\n## Resource Routing\n"):]
			end := strings.Index(rest, "\n## ")
			routing := rest
			if end >= 0 {
				routing = rest[:end]
			}
			if strings.TrimSpace(routing) == "" {
				t.Fatal("empty Resource Routing section")
			}

			for _, subdir := range []string{"references", "scripts", "assets"} {
				subPath := filepath.Join(dir, subdir)
				entries, err := os.ReadDir(subPath)
				if err != nil {
					continue
				}
				for _, e := range entries {
					if e.IsDir() {
						continue
					}
					ref := subdir + "/" + e.Name()
					if !strings.Contains(routing, ref) {
						t.Errorf("%s not listed in Resource Routing section", ref)
					}
				}
			}
		})
	}
}

func maxFencedBlockLines(content string) int {
	inBlock := false
	count := 0
	maxCount := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if inBlock {
				if count > maxCount {
					maxCount = count
				}
				count = 0
				inBlock = false
			} else {
				inBlock = true
			}
			continue
		}
		if inBlock {
			count++
		}
	}
	return maxCount
}

func TestLongReferenceTOCs(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	err := filepath.WalkDir(filepath.Join(root, "skills"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.Contains(path, string(filepath.Separator)+"references"+string(filepath.Separator)) || !strings.HasSuffix(path, ".md") {
			return nil
		}
		lines := readLines(t, path)
		if len(lines) > 300 {
			t.Errorf("%s has %d lines, want <= 300; split or trim oversized references", path, len(lines))
		}
		if len(lines) <= 200 {
			return nil
		}
		limit := 40
		if len(lines) < limit {
			limit = len(lines)
		}
		found := false
		for _, line := range lines[:limit] {
			if strings.TrimSpace(line) == "## Contents" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s has %d lines and no ## Contents section near the top", path, len(lines))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk references: %v", err)
	}
}

func TestRuleOwnershipMap(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "docs", "RULE_OWNERSHIP.md"))
	if err != nil {
		t.Fatalf("read ownership map: %v", err)
	}
	reSkill := regexp.MustCompile("`(go-[a-z0-9-]+)`")
	seen := map[string]bool{}
	for _, match := range reSkill.FindAllStringSubmatch(string(content), -1) {
		name := match[1]
		seen[name] = true
		if _, err := os.Stat(filepath.Join(root, "skills", name)); err != nil {
			t.Errorf("ownership map references missing skill %s", name)
		}
	}
	if len(seen) < len(findSkillDirs(t))/2 {
		t.Fatalf("ownership map references too few skills: %d", len(seen))
	}

	template, err := os.ReadFile(filepath.Join(root, "docs", "SKILL_AUTHORING_TEMPLATE.md"))
	if err != nil {
		t.Fatalf("read skill authoring template: %v", err)
	}
	if !strings.Contains(string(template), "RULE_OWNERSHIP.md") {
		t.Fatal("skill authoring template must route authors to RULE_OWNERSHIP.md before duplicating rules")
	}

	assertOnlySkillDoc(t, root, "loggerFromCtx", map[string]bool{
		"skills/go-logging/references/LOGGING-PATTERNS.md": true,
	})
	assertOnlySkillDoc(t, root, "loggerKey", map[string]bool{
		"skills/go-logging/references/LOGGING-PATTERNS.md": true,
	})
	assertNoSkillDocContains(t, root, "var _ io.Writer = (*MyType)(nil)")

	interfacesSkill, err := os.ReadFile(filepath.Join(root, "skills", "go-interfaces", "SKILL.md"))
	if err != nil {
		t.Fatalf("read go-interfaces SKILL.md: %v", err)
	}
	if strings.Contains(string(interfacesSkill), "Compile-time checks**: See [go-defensive]") {
		t.Fatal("go-interfaces must not route compile-time assertion ownership back to go-defensive")
	}

	optionsRef, err := os.ReadFile(filepath.Join(root, "skills", "go-functional-options", "references", "OPTIONS-VS-STRUCTS.md"))
	if err != nil {
		t.Fatalf("read functional options reference: %v", err)
	}
	if !strings.Contains(string(optionsRef), "db.WithCache(false)") {
		t.Fatal("functional options reference must preserve caller-ergonomics before/after example")
	}
}

func TestKnownReferenceRegressions(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	read := func(rel string) string {
		t.Helper()
		content, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		return string(content)
	}

	webServer := read("skills/go-code-review/references/WEB-SERVER.md")
	if strings.Contains(webServer, "httpSrv.Shutdown(ctx)\n") {
		t.Fatal("WEB-SERVER.md must handle Shutdown errors explicitly")
	}
	if strings.Contains(webServer, `"fmt"`) {
		t.Fatal("WEB-SERVER.md must not keep unused fmt import in the example")
	}
	if !strings.Contains(webServer, "[go-logging]") {
		t.Fatal("WEB-SERVER.md logging row must route logging guidance to go-logging")
	}

	contextPatterns := read("skills/go-context/references/PATTERNS.md")
	staleTimeout := regexp.MustCompile(`(?s)r\.Context\(\).*ReadTimeout|ReadTimeout.*fires|WriteTimeout.*fires`)
	if staleTimeout.MatchString(contextPatterns) {
		t.Fatal("PATTERNS.md must not claim request context cancellation is caused by ReadTimeout/WriteTimeout firing")
	}

	integration := read("skills/go-testing/references/INTEGRATION.md")
	if strings.Contains(integration, "ExerciseGame") {
		t.Fatal("INTEGRATION.md validation helper example must not call stale ExerciseGame")
	}
	if !strings.Contains(integration, "ExercisePlayer") {
		t.Fatal("INTEGRATION.md validation helper example should consistently use ExercisePlayer")
	}
	staleClientContext := "client.GetUser(" + "context.Background()"
	if strings.Contains(integration, staleClientContext) {
		t.Fatal("INTEGRATION.md *testing.T examples should use t.Context on supported Go versions")
	}

	dataStructures := read("skills/go-data-structures/SKILL.md")
	staleSetRow := "| Sets | `" + "map[T]bool`"
	if strings.Contains(dataStructures, staleSetRow) {
		t.Fatal("go-data-structures quick reference must not recommend map[T]bool for membership-only sets")
	}
	if !strings.Contains(dataStructures, "map[T]struct{}") {
		t.Fatal("go-data-structures must document map[T]struct{} for membership-only sets")
	}

	benchmarks := read("skills/go-performance/references/BENCHMARKS.md")
	if !strings.Contains(benchmarks, "b.Loop()") {
		t.Fatal("BENCHMARKS.md should prefer b.Loop for Go 1.24+ benchmark loops")
	}
	if !strings.Contains(benchmarks, "for i := 0; i < b.N; i++") {
		t.Fatal("BENCHMARKS.md should keep b.N fallback guidance for older Go versions")
	}

	advancedConcurrency := read("skills/go-concurrency/references/ADVANCED-PATTERNS.md")
	if !strings.Contains(advancedConcurrency, "wg.Go(") {
		t.Fatal("ADVANCED-PATTERNS.md should use WaitGroup.Go in Go 1.25+ examples")
	}
}

func assertOnlySkillDoc(t *testing.T, root, needle string, allowed map[string]bool) {
	t.Helper()
	err := filepath.WalkDir(filepath.Join(root, "skills"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(content), needle) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !allowed[rel] {
			t.Errorf("%q appears in non-owner document %s", needle, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk skill docs: %v", err)
	}
}

func assertNoSkillDocContains(t *testing.T, root, needle string) {
	t.Helper()
	assertOnlySkillDoc(t, root, needle, map[string]bool{})
}

func TestFixtureLayout(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	matches, err := filepath.Glob(filepath.Join(root, "evals", "fixtures", "*.go"))
	if err != nil {
		t.Fatalf("glob root fixtures: %v", err)
	}
	if len(matches) > 0 {
		t.Fatalf("orphaned root fixture files remain: %v", matches)
	}
}

func requireJSONKeySet(t *testing.T, label string, obj map[string]json.RawMessage, required, optional []string) {
	t.Helper()
	allowed := map[string]bool{}
	for _, key := range required {
		allowed[key] = true
		if _, ok := obj[key]; !ok {
			t.Fatalf("%s missing required key %q", label, key)
		}
	}
	for _, key := range optional {
		allowed[key] = true
	}
	for key := range obj {
		if !allowed[key] {
			t.Fatalf("%s has unknown key %q", label, key)
		}
	}
}

func TestEvalsJSONSchema(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "evals", "evals.json"))
	if err != nil {
		t.Fatalf("read evals.json: %v", err)
	}
	type triggerEval struct {
		Query         string   `json:"query"`
		ShouldTrigger []string `json:"should_trigger"`
		Notes         string   `json:"notes"`
		Set           string   `json:"set"`
	}
	type qualityEval struct {
		ID             int      `json:"id"`
		Prompt         string   `json:"prompt"`
		Files          []string `json:"files"`
		ExpectedOutput string   `json:"expected_output"`
		TargetSkills   []string `json:"target_skills"`
		Assertions     []string `json:"assertions"`
		Set            string   `json:"set"`
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(content, &top); err != nil {
		t.Fatalf("parse evals.json: %v", err)
	}
	requireJSONKeySet(t, "evals.json", top, []string{"description", "trigger_evals", "quality_evals"}, nil)

	var description string
	if err := json.Unmarshal(top["description"], &description); err != nil {
		t.Fatalf("parse evals.json description: %v", err)
	}
	if description == "" {
		t.Fatal("evals.json description is empty")
	}

	var triggerRaw []json.RawMessage
	if err := json.Unmarshal(top["trigger_evals"], &triggerRaw); err != nil {
		t.Fatalf("parse trigger_evals: %v", err)
	}
	if len(triggerRaw) == 0 {
		t.Fatal("evals.json has no trigger_evals")
	}
	triggerEvals := make([]triggerEval, 0, len(triggerRaw))
	for i, raw := range triggerRaw {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(raw, &obj); err != nil {
			t.Fatalf("parse trigger_evals[%d] as object: %v", i, err)
		}
		requireJSONKeySet(t, "trigger_evals", obj, []string{"query", "should_trigger", "notes", "set"}, nil)
		var ev triggerEval
		if err := json.Unmarshal(raw, &ev); err != nil {
			t.Fatalf("parse trigger_evals[%d]: %v", i, err)
		}
		triggerEvals = append(triggerEvals, ev)
	}

	var qualityRaw []json.RawMessage
	if err := json.Unmarshal(top["quality_evals"], &qualityRaw); err != nil {
		t.Fatalf("parse quality_evals: %v", err)
	}
	if len(qualityRaw) == 0 {
		t.Fatal("evals.json has no quality_evals")
	}
	qualityEvals := make([]qualityEval, 0, len(qualityRaw))
	for i, raw := range qualityRaw {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(raw, &obj); err != nil {
			t.Fatalf("parse quality_evals[%d] as object: %v", i, err)
		}
		requireJSONKeySet(t, "quality_evals", obj, []string{"id", "prompt", "expected_output", "target_skills", "assertions", "set"}, []string{"files"})
		var ev qualityEval
		if err := json.Unmarshal(raw, &ev); err != nil {
			t.Fatalf("parse quality_evals[%d]: %v", i, err)
		}
		qualityEvals = append(qualityEvals, ev)
	}

	skills := map[string]bool{}
	for _, dir := range findSkillDirs(t) {
		skills[filepath.Base(dir)] = true
	}
	validSet := func(set string) bool {
		return set == "train" || set == "validation"
	}
	for i, ev := range triggerEvals {
		if ev.Query == "" {
			t.Errorf("trigger_evals[%d] query is empty", i)
		}
		if ev.Notes == "" {
			t.Errorf("trigger_evals[%d] notes is empty", i)
		}
		if !validSet(ev.Set) {
			t.Errorf("trigger_evals[%d] set = %q, want train or validation", i, ev.Set)
		}
		for _, skill := range ev.ShouldTrigger {
			if !skills[skill] {
				t.Errorf("trigger_evals[%d] references missing skill %q", i, skill)
			}
		}
	}

	seenIDs := map[int]bool{}
	for i, ev := range qualityEvals {
		if ev.ID <= 0 {
			t.Errorf("quality_evals[%d] id must be positive, got %d", i, ev.ID)
		}
		if seenIDs[ev.ID] {
			t.Errorf("quality_evals[%d] duplicate id %d", i, ev.ID)
		}
		seenIDs[ev.ID] = true
		if ev.Prompt == "" {
			t.Errorf("quality_evals[%d] prompt is empty", i)
		}
		if ev.ExpectedOutput == "" {
			t.Errorf("quality_evals[%d] expected_output is empty", i)
		}
		if !validSet(ev.Set) {
			t.Errorf("quality_evals[%d] set = %q, want train or validation", i, ev.Set)
		}
		if len(ev.TargetSkills) == 0 {
			t.Errorf("quality_evals[%d] target_skills is empty", i)
		}
		for _, skill := range ev.TargetSkills {
			if !skills[skill] {
				t.Errorf("quality_evals[%d] references missing skill %q", i, skill)
			}
		}
		if len(ev.Assertions) == 0 {
			t.Errorf("quality_evals[%d] assertions is empty", i)
		}
		for j, assertion := range ev.Assertions {
			if assertion == "" {
				t.Errorf("quality_evals[%d].assertions[%d] is empty", i, j)
			}
		}
		for _, rel := range ev.Files {
			if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
				t.Errorf("quality_evals[%d] file %q missing: %v", i, rel, err)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// TestCrossRefs - verify all file references between skills resolve
// ---------------------------------------------------------------------------

func TestCrossRefs(t *testing.T) {
	t.Parallel()
	skillDirs := findSkillDirs(t)
	reFileRef := regexp.MustCompile(`\((?:references|scripts|assets)/[^)]+\)`)
	reCrossSkill := regexp.MustCompile(`\(\.\./go-[^/]+/SKILL\.md\)`)

	for _, dir := range skillDirs {
		dirName := filepath.Base(dir)
		dir := dir
		t.Run(dirName+"/links", func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
			if err != nil {
				t.Fatalf("read SKILL.md: %v", err)
			}

			refs := reFileRef.FindAllString(string(content), -1)
			for _, ref := range refs {
				relPath := ref[1 : len(ref)-1] // strip parens
				absPath := filepath.Join(dir, relPath)
				if _, err := os.Stat(absPath); os.IsNotExist(err) {
					t.Errorf("broken reference: %s -> %s", dirName, relPath)
				}
			}

			crossRefs := reCrossSkill.FindAllString(string(content), -1)
			for _, ref := range crossRefs {
				relPath := ref[1 : len(ref)-1]
				absPath := filepath.Join(dir, relPath)
				if _, err := os.Stat(absPath); os.IsNotExist(err) {
					t.Errorf("broken cross-skill reference: %s -> %s", dirName, relPath)
				}
			}
		})

		t.Run(dirName+"/orphans", func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
			if err != nil {
				t.Fatalf("read SKILL.md: %v", err)
			}
			skillContent := string(content)

			for _, subdir := range []string{"references", "scripts", "assets"} {
				subPath := filepath.Join(dir, subdir)
				entries, err := os.ReadDir(subPath)
				if err != nil {
					continue
				}
				for _, e := range entries {
					if e.IsDir() {
						continue
					}
					ref := subdir + "/" + e.Name()
					if !strings.Contains(skillContent, ref) {
						t.Errorf("orphaned file: %s/%s not referenced in SKILL.md", dirName, ref)
					}
				}
			}
		})
	}
}
