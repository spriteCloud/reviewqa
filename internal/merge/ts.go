package merge

import (
	"bytes"
	"regexp"
	"strings"
)

var tsImportRE = regexp.MustCompile(`(?m)^import\s+.*?from\s+['"][^'"]+['"];?\s*$`)

func appendTS(existing, generated []byte) ([]byte, bool) {
	oldImports, oldRest := splitTSImports(existing)
	newImports, newRest := splitTSImports(generated)

	have := map[string]bool{}
	for _, imp := range oldImports {
		have[strings.TrimSpace(imp)] = true
	}
	var add []string
	for _, imp := range newImports {
		k := strings.TrimSpace(imp)
		if !have[k] {
			have[k] = true
			add = append(add, imp)
		}
	}

	var buf bytes.Buffer
	for _, imp := range oldImports {
		buf.WriteString(imp)
		buf.WriteByte('\n')
	}
	for _, imp := range add {
		buf.WriteString(imp)
		buf.WriteByte('\n')
	}
	if buf.Len() > 0 {
		buf.WriteByte('\n')
	}
	buf.WriteString(strings.TrimRight(oldRest, "\n"))
	buf.WriteString("\n\n")
	buf.WriteString(strings.TrimLeft(newRest, "\n"))
	if !bytes.HasSuffix(buf.Bytes(), []byte("\n")) {
		buf.WriteByte('\n')
	}
	return buf.Bytes(), true
}

func splitTSImports(src []byte) (imports []string, rest string) {
	matches := tsImportRE.FindAllIndex(src, -1)
	if len(matches) == 0 {
		return nil, string(src)
	}
	var b strings.Builder
	last := 0
	for _, m := range matches {
		imports = append(imports, string(bytes.TrimRight(src[m[0]:m[1]], "\r\n ")))
		b.Write(src[last:m[0]])
		last = m[1]
		if last < len(src) && src[last] == '\n' {
			last++
		}
	}
	b.Write(src[last:])
	return imports, b.String()
}
