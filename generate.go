package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	//go:embed codegen.go.tmpl
	tmplSource string

	//go:embed utils.go.tmpl
	utilsSource string

	funcMap = template.FuncMap{
		"FormatName": formatName,
		"Comment":    formatComment,
	}
)

// formatName formats a GraphQL name (e.g. object, field, arg) into a Go equivalent
// Example: `fooId` -> `FooID`
func formatName(s string) string {
	if len(s) == 0 {
		return s
	}

	fragments := strings.Split(s, "-")
	for i, frag := range fragments {
		if len(frag) == 0 {
			continue
		}
		fragments[i] = strings.ToUpper(string(frag[0])) + frag[1:]
	}

	// function humanize(str) {
	// 	var i, frags = str.split('_');
	// 	for (i=0; i<frags.length; i++) {
	// 	  frags[i] = frags[i].charAt(0).toUpperCase() + frags[i].slice(1);
	// 	}
	// 	return frags.join(' ');
	//   }

	// if len(s) > 0 {
	// 	s = strings.ToUpper(string(s[0])) + s[1:]
	// }
	return lintName(strings.Join(fragments, ""))
}

func formatComment(s string) string {
	if s == "" {
		return ""
	}

	lines := strings.Split(s, "\n")

	for i, l := range lines {
		lines[i] = "// " + l
	}
	return strings.Join(lines, "\n")
}

func generate(uses string, spec *Spec, destination string) error {
	tmpl, err := template.New("tmpl").Funcs(funcMap).Parse(tmplSource)
	if err != nil {
		panic(err)
	}

	var input = struct {
		Uses string
		Name string
		Spec *Spec
	}{
		Uses: uses,
		Name: filepath.Base(destination),
		Spec: spec,
	}

	var output bytes.Buffer
	if err := tmpl.Execute(&output, input); err != nil {
		return err
	}

	formatted, err := format.Source(output.Bytes())
	if err != nil {
		formatted = output.Bytes()
	}

	if err := os.MkdirAll(destination, 0755); err != nil {
		return err
	}

	if _, err := os.Stat(path.Join(destination, "dagger.json")); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		dagger, err := exec.LookPath("dagger")
		if err != nil {
			return err
		}
		cmd := exec.Cmd{
			Path: dagger,
			Args: []string{
				"dagger", "mod", "init",
				"--name", filepath.Base(destination),
				"--sdk", "go",
			},
			Env:    os.Environ(),
			Dir:    destination,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}
		fmt.Println(cmd.Args)

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if err := os.WriteFile(path.Join(destination, "main.go"), formatted, 0644); err != nil {
		return err
	}

	formattedUtils, err := format.Source([]byte(utilsSource))
	if err != nil {
		return err
	}

	if err := os.WriteFile(path.Join(destination, "utils.go"), formattedUtils, 0644); err != nil {
		return err
	}

	return nil
}
