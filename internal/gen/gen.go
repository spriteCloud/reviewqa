// Package gen renders deterministic test scaffolds from plan.Items using
// Go text/template. Templates are embedded into the binary so the CLI ships
// as a single artifact with no runtime asset path.
package gen

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/reviewqa/reviewqa/internal/ast"
	"github.com/reviewqa/reviewqa/internal/log"
	"github.com/reviewqa/reviewqa/internal/plan"
)

//go:embed all:templates
var templatesFS embed.FS

// Mount point: we embed via a re-export below; the embed directive can't
// reach ../../templates, so we duplicate-link at build time via a //go:embed
// in the wrapper file.

type Rendered struct {
	Path    string
	Content []byte
	Symbol  ast.Symbol
}

func Render(items []plan.Item, workDir string) ([]Rendered, error) {
	var out []Rendered
	for _, it := range items {
		tmpl, err := load(it.Template)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", it.Template, err)
		}
		data := buildData(it, workDir)
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return nil, fmt.Errorf("render %s for %s: %w", it.Template, it.Symbol.Name, err)
		}
		out = append(out, Rendered{Path: it.OutPath, Content: buf.Bytes(), Symbol: it.Symbol})
		log.Debug("rendered scaffold", "template", it.Template, "symbol", it.Symbol.Name, "path", it.OutPath)
	}
	return out, nil
}

func load(t plan.Template) (*template.Template, error) {
	sub, file := templateLocation(t)
	body, err := templatesFS.ReadFile(path.Join("templates", sub, file))
	if err != nil {
		return nil, err
	}
	return template.New(string(t)).Funcs(funcs).Parse(string(body))
}

func templateLocation(t plan.Template) (string, string) {
	switch t {
	case plan.TmplJestUnit:
		return "ts", "jest_unit.tmpl"
	case plan.TmplJestAPI:
		return "ts", "jest_api.tmpl"
	case plan.TmplPlaywrightE2E:
		return "ts", "pw_e2e.tmpl"
	case plan.TmplPlaywrightHappyFlow:
		return "ts", "pw_happyflow.tmpl"
	case plan.TmplPytestUnit:
		return "py", "pytest_unit.tmpl"
	case plan.TmplPytestAPI:
		return "py", "pytest_api.tmpl"
	case plan.TmplGoUnit:
		return "go", "gotest_unit.tmpl"
	case plan.TmplGoHTTPTest:
		return "go", "gotest_httptest.tmpl"
	case plan.TmplJUnit5Unit:
		return "java", "junit5_unit.tmpl"
	case plan.TmplJUnit5RestAssured:
		return "java", "junit5_restassured.tmpl"
	}
	return "", ""
}

var funcs = template.FuncMap{
	"lower":     strings.ToLower,
	"upper":     strings.ToUpper,
	"hasPrefix": strings.HasPrefix,
	"firstClickable": func(as []ast.LocatorAnchor) []ast.LocatorAnchor {
		for _, a := range as {
			switch a.Tag {
			case "button", "summary", "a", "input":
				return []ast.LocatorAnchor{a}
			}
		}
		return nil
	},
	"locatorFor": func(a ast.LocatorAnchor) string {
		switch {
		case a.TestID != "":
			return fmt.Sprintf("getByTestId('%s')", a.TestID)
		case a.Aria != "":
			return fmt.Sprintf("getByLabel('%s')", a.Aria)
		case a.Role != "":
			return fmt.Sprintf("getByRole('%s')", a.Role)
		}
		return "locator('body')"
	},
	"anchorLabel": func(a ast.LocatorAnchor) string {
		switch {
		case a.TestID != "":
			return a.TestID
		case a.Aria != "":
			return a.Aria
		case a.Role != "":
			return a.Role
		}
		return "element"
	},
	"isPrimitiveType": func(t string) bool {
		switch strings.TrimSpace(t) {
		case "int", "long", "short", "byte", "double", "float", "boolean", "char":
			return true
		}
		return false
	},
	"defaultForType": func(t string) string {
		switch strings.TrimSpace(t) {
		case "int", "long", "short", "byte", "double", "float":
			return "0"
		case "boolean":
			return "false"
		case "char":
			return "'\\0'"
		case "string", "String":
			return "\"\""
		}
		return "null"
	},
}

type renderData struct {
	Symbol          ast.Symbol
	Symbols         []ast.Symbol // populated for happy-flow; first == Symbol
	PageURL         string       // populated for happy-flow; "/" default
	ImportPath      string
	AppImportPath   string
	SupertestMethod string
	HappyArgs       string
	SnakeName       string
	Package         string
}

func buildData(it plan.Item, workDir string) renderData {
	d := renderData{Symbol: it.Symbol}
	d.Symbols = it.Symbols
	if len(d.Symbols) == 0 {
		d.Symbols = []ast.Symbol{it.Symbol}
	}
	d.PageURL = it.PageURL
	if d.PageURL == "" {
		d.PageURL = "/"
	}
	d.HappyArgs = happyArgs(it.Symbol)
	d.SnakeName = toSnake(it.Symbol.Name)
	d.SupertestMethod = strings.ToLower(it.Symbol.Method)
	switch it.Symbol.Language {
	case "ts":
		d.ImportPath = relativeImport(it.OutPath, it.Symbol.File)
		d.AppImportPath = relativeImport(it.OutPath, deriveAppEntry(workDir, it.Symbol.File))
	case "python":
		d.ImportPath = pythonModule(it.Symbol.File)
		d.AppImportPath = pythonModule(deriveAppEntry(workDir, it.Symbol.File))
	case "go":
		d.Package = goPackageFor(it.OutPath)
	case "java":
		d.Package = javaPackageFor(it.OutPath)
	}
	return d
}

func happyArgs(s ast.Symbol) string {
	parts := make([]string, 0, len(s.Params))
	for _, p := range s.Params {
		parts = append(parts, defaultForType(s.Language, p.Type))
	}
	return strings.Join(parts, ", ")
}

func defaultForType(lang, typ string) string {
	t := strings.ToLower(strings.TrimSpace(typ))
	switch lang {
	case "ts":
		switch {
		case t == "" || strings.Contains(t, "any") || strings.Contains(t, "unknown"):
			return "undefined"
		case strings.Contains(t, "number") || strings.Contains(t, "int") || strings.Contains(t, "float"):
			return "0"
		case strings.Contains(t, "string"):
			return `''`
		case strings.Contains(t, "bool"):
			return "false"
		case strings.HasPrefix(t, "array<") || strings.HasSuffix(t, "[]"):
			return "[]"
		default:
			return "undefined"
		}
	case "python":
		switch {
		case strings.Contains(t, "int"):
			return "0"
		case strings.Contains(t, "float"):
			return "0.0"
		case strings.Contains(t, "str"):
			return `""`
		case strings.Contains(t, "bool"):
			return "False"
		case strings.Contains(t, "list"):
			return "[]"
		case strings.Contains(t, "dict"):
			return "{}"
		default:
			return "None"
		}
	case "go":
		switch t {
		case "string":
			return `""`
		case "int", "int32", "int64", "uint", "uint32", "uint64", "byte", "rune":
			return "0"
		case "float32", "float64":
			return "0"
		case "bool":
			return "false"
		default:
			return "nil"
		}
	case "java":
		switch {
		case t == "int" || t == "long" || t == "short" || t == "byte":
			return "0"
		case t == "double" || t == "float":
			return "0.0"
		case t == "boolean":
			return "false"
		case t == "string":
			return `""`
		default:
			return "null"
		}
	}
	return "null"
}

func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		if r >= 'A' && r <= 'Z' {
			r = r + ('a' - 'A')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func relativeImport(testFile, srcFile string) string {
	if srcFile == "" {
		return "../src"
	}
	rel, err := filepath.Rel(filepath.Dir(testFile), srcFile)
	if err != nil {
		return strings.TrimSuffix(srcFile, filepath.Ext(srcFile))
	}
	rel = strings.TrimSuffix(rel, filepath.Ext(rel))
	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, ".") {
		rel = "./" + rel
	}
	return rel
}

func pythonModule(srcFile string) string {
	rel := strings.TrimSuffix(filepath.ToSlash(srcFile), ".py")
	return strings.ReplaceAll(rel, "/", ".")
}

func deriveAppEntry(workDir, source string) string {
	for _, c := range []string{"src/app.ts", "src/app.js", "src/index.ts", "src/server.ts", "app/main.py", "main.py"} {
		if _, err := os.Stat(filepath.Join(workDir, c)); err == nil {
			return c
		}
	}
	return source
}

func goPackageFor(testPath string) string {
	parts := strings.Split(filepath.ToSlash(testPath), "/")
	if len(parts) < 2 {
		return "main"
	}
	return parts[len(parts)-2]
}

func javaPackageFor(testPath string) string {
	rel := strings.TrimPrefix(filepath.ToSlash(testPath), "src/test/java/")
	dir := path.Dir(rel)
	if dir == "." || dir == "" {
		return "tests"
	}
	return strings.ReplaceAll(dir, "/", ".")
}
