package gen_test

// End-to-end exercise of the plan→gen pipeline against a synthetic in-memory
// PR diff. Network and GitHub are mocked out; the goal is to prove all four
// languages produce buildable test files for a representative happy-path
// symbol.

import (
	"strings"
	"testing"

	_ "github.com/reviewqa/reviewqa/internal/ast/golang"
	_ "github.com/reviewqa/reviewqa/internal/ast/java"
	_ "github.com/reviewqa/reviewqa/internal/ast/python"
	_ "github.com/reviewqa/reviewqa/internal/ast/ts"
	"github.com/reviewqa/reviewqa/internal/diff"
	"github.com/reviewqa/reviewqa/internal/gen"
	"github.com/reviewqa/reviewqa/internal/plan"
)

func TestEndToEnd_AllLanguages(t *testing.T) {
	files := []diff.File{
		{
			Path: "src/math.ts", Status: "modified", Added: []diff.Range{{Start: 1, End: 10}},
			NewBlob: "export function add(a: number, b: number) {\n  return a + b\n}\n",
		},
		{
			Path: "app/users.py", Status: "added", Added: []diff.Range{{Start: 1, End: 10}},
			NewBlob: "from fastapi import FastAPI\napp = FastAPI()\n\n@app.get(\"/u\")\ndef list_users():\n    return []\n",
		},
		{
			Path: "server/health.go", Status: "added", Added: []diff.Range{{Start: 1, End: 5}},
			NewBlob: "package server\nimport \"net/http\"\nfunc Health(w http.ResponseWriter, r *http.Request) { w.Write([]byte(\"ok\")) }\n",
		},
		{
			Path: "src/main/java/com/acme/UserController.java", Status: "added", Added: []diff.Range{{Start: 1, End: 20}},
			NewBlob: "package com.acme;\nimport org.springframework.web.bind.annotation.*;\n@RestController\npublic class UserController {\n    @GetMapping(\"/u\")\n    public String list() { return \"\"; }\n}\n",
		},
	}
	items := plan.Build(files, plan.Layout{HasMavenLayout: true})
	if len(items) < 4 {
		t.Fatalf("expected >=4 items, got %d: %+v", len(items), items)
	}
	rs, err := gen.Render(items, ".")
	if err != nil {
		t.Fatal(err)
	}
	gotByLang := map[string]string{}
	for _, r := range rs {
		gotByLang[r.Symbol.Language] = string(r.Content)
	}
	checks := map[string][]string{
		"ts":     {"describe(", "add(0, 0)", "it.each(", "'returns a value for arguments"},
		"python": {"def test_", "client.get", "application/json", "body is not None"},
		"go":     {"package server", "httptest.NewRequest", `t.Run("content-type header set"`, `t.Run("response body non-empty"`},
		"java":   {"package com.acme;", "RestAssured", "contentType(ContentType.JSON)", "notNullValue()"},
	}
	for lang, want := range checks {
		body, ok := gotByLang[lang]
		if !ok {
			t.Errorf("no rendered file for %s", lang)
			continue
		}
		for _, s := range want {
			if !strings.Contains(body, s) {
				t.Errorf("%s missing %q in:\n%s", lang, s, body)
			}
		}
	}
}
