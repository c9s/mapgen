package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	log "github.com/sirupsen/logrus"

	"golang.org/x/tools/go/packages"
)

var (
	typeName     = flag.String("type", "", "Type name to parse constants from; required")
	output       = flag.String("output", "", "Output file name; default: <type>map.go")
	outputStdout = flag.Bool("stdout", false, "Output generated content to stdout")
)

type ConstantInfo struct {
	Type         string
	Constants    []string
	HasIntValues bool
	Groups       map[string][]string
}

type Generator struct {
	pkg          *packages.Package
	Type         string
	constantInfo ConstantInfo
	outputBuffer bytes.Buffer
}

func main() {
	flag.Parse()

	if *typeName == "" {
		fmt.Println("Error: -type is required")
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	// Load the current package
	pkgs, err := loadPackages(args[0])
	if err != nil {
		log.Errorf("Error loading package: %v", err)
		os.Exit(1)
	}

	if len(pkgs) != 1 {
		fmt.Println("Error: Expected a single package")
		os.Exit(1)
	}

	gen := Generator{
		pkg:  pkgs[0],
		Type: *typeName,
		constantInfo: ConstantInfo{
			Type:      *typeName,
			Constants: []string{},
			Groups:    map[string][]string{},
		},
	}

	gen.parseConstants()

	var outputFile string
	if *outputStdout {
		err = gen.generateCode(os.Stdout)
	} else {
		outputFile = *output
		if outputFile == "" {
			outputFile = fmt.Sprintf("%smap.go", strings.ToLower(*typeName))
		}

		file, err := os.Create(outputFile)
		if err != nil {
			log.Errorf("Error creating output file: %v", err)
			os.Exit(1)
		}
		defer file.Close()

		err = gen.generateCode(file)
	}

	if err != nil {
		log.Errorf("Error generating code: %v", err)
		os.Exit(1)
	}

	if !*outputStdout {
		log.Infof("Generated file: %s", outputFile)
	}
}

func loadPackages(pattern string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedImports |
			packages.NeedTypesInfo,
	}

	return packages.Load(cfg, pattern)
}

func toTitleCase(input string) string {
	if len(input) == 0 {
		return input
	}
	runes := []rune(input)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func (g *Generator) parseConstants() {
	groupRegex := regexp.MustCompile(`^//\s*@group\s+(\S+)`) // Match lines like "// @group groupName"

	for _, file := range g.pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if !ok || decl.Tok != token.CONST {
				return true
			}

			log.Debugf("found const decl %+v", decl)

			var currentGroup string

			for _, spec := range decl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok || len(valueSpec.Names) == 0 {
					continue
				}

				log.Debugf("valueSpec.Names: %+v, type: %+v", valueSpec.Names, valueSpec.Type)

				// Ensure the constant is explicitly associated with the target type
				if valueSpec.Type == nil {
					continue
				}
				
				ident, ok := valueSpec.Type.(*ast.Ident)
				if !ok || ident.Name != g.constantInfo.Type {
					continue
				}

				// Check for @group annotations
				doc := decl.Doc
				if doc == nil {
					doc = valueSpec.Doc
				}
				if doc != nil {
					for _, comment := range doc.List {
						matches := groupRegex.FindStringSubmatch(comment.Text)
						if len(matches) == 2 {
							currentGroup = toTitleCase(strings.TrimSpace(matches[1]))
						}
					}
				}

				// Collect constants
				for _, name := range valueSpec.Names {
					constName := name.Name
					g.constantInfo.Constants = append(g.constantInfo.Constants, constName)
					if currentGroup != "" {
						g.constantInfo.Groups[currentGroup] = append(g.constantInfo.Groups[currentGroup], constName)
					}
				}
			}

			return true
		})
	}
}

func (g *Generator) generateCode(output *os.File) error {
	const templateText = `// Code generated by go:generate; DO NOT EDIT.
package {{.PkgName}}

var All{{.Type}}s = map[{{.Type}}]struct{}{
{{- range .Constants }}
	{{.}}: {},
{{- end }}
}

{{- range $group, $constants := .Groups }}
var All{{$group}}{{$.Type}}s = map[{{$.Type}}]struct{}{
{{- range $constants }}
	{{.}}: {},
{{- end }}
}

// All{{$group}}{{$.Type}}sKeys converts the {{$group}} group map of {{$.Type}} to a slice of {{$.Type}}
func All{{$group}}{{$.Type}}sKeys() []{{$.Type}} {
	keys := make([]{{$.Type}}, 0, len(All{{$group}}{{$.Type}}s))
	for k := range All{{$group}}{{$.Type}}s {
		keys = append(keys, k)
	}
	return keys
}

// Validate{{$group}}{{$.Type}}s validates if a value belongs to the {{$group}} group of {{$.Type}}
func Validate{{$group}}{{$.Type}}s(ch {{$.Type}}) bool {
	_, ok := All{{$group}}{{$.Type}}s[ch]
	return ok
}

// Is{{$group}}{{$.Type}} checks if the value is in the {{$group}} group of {{$.Type}}
func Is{{$group}}{{$.Type}}(ch {{$.Type}}) bool {
	_, exist := All{{$group}}{{$.Type}}s[ch]
	return exist
}
{{- end }}

var All{{.Type}}sSlice = []{{.Type}}{
{{- range .Constants }}
	{{.}},
{{- end }}
}

// {{.Type}}Strings converts a slice of {{.Type}} to a slice of {{if .HasIntValues}}int{{else}}string{{end}}
func {{.Type}}Strings(slice []{{.Type}}) (out []{{if .HasIntValues}}int{{else}}string{{end}}) {
	for _, el := range slice {
		out = append(out, {{if .HasIntValues}}int{{else}}string{{end}}(el))
	}
	return out
}

// {{.Type}}Keys converts a map of {{.Type}} to a slice of {{.Type}}
func {{.Type}}Keys(values map[{{.Type}}]struct{}) (slice []{{.Type}}) {
	for k := range values {
		slice = append(slice, k)
	}
	return slice
}

// Validate{{.Type}} validates a value of type {{.Type}}
func Validate{{.Type}}(ch {{.Type}}) bool {
	_, ok := All{{.Type}}s[ch]
	return ok
}
`

	tmpl, err := template.New("code").Parse(templateText)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	data := struct {
		PkgName      string
		Type         string
		Constants    []string
		HasIntValues bool
		Groups       map[string][]string
	}{
		PkgName:      g.pkg.Name,
		Type:         g.constantInfo.Type,
		Constants:    g.constantInfo.Constants,
		HasIntValues: g.constantInfo.HasIntValues,
		Groups:       g.constantInfo.Groups,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("error generating template: %v", err)
	}

	// Use go/format to format the generated source code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("error formatting source: %v", err)
	}

	// Write the formatted source code to the output file
	if _, err := output.Write(formatted); err != nil {
		return fmt.Errorf("error writing to output: %v", err)
	}

	return nil
}
