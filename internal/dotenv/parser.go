// Package dotenv parses .env files (godotenv-compatible behavior, stdlib only).
package dotenv

import (
	"fmt"
	"os"
	"strings"
)

// ParseError reports a malformed line.
type ParseError struct {
	Line int
	Msg  string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("dotenv: line %d: %s", e.Line, e.Msg)
}

// Parse parses src into a map. Duplicate keys: last wins.
// Does not touch OS env. Use Load to apply with OS>file precedence.
func Parse(src []byte) (map[string]string, error) {
	out := make(map[string]string)
	// Normalize line endings, keep logical lines for multiline "...".
	text := strings.ReplaceAll(string(src), "\r\n", "\n")
	lines := strings.Split(text, "\n")

	var (
		contKey   string
		contLines []string
		contStart int
		inML      bool
	)

	flush := func(lineNo int) error {
		raw := strings.Join(contLines, "\n")
		// strip closing quote on last line
		idx := strings.LastIndex(raw, `"`)
		if idx < 0 {
			return &ParseError{Line: contStart, Msg: "unterminated double-quoted value"}
		}
		inner := raw[:idx]
		trailer := strings.TrimSpace(raw[idx+1:])
		if trailer != "" && !strings.HasPrefix(trailer, "#") {
			return &ParseError{Line: lineNo, Msg: "unexpected content after closing quote"}
		}
		v := unescapeDouble(inner)
		out[contKey] = expand(v, out)
		inML = false
		contLines = nil
		return nil
	}

	for i, line := range lines {
		no := i + 1
		if inML {
			contLines = append(contLines, line)
			if isClosedMultiline(contLines) {
				if err := flush(no); err != nil {
					return nil, err
				}
			}
			continue
		}

		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		// export prefix
		if strings.HasPrefix(t, "export ") || strings.HasPrefix(t, "export\t") {
			t = strings.TrimSpace(strings.TrimPrefix(t, "export"))
			// TrimPrefix with "export" leaves rest; re-trim:
			t = strings.TrimSpace(trimExportPrefix(line))
		}
		eq := strings.Index(t, "=")
		if eq < 0 {
			return nil, &ParseError{Line: no, Msg: fmt.Sprintf("missing '=': %q", line)}
		}
		key := strings.TrimSpace(t[:eq])
		if !validKey(key) {
			return nil, &ParseError{Line: no, Msg: fmt.Sprintf("invalid key %q", key)}
		}
		rawVal := strings.TrimSpace(t[eq+1:])

		// empty
		if rawVal == "" || strings.HasPrefix(rawVal, "#") {
			out[key] = ""
			continue
		}
		switch rawVal[0] {
		case '"':
			if len(rawVal) >= 2 && strings.HasSuffix(rawVal, `"`) && !strings.HasSuffix(rawVal, `\"`) && countUnescaped(rawVal) >= 2 {
				inner := rawVal[1 : len(rawVal)-1]
				out[key] = expand(unescapeDouble(inner), out)
			} else {
				// start multiline
				inML = true
				contKey = key
				contStart = no
				contLines = []string{rawVal[1:]} // after opening quote
				// single-line "abc (unterminated on same line) stays open
			}
		case '\'':
			end := strings.Index(rawVal[1:], "'")
			if end < 0 {
				return nil, &ParseError{Line: no, Msg: "unterminated single-quoted value"}
			}
			v := rawVal[1 : 1+end]
			trailer := strings.TrimSpace(rawVal[1+end+1:])
			if trailer != "" && !strings.HasPrefix(trailer, "#") {
				return nil, &ParseError{Line: no, Msg: "unexpected content after closing quote"}
			}
			out[key] = v // no escape, no expansion in single quotes
		default:
			// unquoted: strip inline comment " #..."
			v := stripInlineComment(rawVal)
			out[key] = expand(strings.TrimSpace(v), out)
		}
	}
	if inML {
		return nil, &ParseError{Line: contStart, Msg: "unterminated double-quoted value"}
	}
	return out, nil
}

func trimExportPrefix(line string) string {
	t := strings.TrimSpace(line)
	if strings.HasPrefix(t, "export") {
		return strings.TrimSpace(t[len("export"):])
	}
	return t
}

func validKey(k string) bool {
	if k == "" {
		return false
	}
	for i, c := range k {
		if c == '_' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || (i > 0 && c >= '0' && c <= '9') {
			continue
		}
		return false
	}
	return true
}

func stripInlineComment(s string) string {
	// cut at " #" outside quotes (value here is already unquoted)
	if idx := strings.Index(s, " #"); idx >= 0 {
		return s[:idx]
	}
	if strings.HasPrefix(s, "#") {
		return ""
	}
	return s
}

func unescapeDouble(s string) string {
	r := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\"`, `"`, `\\`, "\\")
	return r.Replace(s)
}

func countUnescaped(s string) int {
	n := 0
	esc := false
	for _, c := range s {
		if esc {
			esc = false
			continue
		}
		if c == '\\' {
			esc = true
			continue
		}
		if c == '"' {
			n++
		}
	}
	return n
}

func isClosedMultiline(parts []string) bool {
	joined := strings.Join(parts, "\n")
	// odd number of unescaped trailing quotes on last segment closes it
	esc := false
	quotes := 0
	for _, c := range joined {
		if esc {
			esc = false
			continue
		}
		if c == '\\' {
			esc = true
			continue
		}
		if c == '"' {
			quotes++
		}
	}
	// first line's opening quote already consumed, so we need >=1 close
	return quotes >= 1
}

// expand supports $VAR and ${VAR} against OS env first, then already-parsed keys.
func expand(s string, parsed map[string]string) string {
	return os.Expand(s, func(k string) string {
		if v, ok := os.LookupEnv(k); ok {
			return v
		}
		if v, ok := parsed[k]; ok {
			return v
		}
		return ""
	})
}
