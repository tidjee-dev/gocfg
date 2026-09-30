// Package dotenv parses .env files (godotenv-compatible behavior, stdlib only).
//
// Documented deltas from joho/godotenv: keys are ASCII `[A-Za-z_][A-Za-z0-9_.-]*`
// (no leading digit, no `:` separator); `${VAR}` expansion prefers OS env over
// earlier file keys (matching OS > .env precedence); `\t` maps to tab.
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
		contQuote byte
		inML      bool
	)

	flush := func(lineNo int) error {
		raw := strings.Join(contLines, "\n")
		// The closer is the first unescaped quote; anything after it
		// must be empty or a comment.
		idx := indexUnescaped(contQuote, raw)
		if idx < 0 {
			return &ParseError{Line: contStart, Msg: "unterminated quoted value"}
		}
		inner := raw[:idx]
		trailer := strings.TrimSpace(raw[idx+1:])
		if trailer != "" && !strings.HasPrefix(trailer, "#") {
			return &ParseError{Line: lineNo, Msg: "unexpected content after closing quote"}
		}
		if contQuote == '"' {
			inner = unescapeDouble(inner)
			inner = expand(inner, out)
		}
		// Single-quoted values stay literal: no escapes, no expansion.
		out[contKey] = inner
		inML = false
		contLines = nil
		return nil
	}

	for i, line := range lines {
		no := i + 1
		if inML {
			contLines = append(contLines, line)
			if isClosedMultiline(contQuote, contLines) {
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
		// export prefix (only when followed by space/tab, so `exported=1` is a key)
		if rest, ok := cutExportPrefix(t); ok {
			t = strings.TrimSpace(rest)
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
			body := rawVal[1:]
			if idx := indexUnescaped('"', body); idx >= 0 {
				// Single-line (may carry a trailing comment).
				trailer := strings.TrimSpace(body[idx+1:])
				if trailer != "" && !strings.HasPrefix(trailer, "#") {
					return nil, &ParseError{Line: no, Msg: "unexpected content after closing quote"}
				}
				out[key] = expand(unescapeDouble(body[:idx]), out)
			} else {
				// Start multiline; stays open until first unescaped quote.
				inML = true
				contQuote = '"'
				contKey = key
				contStart = no
				contLines = []string{body}
			}
		case '\'':
			body := rawVal[1:]
			if idx := indexUnescaped('\'', body); idx >= 0 {
				trailer := strings.TrimSpace(body[idx+1:])
				if trailer != "" && !strings.HasPrefix(trailer, "#") {
					return nil, &ParseError{Line: no, Msg: "unexpected content after closing quote"}
				}
				out[key] = body[:idx] // no escape, no expansion in single quotes
			} else {
				// Start single-quoted multiline (literal, like godotenv).
				inML = true
				contQuote = '\''
				contKey = key
				contStart = no
				contLines = []string{body}
			}
		default:
			// unquoted: strip inline comment " #..."
			v := stripInlineComment(rawVal)
			out[key] = expand(strings.TrimSpace(v), out)
		}
	}
	if inML {
		return nil, &ParseError{Line: contStart, Msg: "unterminated quoted value"}
	}
	return out, nil
}

func cutExportPrefix(s string) (string, bool) {
	if !strings.HasPrefix(s, "export") {
		return s, false
	}
	rest := s[len("export"):]
	if rest == "" || rest[0] != ' ' && rest[0] != '\t' {
		return s, false
	}
	return rest, true
}

func validKey(k string) bool {
	if k == "" {
		return false
	}
	for i, c := range k {
		switch {
		case c == '_' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z':
		case i > 0 && (c >= '0' && c <= '9' || c == '-' || c == '.'):
		default:
			return false
		}
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

// indexUnescaped returns the byte index of the first unescaped
// occurrence of q in s (a quote preceded by an odd number of
// backslashes does not count), or -1 if there is none.
func indexUnescaped(q byte, s string) int {
	esc := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if esc {
			esc = false
			continue
		}
		if c == '\\' {
			esc = true
			continue
		}
		if c == q {
			return i
		}
	}
	return -1
}

func isClosedMultiline(q byte, parts []string) bool {
	return indexUnescaped(q, strings.Join(parts, "\n")) >= 0
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
