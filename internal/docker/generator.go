package docker

import (
	"fmt"

	"github.com/asim9115/containerix/internal/detector"
)

// GenerateNode dispatches to the correct Node/JS template.
func GenerateNode(d detector.DetectResult) (string, error) {
	return generateNode(d)
}

// GeneratePython dispatches to the correct Python template.
func GeneratePython(d detector.DetectResult) (string, error) {
	return generatePython(d)
}

// GenerateGo dispatches to the correct Go template.
func GenerateGo(d detector.DetectResult) (string, error) {
	return generateGo(d)
}

// Generate is a convenience function that dispatches on Language.
func Generate(d detector.DetectResult) (string, error) {
	switch d.Language {
	case detector.LangNode:
		return GenerateNode(d)
	case detector.LangPython:
		return GeneratePython(d)
	case detector.LangGo:
		return GenerateGo(d)
	default:
		return "", fmt.Errorf("unsupported language: %q", d.Language)
	}
}