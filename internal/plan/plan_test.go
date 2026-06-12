package plan

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/reviewqa/reviewqa/internal/ast/golang"
	_ "github.com/reviewqa/reviewqa/internal/ast/java"
	_ "github.com/reviewqa/reviewqa/internal/ast/python"
	_ "github.com/reviewqa/reviewqa/internal/ast/ts"
	"github.com/reviewqa/reviewqa/internal/diff"
)

func TestBuildPicksTemplatesPerLanguage(t *testing.T) {
	dir := t.TempDir()
	must := func(rel, body string) string {
		full := filepath.Join(dir, rel)
		os.MkdirAll(filepath.Dir(full), 0o755)
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return rel
	}
	tsPath := must("src/math.ts", "export function add(a: number, b: number) { return a + b }\n")
	pyPath := must("app/users.py", "from fastapi import FastAPI\napp = FastAPI()\n\n@app.get(\"/u\")\ndef list_users():\n    return []\n")
	goPath := must("server/health.go", "package server\nimport \"net/http\"\nfunc Health(w http.ResponseWriter, r *http.Request) { w.Write([]byte(\"ok\")) }\n")
	javaPath := must("src/main/java/com/acme/UserController.java", "package com.acme;\nimport org.springframework.web.bind.annotation.*;\n@RestController\npublic class UserController {\n    @GetMapping(\"/u\")\n    public String list() { return \"\"; }\n}\n")

	files := []diff.File{
		{Path: tsPath, Added: []diff.Range{{Start: 1, End: 5}}},
		{Path: pyPath, Added: []diff.Range{{Start: 1, End: 10}}},
		{Path: goPath, Added: []diff.Range{{Start: 1, End: 10}}},
		{Path: javaPath, Added: []diff.Range{{Start: 1, End: 20}}, Status: "added"},
	}
	items := Build(files, Detect(dir))
	var byLang = map[string][]Item{}
	for _, it := range items {
		byLang[it.Symbol.Language] = append(byLang[it.Symbol.Language], it)
	}
	if len(byLang["ts"]) == 0 || byLang["ts"][0].Template != TmplJestUnit {
		t.Errorf("ts: %+v", byLang["ts"])
	}
	if len(byLang["python"]) == 0 || byLang["python"][0].Template != TmplPytestAPI {
		t.Errorf("python: %+v", byLang["python"])
	}
	if len(byLang["go"]) == 0 || byLang["go"][0].Template != TmplGoHTTPTest {
		t.Errorf("go: %+v", byLang["go"])
	}
	if len(byLang["java"]) == 0 || byLang["java"][0].Template != TmplJUnit5RestAssured {
		t.Errorf("java: %+v", byLang["java"])
	}
}

func TestLayoutDetection(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "__tests__"), 0o755)
	os.MkdirAll(filepath.Join(dir, "src", "test", "java"), 0o755)
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"devDependencies":{"vitest":"^1"}}`), 0o644)
	l := Detect(dir)
	if !l.HasJestDir {
		t.Error("expected HasJestDir")
	}
	if !l.UsesVitest {
		t.Error("expected UsesVitest")
	}
	if !l.HasMavenLayout {
		t.Error("expected HasMavenLayout")
	}
}
