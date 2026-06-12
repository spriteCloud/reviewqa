package gen

import (
	"strings"
	"testing"

	"github.com/reviewqa/reviewqa/internal/ast"
	"github.com/reviewqa/reviewqa/internal/plan"
)

func TestRenderJestUnit(t *testing.T) {
	items := []plan.Item{{
		Symbol: ast.Symbol{
			Kind: ast.KindFunction, Name: "add", File: "src/math.ts", Language: "ts",
			Params: []ast.Param{{Name: "a", Type: "number"}, {Name: "b", Type: "number"}},
		},
		Template: plan.TmplJestUnit,
		OutPath:  "src/__tests__/math.test.ts",
	}}
	out, err := Render(items, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("got %d", len(out))
	}
	body := string(out[0].Content)
	for _, want := range []string{"import { add }", "../math", "describe('add'", "add(0, 0)"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
}

func TestRenderPytestAPI(t *testing.T) {
	items := []plan.Item{{
		Symbol: ast.Symbol{
			Kind: ast.KindRoute, Name: "get_user", Method: "GET", Path: "/users/{uid}",
			File: "app/users.py", Language: "python", FrameworkHint: "fastapi",
		},
		Template: plan.TmplPytestAPI,
		OutPath:  "tests/test_users.py",
	}}
	out, err := Render(items, ".")
	if err != nil {
		t.Fatal(err)
	}
	body := string(out[0].Content)
	for _, want := range []string{"from fastapi.testclient", "TestClient(app)", `client.get("/users/{uid}")`, "test_get_get_user"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
}

func TestRenderGoHTTPTest(t *testing.T) {
	items := []plan.Item{{
		Symbol: ast.Symbol{
			Kind: ast.KindRoute, Name: "Health", File: "server/handlers.go", Language: "go",
		},
		Template: plan.TmplGoHTTPTest,
		OutPath:  "server/handlers_test.go",
	}}
	out, err := Render(items, ".")
	if err != nil {
		t.Fatal(err)
	}
	body := string(out[0].Content)
	for _, want := range []string{"package server", "httptest.NewRequest", "Health(rr, req)"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
}

func TestRenderJUnit5RestAssured(t *testing.T) {
	items := []plan.Item{{
		Symbol: ast.Symbol{
			Kind: ast.KindRoute, Name: "getById", Receiver: "UserController",
			Method: "GET", Path: "/users/{id}", File: "src/main/java/com/acme/UserController.java",
			Language: "java",
		},
		Template: plan.TmplJUnit5RestAssured,
		OutPath:  "src/test/java/com/acme/UserControllerTest.java",
	}}
	out, err := Render(items, ".")
	if err != nil {
		t.Fatal(err)
	}
	body := string(out[0].Content)
	for _, want := range []string{"package com.acme;", "RestAssured", `get("/users/{id}")`, "class UserControllerTest"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
}
