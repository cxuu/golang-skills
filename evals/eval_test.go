package evals_test

import (
	"bufio"
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

// parseFrontmatter extracts name, description, and body from SKILL.md content.
func parseFrontmatter(content []byte) (name, desc, body string) {
	s := string(content)
	parts := strings.SplitN(s, "---", 3)
	if len(parts) < 3 {
		return "", "", s
	}
	fm := parts[1]
	body = parts[2]

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
		cmd := exec.Command("bash", script, "--json", fixturesDir)
		out, _ := cmd.CombinedOutput()

		if !json.Valid(out) {
			t.Fatalf("--json output is not valid JSON:\n%s", out)
		}

		var result struct {
			Total int `json:"total"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatalf("parse JSON: %v\n%s", err, out)
		}
		if result.Total < 3 {
			t.Errorf("expected >= 3 naming violations, got %d\n%s", result.Total, out)
		}
	})

	t.Run("CheckDocs", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-documentation", "check-docs.sh")
		fixture := filepath.Join(fixturesDir, "undocumented.go")
		cmd := exec.Command("bash", script, "--json", fixture)
		out, _ := cmd.CombinedOutput()

		if !json.Valid(out) {
			t.Fatalf("--json output is not valid JSON:\n%s", out)
		}

		var result struct {
			Total int `json:"total"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatalf("parse JSON: %v\n%s", err, out)
		}
		if result.Total < 2 {
			t.Errorf("expected >= 2 undocumented symbols, got %d\n%s", result.Total, out)
		}
	})

	t.Run("CheckErrors", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-error-handling", "check-errors.sh")
		fixture := filepath.Join(fixturesDir, "bad_errors.go")
		cmd := exec.Command("bash", script, "--json", fixture)
		out, _ := cmd.CombinedOutput()

		if !json.Valid(out) {
			t.Fatalf("--json output is not valid JSON:\n%s", out)
		}

		var result struct {
			Total int `json:"total"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatalf("parse JSON: %v\n%s", err, out)
		}
		if result.Total < 2 {
			t.Errorf("expected >= 2 error anti-patterns, got %d\n%s", result.Total, out)
		}
	})

	t.Run("CheckInterfaceCompliance", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-interfaces", "check-interface-compliance.sh")
		cmd := exec.Command("bash", script, "--json", fixturesDir)
		out, _ := cmd.CombinedOutput()

		if !json.Valid(out) {
			t.Fatalf("--json output is not valid JSON:\n%s", out)
		}

		var result struct {
			CountMissing int `json:"count_missing"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatalf("parse JSON: %v\n%s", err, out)
		}
		if result.CountMissing < 1 {
			t.Errorf("expected >= 1 missing interface check, got %d\n%s", result.CountMissing, out)
		}
	})

	t.Run("GenTableTest", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-testing", "gen-table-test.sh")
		cmd := exec.Command("bash", script, "--parallel", "ParseDuration", "parser")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("gen-table-test.sh failed: %v\n%s", err, out)
		}

		if _, fmtErr := format.Source(out); fmtErr != nil {
			t.Errorf("generated Go code is not valid:\n%v\n%s", fmtErr, out)
		}

		if !strings.Contains(string(out), "t.Parallel()") {
			t.Error("--parallel flag did not produce t.Parallel() in output")
		}
	})

	t.Run("SetupLintDryRun", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-linting", "setup-lint.sh")
		cmd := exec.Command("bash", script, "--dry-run")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("setup-lint.sh --dry-run failed: %v\n%s", err, out)
		}
		s := string(out)
		if !strings.Contains(s, "linters:") {
			t.Error("dry-run output missing 'linters:' section")
		}
		if !strings.Contains(s, "linters-settings:") {
			t.Error("dry-run output missing 'linters-settings:' section")
		}
	})

	t.Run("PreReview", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-code-review", "pre-review.sh")
		cmd := exec.Command("bash", script, "--json", "--force", fixturesDir)
		out, _ := cmd.CombinedOutput()
		if !json.Valid(out) {
			t.Fatalf("--json output is not valid JSON:\n%s", out)
		}
		if !strings.Contains(string(out), "gofmt") {
			t.Error("JSON output missing gofmt section")
		}
	})

	t.Run("CheckDocsStrict", func(t *testing.T) {
		t.Parallel()
		script := scriptPath("go-documentation", "check-docs.sh")
		fixture := filepath.Join(fixturesDir, "unexported.go")
		cmd := exec.Command("bash", script, "--json", "--strict", fixture)
		out, _ := cmd.CombinedOutput()
		if !json.Valid(out) {
			t.Fatalf("--strict --json output is not valid JSON:\n%s", out)
		}
		var result struct {
			Total int `json:"total"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatalf("parse JSON: %v\n%s", err, out)
		}
		if result.Total < 1 {
			t.Errorf("--strict should find >= 1 undocumented unexported symbol, got %d\n%s", result.Total, out)
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
			if len(desc) > 1024 {
				t.Errorf("description is %d chars (max 1024)", len(desc))
			}

			bodyLines := strings.Count(body, "\n")
			if bodyLines >= 500 {
				t.Errorf("body is %d lines (spec recommends < 500)", bodyLines)
			}

			// Check shebang on all .sh files
			scriptDir := filepath.Join(dir, "scripts")
			if entries, err := os.ReadDir(scriptDir); err == nil {
				for _, e := range entries {
					if !strings.HasSuffix(e.Name(), ".sh") {
						continue
					}
					shContent, err := os.ReadFile(filepath.Join(scriptDir, e.Name()))
					if err != nil {
						t.Errorf("read %s: %v", e.Name(), err)
						continue
					}
					if !strings.HasPrefix(string(shContent), "#!/usr/bin/env bash") {
						t.Errorf("%s: missing #!/usr/bin/env bash shebang", e.Name())
					}
				}
			}
		})
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
