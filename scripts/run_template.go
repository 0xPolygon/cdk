package main

import (
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

func main() {
	tmpl := template.New("t1")
	content, err := readFile(os.Args[1])
	if err != nil {
		log.Fatalf("Error loading template: %v", err)
	}
	content = replaceDotsInTemplateVariables(content)
	tmpl = template.Must(tmpl.Parse(content))

	if err := tmpl.Execute(os.Stdout, environmentToMap()); err != nil {
		log.Fatalf("Error executing template: %v", err)
	}
}
func replaceDotsInTemplateVariables(template string) string {
	re := regexp.MustCompile(`{{\s*\.([^{}]*)\s*}}`)
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		match = strings.ReplaceAll(match[3:], ".", "_")
		return "{{." + match
	})
	return result
}

func readFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func environmentToMap() map[string]any {
	envVars := make(map[string]any)
	for _, e := range os.Environ() {
		pair := splitAtFirst(e, '=')
		envVars[pair[0]] = pair[1]
	}
	return envVars
}

func splitAtFirst(s string, sep rune) [2]string {
	for i, c := range s {
		if c == sep {
			return [2]string{s[:i], s[i+1:]}
		}
	}
	return [2]string{s, ""}
}
