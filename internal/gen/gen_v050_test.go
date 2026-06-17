package gen

import (
	"strings"
	"testing"

	"github.com/reviewqa/reviewqa/internal/ast"
	"github.com/reviewqa/reviewqa/internal/plan"
)

func TestAPITemplate_EmitsExtendedNegatives_v050(t *testing.T) {
	form := &ast.FormSpec{
		Action:  "/api/contact",
		Method:  "post",
		EncType: "application/x-www-form-urlencoded",
		Inputs: []ast.FormInput{
			{Name: "email", Type: "email", Required: true},
			{Name: "message", Type: "textarea", Required: true},
		},
	}
	sym := ast.Symbol{Name: "ContactAPI", Kind: ast.KindComponent, File: "https://x.test/api/contact", Language: "ts"}
	it := plan.Item{
		Symbol:   sym,
		Symbols:  []ast.Symbol{sym},
		PageURL:  "https://x.test/api/contact",
		Template: plan.TmplPlaywrightAPI,
		OutPath:  "tests/e2e/api/x-test-api-contact.api.spec.ts",
		Form:     form,
	}
	out, err := Render([]plan.Item{it}, t.TempDir())
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	body := string(out[0].Content)
	for _, needle := range []string{
		"@kind:api @negative unicode payload",
		"@kind:api @negative sql-injection-shaped",
		"@kind:api @negative xss-shaped",
		"@kind:api @negative null-byte injection",
		"@kind:api @negative rapid burst",
		"DROP TABLE users",
		"window.__rqXSS=1",
		"value\\x00malicious",
		"Array.from({ length: 10 }",
	} {
		if !strings.Contains(body, needle) {
			t.Errorf("v0.50 API template missing %q", needle)
		}
	}
}

func TestAPITemplate_CountOfNegatives_v050(t *testing.T) {
	// At least 9 @kind:api @negative blocks (4 from v0.20 + 5 from v0.50).
	form := &ast.FormSpec{
		Action: "/api/contact", Method: "post",
		EncType: "application/x-www-form-urlencoded",
		Inputs:  []ast.FormInput{{Name: "email", Type: "email", Required: true}},
	}
	sym := ast.Symbol{Name: "X", Kind: ast.KindComponent, File: "https://x.test/", Language: "ts"}
	it := plan.Item{
		Symbol: sym, Symbols: []ast.Symbol{sym},
		PageURL: "https://x.test/api/contact",
		Template: plan.TmplPlaywrightAPI,
		OutPath: "tests/e2e/api/x.api.spec.ts",
		Form: form,
	}
	out, err := Render([]plan.Item{it}, t.TempDir())
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	body := string(out[0].Content)
	count := strings.Count(body, "@kind:api @negative")
	if count < 9 {
		t.Errorf("expected ≥9 @kind:api @negative blocks (4 v0.20 + 5 v0.50); got %d", count)
	}
}
