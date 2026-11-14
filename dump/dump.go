package dump

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"time"
)

func Dump(v interface{}) string {
	buf := bytes.Buffer{}
	p := &printer{
		writer:   &buf,
		indent:   "  ",
		visited:  make(map[uintptr]bool),
		maxDepth: 10,
	}
	p.format(reflect.ValueOf(v), 0)
	fmt.Fprintln(p.writer)
	return buf.String()
}

type printer struct {
	writer   io.Writer
	indent   string
	visited  map[uintptr]bool
	maxDepth int
}

func (p *printer) format(val reflect.Value, depth int) {
	if depth > p.maxDepth {
		fmt.Fprint(p.writer, "<max depth reached>")
		return
	}

	if !val.IsValid() {
		fmt.Fprint(p.writer, "nil")
		return
	}

	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			fmt.Fprint(p.writer, "nil")
			return
		}
		// Handle cyclic references for pointers
		ptr := val.Pointer()
		if p.visited[ptr] {
			fmt.Fprintf(p.writer, "&<cyclic %s>", val.Type().Elem())
			return
		}
		p.visited[ptr] = true
		defer delete(p.visited, ptr)

		fmt.Fprint(p.writer, "&")
		p.format(val.Elem(), depth)

	case reflect.Interface:
		if val.IsNil() {
			fmt.Fprint(p.writer, "nil")
			return
		}
		p.format(val.Elem(), depth)

	case reflect.Struct:
		t := val.Type()

		if t == reflect.TypeFor[time.Time]() {
			fmt.Fprintf(p.writer, "%v", val.Interface().(time.Time).Format(time.RFC3339))
			return
		}

		fmt.Fprintf(p.writer, "%s {\n", t.Name())
		for i := 0; i < val.NumField(); i++ {
			fieldVal := val.Field(i)
			fieldType := t.Field(i)

			// Skip unexported fields or handle them gracefully
			if fieldType.PkgPath != "" && !fieldType.Anonymous {
				continue
			}

			p.writeIndent(depth + 1)
			fmt.Fprintf(p.writer, "%s: ", fieldType.Name)
			p.format(fieldVal, depth+1)
			fmt.Fprint(p.writer, ",\n")
		}
		p.writeIndent(depth)
		fmt.Fprint(p.writer, "}")

	case reflect.Map:
		if val.IsNil() {
			fmt.Fprint(p.writer, "nil")
			return
		}
		t := val.Type()
		fmt.Fprintf(p.writer, "map[%s][%s]{\n", t.Key(), t.Elem())

		// Sort keys for deterministic output
		keys := val.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprintf("%v", keys[i].Interface()) < fmt.Sprintf("%v", keys[j].Interface())
		})

		for _, k := range keys {
			p.writeIndent(depth + 1)
			p.format(k, depth+1)
			fmt.Fprint(p.writer, ": ")
			p.format(val.MapIndex(k), depth+1)
			fmt.Fprint(p.writer, ",\n")
		}
		p.writeIndent(depth)
		fmt.Fprint(p.writer, "}")

	case reflect.Slice, reflect.Array:
		if val.Kind() == reflect.Slice && val.IsNil() {
			fmt.Fprint(p.writer, "nil")
			return
		}
		fmt.Fprintf(p.writer, "%s{\n", val.Type())
		for i := 0; i < val.Len(); i++ {
			p.writeIndent(depth + 1)
			p.format(val.Index(i), depth+1)
			fmt.Fprint(p.writer, ",\n")
		}
		p.writeIndent(depth)
		fmt.Fprint(p.writer, "}")

	case reflect.String:
		fmt.Fprintf(p.writer, "%q", val.String())

	default:
		fmt.Fprintf(p.writer, "%v", val.Interface())
	}
}

func (p *printer) writeIndent(depth int) {
	fmt.Fprint(p.writer, strings.Repeat(p.indent, depth))
}
