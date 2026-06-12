// Package ast defines a uniform Symbol shape and an Extractor registry
// keyed by file extension. Per-language extractors live in subpackages.
//
// Implementation note: v1 uses regex/heuristic extractors so the binary is
// pure Go and trivially cross-compiles. The interface is designed so each
// extractor can be swapped for a tree-sitter implementation later without
// touching plan/, gen/, or heal/.
package ast

import (
	"path/filepath"
	"strings"
)

type Kind string

const (
	KindFunction  Kind = "function"
	KindMethod    Kind = "method"
	KindRoute     Kind = "route"     // HTTP route handler
	KindComponent Kind = "component" // UI component (TSX/Vue/Svelte)
)

type Param struct{ Name, Type string }

type Symbol struct {
	Kind     Kind
	Name     string
	Receiver string
	Params   []Param
	Returns  string
	File     string
	Language string
	Line     int // 1-based start line in NEW file
	EndLine  int
	// FrameworkHint is a free-form tag the emitter consults
	// (e.g. "express", "fastify", "fastapi", "flask", "spring", "nethttp").
	FrameworkHint string
	// HTTP-shaped data when Kind == KindRoute.
	Method string
	Path   string
	// Return-shape hints used by the Go scaffold template. Only the golang
	// extractor populates these today; other languages may leave them zero.
	HasError      bool
	HasResult     bool
	PrimaryReturn string
}

type LocatorAnchor struct {
	Role   string
	Name   string
	TestID string
	Aria   string
	Text   string
	CSS    string
	File   string
	Line   int
}

type Extractor interface {
	Language() string
	Extract(file string, content []byte) ([]Symbol, []LocatorAnchor)
}

var registry = map[string]Extractor{}

func Register(exts []string, e Extractor) {
	for _, x := range exts {
		registry[strings.ToLower(x)] = e
	}
}

func ForFile(path string) Extractor {
	return registry[strings.ToLower(filepath.Ext(path))]
}

func Registered() []string {
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	return out
}
