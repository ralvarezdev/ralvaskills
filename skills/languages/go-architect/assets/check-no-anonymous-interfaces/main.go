// Command check-no-anonymous-interfaces reports interface types written inline
// (x.(interface{ Policy() Policy }), a parameter, a variable) instead of
// declared with `type Name interface`.
//
// It parses the AST rather than grepping, so a multi-line declaration is not
// confused with an inline one, and it skips generated code (a "Code generated"
// marker, .pb.go files, gen/ and vendor/). Interfaces without methods
// (interface{} / any) are not reported: any is the right spelling for a truly
// unconstrained value, and what this check wants to catch is a capability that
// was never given a name.
//
// Usage:
//
//	go run main.go [-tests] [dir]
//
// It exits 0 when every interface has a name, 1 when one does not, and 2 when a
// file could not be parsed.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MaxSnippet bounds the source line the report quotes, so a one-line interface
// with many methods does not flood the terminal.
const MaxSnippet = 110

// generatedMarkerWindow is how many leading bytes of a file are searched for a
// "Code generated" marker before the parser runs.
const generatedMarkerWindow = 1024

// anonymousHint is printed (on stderr) when at least one inline interface was
// found, with the count as its only argument.
const anonymousHint = "\n%d anonymous interface(s): declare each with `type Name interface { ... }` next to its consumer.\n"

// finding is one anonymous interface in the tree.
type finding struct {
	file string
	fn   string
	text string
	line int
}

func main() {
	tests := flag.Bool("tests", false, "also scan _test.go files")
	//nolint:reassign // flag.Usage is the documented way to customize usage output
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: check-no-anonymous-interfaces [-tests] [dir]\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	findings, err := scanTree(root, *tests)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].file != findings[j].file {
			return findings[i].file < findings[j].file
		}
		return findings[i].line < findings[j].line
	})
	for _, f := range findings {
		fn := ""
		if f.fn != "" {
			fn = " in " + f.fn
		}
		fmt.Printf("%s:%d%s\n  %s\n", f.file, f.line, fn, f.text)
	}
	if len(findings) == 0 {
		fmt.Println("no anonymous interfaces")

		return
	}
	fmt.Fprintf(os.Stderr, anonymousHint, len(findings))
	os.Exit(1)
}

// scanTree walks root and collects every finding.
func scanTree(root string, tests bool) ([]finding, error) {
	var findings []finding
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "node_modules", "gen":
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || (!tests && strings.HasSuffix(path, "_test.go")) {
			return nil
		}
		found, err := scanFile(path)
		if err != nil {
			return err
		}
		findings = append(findings, found...)

		return nil
	})

	return findings, err
}

// scanFile reports the inline interfaces of one file.
func scanFile(path string) ([]finding, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if bytes.Contains(src[:min(len(src), generatedMarkerWindow)], []byte("Code generated")) {
		return nil, nil
	}
	lines := strings.Split(string(src), "\n")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	// named holds the interfaces that are type declarations; those are the
	// ones the rule wants.
	named := map[token.Pos]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		if iface, isIface := ts.Type.(*ast.InterfaceType); isIface {
			named[iface.Pos()] = true
		}

		return true
	})

	var out []finding
	ast.Inspect(file, func(n ast.Node) bool {
		it, ok := n.(*ast.InterfaceType)
		if !ok || named[it.Pos()] || len(it.Methods.List) == 0 {
			return true
		}
		out = append(out, finding{
			file: filepath.ToSlash(path),
			line: fset.Position(it.Pos()).Line,
			fn:   enclosingFunc(file, fset, it.Pos()),
			text: snippet(fset, it, lines),
		})

		return true
	})

	return out, nil
}

// enclosingFunc names the function around pos, when there is one.
func enclosingFunc(file *ast.File, fset *token.FileSet, pos token.Pos) string {
	name := ""
	ast.Inspect(file, func(n ast.Node) bool {
		fd, ok := n.(*ast.FuncDecl)
		if !ok || fd.Body == nil || fd.Name == nil {
			return true
		}
		if fd.Body.Pos() <= pos && pos <= fd.Body.End() {
			name = fd.Name.Name
		}

		return true
	})

	return name
}

// snippet is the source line where the interface starts, compacted.
func snippet(fset *token.FileSet, it *ast.InterfaceType, lines []string) string {
	line := fset.Position(it.Pos()).Line
	if line < 1 || line > len(lines) {
		return ""
	}
	text := strings.Join(strings.Fields(strings.TrimSpace(lines[line-1])), " ")
	if len(text) > MaxSnippet {
		text = text[:MaxSnippet-3] + "..."
	}

	return text
}
