package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
)

type LoggerOptions struct {
	Level  slog.Level
	JSON   bool
	Writer io.Writer
}

type prettyWriter struct {
	target io.Writer
	format string
}

func (w *prettyWriter) Write(p []byte) (int, error) {
	switch w.format {
	case "json":
		var out bytes.Buffer

		// Use json.Decoder to decode tokens in order
		decoder := json.NewDecoder(bytes.NewReader(p))
		var orderedFields []struct {
			Key   string
			Value any
		}

		// Expect start of object
		t, err := decoder.Token()
		if err != nil || t != json.Delim('{') {
			return w.target.Write(p)
		}

		for decoder.More() {
			// Read key
			keyToken, err := decoder.Token()
			if err != nil {
				return w.target.Write(p)
			}
			key := keyToken.(string)

			// Read value
			var val any
			if err := decoder.Decode(&val); err != nil {
				return w.target.Write(p)
			}

			orderedFields = append(orderedFields, struct {
				Key   string
				Value any
			}{Key: key, Value: val})
		}

		// Expect end of object
		if _, err := decoder.Token(); err != nil {
			return w.target.Write(p)
		}

		// Write manually with indent
		out.WriteString("{\n")
		for i, kv := range orderedFields {
			v, _ := json.MarshalIndent(kv.Value, "  ", "  ")
			fmt.Fprintf(&out, "  %q: %s", kv.Key, string(v))
			if i < len(orderedFields)-1 {
				out.WriteString(",")
			}
			out.WriteString("\n")
		}
		out.WriteString("}\n")

		return w.target.Write(out.Bytes())

	case "text":
		var out bytes.Buffer
		var buf bytes.Buffer
		inQuotes := false
		escapeNext := false

		for i := range p {
			ch := p[i]

			if escapeNext {
				buf.WriteByte(ch)
				escapeNext = false
				continue
			}

			switch ch {
			case '\\':
				escapeNext = true
			case '"':
				inQuotes = !inQuotes
				buf.WriteByte(ch)
			case ' ':
				if inQuotes {
					buf.WriteByte(ch)
				} else {
					// write complete field and newline
					out.WriteString("  ")
					out.Write(buf.Bytes())
					out.WriteByte('\n')
					buf.Reset()
				}
			default:
				buf.WriteByte(ch)
			}
		}

		// final field
		if buf.Len() > 0 {
			out.WriteString("  ")
			out.Write(buf.Bytes())
			out.WriteByte('\n')
		}

		return w.target.Write(out.Bytes())

	default:
		return w.target.Write(p)
	}
}

func NewLogger(opts LoggerOptions) *slog.Logger {
	prettyWriter := &prettyWriter{
		target: opts.Writer,
		format: map[bool]string{
			true:  "json",
			false: "text",
		}[opts.JSON],
	}

	var handler slog.Handler
	if opts.JSON {
		handler = slog.NewJSONHandler(prettyWriter, &slog.HandlerOptions{
			Level: opts.Level,
		})
	} else {
		handler = slog.NewTextHandler(prettyWriter, &slog.HandlerOptions{
			Level: opts.Level,
		})
	}
	return slog.New(handler)
}
