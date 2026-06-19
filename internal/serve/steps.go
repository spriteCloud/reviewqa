package serve

import (
	"os"
	"regexp"
	"strings"
)

// StepPattern is one entry from the generated quail.steps.ts file —
// the bound Gherkin pattern + its keyword family. The UI surfaces this
// list so users editing a Scenario can pick from the registered
// vocabulary instead of guessing.
type StepPattern struct {
	Keyword string `json:"keyword"`           // Given | When | Then
	Pattern string `json:"pattern"`           // the literal or regex source
	IsRegex bool   `json:"isRegex,omitempty"` // true when the binding used /…/ instead of '…'
}

// ParseStepsFile reads a playwright-bdd step-defs file (TypeScript) and
// extracts the registered patterns. The parser is regex-based — it does
// NOT walk a TS AST. It looks for top-level `Given(…)`, `When(…)`,
// `Then(…)` invocations whose first argument is either a string literal
// or a regex literal.
func ParseStepsFile(path string) ([]StepPattern, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseStepsBytes(b), nil
}

// ParseStepsBytes is the in-memory variant.
func ParseStepsBytes(b []byte) []StepPattern {
	src := string(b)
	out := make([]StepPattern, 0, 32)
	for _, m := range stepCallRe.FindAllStringSubmatchIndex(src, -1) {
		kw := src[m[2]:m[3]]
		argStart := m[1]
		pat, isRegex, ok := readFirstArg(src, argStart)
		if !ok {
			continue
		}
		out = append(out, StepPattern{Keyword: kw, Pattern: strings.TrimSpace(pat), IsRegex: isRegex})
	}
	return out
}

// stepCallRe matches the start of a Given/When/Then call. The first
// submatch is the keyword. The regex ends at the opening paren so the
// argument reader below picks up wherever this match ends.
// Only column-0 calls are matched. playwright-bdd registers patterns
// at module top level; anything indented is inside a function body and
// not a real registration.
var stepCallRe = regexp.MustCompile(`(?m)^(Given|When|Then)\s*\(`)

// readFirstArg consumes the first argument of a Given/When/Then call
// starting at the byte AFTER the opening paren. Returns the pattern
// body (without surrounding quotes / slashes), whether it was a regex
// literal, and a success flag. Quote-aware so embedded escapes don't
// confuse it.
func readFirstArg(src string, start int) (string, bool, bool) {
	// Skip whitespace.
	for start < len(src) && (src[start] == ' ' || src[start] == '\t' || src[start] == '\n' || src[start] == '\r') {
		start++
	}
	if start >= len(src) {
		return "", false, false
	}
	switch src[start] {
	case '\'', '"', '`':
		quote := src[start]
		i := start + 1
		var b strings.Builder
		for i < len(src) {
			c := src[i]
			if c == '\\' && i+1 < len(src) {
				b.WriteByte(src[i+1])
				i += 2
				continue
			}
			if c == quote {
				return b.String(), false, true
			}
			b.WriteByte(c)
			i++
		}
	case '/':
		i := start + 1
		var b strings.Builder
		for i < len(src) {
			c := src[i]
			if c == '\\' && i+1 < len(src) {
				b.WriteByte(c)
				b.WriteByte(src[i+1])
				i += 2
				continue
			}
			if c == '/' {
				return b.String(), true, true
			}
			b.WriteByte(c)
			i++
		}
	}
	return "", false, false
}
