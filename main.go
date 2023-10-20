package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Spec struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Inputs      Inputs `yaml:"inputs"`
	Outputs     Inputs `yaml:"outputs"`
}

type Inputs map[string]InputSpec

type InputSpec struct {
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     string `yaml:"default"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <action>@<version> <destination>\n", os.Args[0])
		os.Exit(1)
	}

	target := strings.SplitN(os.Args[1], "@", 2)
	if len(target) != 2 {
		fmt.Fprintf(os.Stderr, "malformed action name %q\n", target)
		os.Exit(1)
	}

	actionName, actionVersion := target[0], target[1]

	destination := os.Args[2]

	resp, err := http.Get(
		fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/action.yml", actionName, actionVersion),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error downloading action.yml: %v\n", err)
		os.Exit(1)
	}
	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	spec := &Spec{}
	if err := yaml.Unmarshal(contents, spec); err != nil {
		panic(err)
	}

	if err := generate(os.Args[1], spec, destination); err != nil {
		panic(err)
	}
}
