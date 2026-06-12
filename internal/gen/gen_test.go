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

func TestRenderPlaywrightE2E_Rich(t *testing.T) {
	items := []plan.Item{{
		Symbol: ast.Symbol{
			Kind: ast.KindComponent, Name: "FAQ",
			File: "web/src/components/FAQ.tsx", Language: "ts",
			Line: 1, EndLine: 10, FrameworkHint: "react",
			HasState: true, HasOnClick: true,
			Anchors: []ast.LocatorAnchor{
				{TestID: "faq-list", Tag: "div"},
				{TestID: "faq-toggle", Tag: "button"},
			},
		},
		Template: plan.TmplPlaywrightE2E,
		OutPath:  "tests/e2e/FAQ.spec.ts",
	}}
	out, err := Render(items, ".")
	if err != nil {
		t.Fatal(err)
	}
	body := string(out[0].Content)
	for _, want := range []string{
		"test.describe('FAQ'",
		"getByTestId('faq-list')",
		"getByTestId('faq-toggle')",
		"toBeVisible()",
		"target.click()",
		"aria-expanded",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
	if got := strings.Count(body, "  test("); got < 4 {
		t.Errorf("expected ≥4 test blocks, got %d:\n%s", got, body)
	}
}

func TestRenderPlaywrightE2E_Fallback(t *testing.T) {
	items := []plan.Item{{
		Symbol: ast.Symbol{
			Kind: ast.KindComponent, Name: "Empty",
			File: "web/src/components/Empty.tsx", Language: "ts",
			Line: 1, EndLine: 3, FrameworkHint: "react",
		},
		Template: plan.TmplPlaywrightE2E,
		OutPath:  "tests/e2e/Empty.spec.ts",
	}}
	out, err := Render(items, ".")
	if err != nil {
		t.Fatal(err)
	}
	body := string(out[0].Content)
	for _, want := range []string{"test.describe('Empty'", "renders the Empty component", "getByRole('main')"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
	if got := strings.Count(body, "  test("); got != 1 {
		t.Errorf("expected exactly 1 test block, got %d:\n%s", got, body)
	}
}

func TestRenderJestAPI_RichAssertions(t *testing.T) {
	items := []plan.Item{{
		Symbol: ast.Symbol{
			Kind: ast.KindRoute, Name: "getUser", Method: "GET", Path: "/users/:id",
			File: "src/routes/users.ts", Language: "ts",
		},
		Template: plan.TmplJestAPI,
		OutPath:  "src/routes/__tests__/users.test.ts",
	}}
	out, _ := Render(items, ".")
	body := string(out[0].Content)
	for _, want := range []string{"toBeGreaterThanOrEqual(200)", "application\\/json", "JSON.stringify(res.body)"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
}

func TestRenderPytestAPI_RichAssertions(t *testing.T) {
	items := []plan.Item{{
		Symbol: ast.Symbol{
			Kind: ast.KindRoute, Name: "list_items", Method: "GET", Path: "/items",
			File: "app/items.py", Language: "python", FrameworkHint: "fastapi",
		},
		Template: plan.TmplPytestAPI,
		OutPath:  "tests/test_items.py",
	}}
	out, _ := Render(items, ".")
	body := string(out[0].Content)
	for _, want := range []string{"200 <= res.status_code", "application/json", "body is not None"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
}

func TestRenderGoHTTPTest_RichAssertions(t *testing.T) {
	items := []plan.Item{{
		Symbol: ast.Symbol{
			Kind: ast.KindRoute, Name: "Ping", File: "server/handlers.go", Language: "go",
		},
		Template: plan.TmplGoHTTPTest,
		OutPath:  "server/handlers_test.go",
	}}
	out, _ := Render(items, ".")
	body := string(out[0].Content)
	for _, want := range []string{`t.Run("happy path 2xx"`, `t.Run("content-type header set"`, `t.Run("response body non-empty"`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
}

func TestRenderGoUnit_TypeSubtestWhenResultKnown(t *testing.T) {
	items := []plan.Item{{
		Symbol: ast.Symbol{
			Kind: ast.KindFunction, Name: "Add", File: "math/math.go", Language: "go",
			HasResult: true, PrimaryReturn: "int",
		},
		Template: plan.TmplGoUnit,
		OutPath:  "math/math_test.go",
	}}
	out, _ := Render(items, ".")
	body := string(out[0].Content)
	for _, want := range []string{`t.Run("happy path"`, `t.Run("returns expected type"`, `reflect.TypeOf(got)`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
}

func TestRenderJUnit5Unit_DoesNotThrow(t *testing.T) {
	items := []plan.Item{{
		Symbol: ast.Symbol{
			Kind: ast.KindMethod, Name: "compute", Receiver: "Calc", Returns: "int",
			File: "src/main/java/com/acme/Calc.java", Language: "java",
		},
		Template: plan.TmplJUnit5Unit,
		OutPath:  "src/test/java/com/acme/CalcTest.java",
	}}
	out, _ := Render(items, ".")
	body := string(out[0].Content)
	for _, want := range []string{"@DisplayName", "doesNotThrow", "assertNotEquals(0, result)"} {
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
	for _, want := range []string{"package com.acme;", "RestAssured", `get("/users/{id}")`, "class UserControllerTest", "contentType(ContentType.JSON)", "notNullValue()"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
}
