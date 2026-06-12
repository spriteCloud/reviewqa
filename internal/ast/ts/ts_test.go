package ts

import (
	"testing"

	"github.com/reviewqa/reviewqa/internal/ast"
)

func TestMultiLineExportFunction(t *testing.T) {
	src := []byte(`export async function loadUser(
  id: number,
  opts?: { trace?: boolean },
): Promise<User> {
  return null as any
}
`)
	syms, _ := New().Extract("src/api.ts", src)
	if len(syms) != 1 {
		t.Fatalf("want 1 sym, got %d: %+v", len(syms), syms)
	}
	s := syms[0]
	if s.Name != "loadUser" || s.Kind != ast.KindFunction {
		t.Errorf("multi-line export mismatch: %+v", s)
	}
	if len(s.Params) != 2 {
		t.Errorf("expected 2 params, got %d: %+v", len(s.Params), s.Params)
	}
}

func TestAngularComponentDecorator(t *testing.T) {
	src := []byte(`import { Component } from '@angular/core'

@Component({
  selector: 'app-foo',
  template: '<button (click)="onClick()"></button>',
})
export class FooComponent {
  onClick() { return 1 }
}
`)
	syms, _ := New().Extract("src/foo.component.ts", src)
	var comp, method *ast.Symbol
	for i := range syms {
		s := &syms[i]
		if s.Kind == ast.KindComponent && s.Name == "FooComponent" {
			comp = s
		}
		if s.Kind == ast.KindMethod && s.Name == "onClick" {
			method = s
		}
	}
	if comp == nil {
		t.Errorf("missing component symbol; got %+v", syms)
	} else if comp.FrameworkHint != "angular" {
		t.Errorf("framework hint: %q", comp.FrameworkHint)
	}
	if method == nil {
		t.Errorf("missing method symbol; got %+v", syms)
	} else if method.Receiver != "FooComponent" {
		t.Errorf("receiver: %q", method.Receiver)
	}
}

func TestNestController(t *testing.T) {
	src := []byte(`import { Controller, Get } from '@nestjs/common'

@Controller('/users')
export class UsersController {
  @Get('list')
  list() { return [] }
}
`)
	syms, _ := New().Extract("src/users.controller.ts", src)
	var found bool
	for _, s := range syms {
		if s.Kind == ast.KindRoute && s.Name == "list" {
			found = true
			if s.Path != "/users/list" {
				t.Errorf("path: %q", s.Path)
			}
			if s.Method != "GET" {
				t.Errorf("method: %q", s.Method)
			}
		}
	}
	if !found {
		t.Errorf("missing route symbol; got %+v", syms)
	}
}
