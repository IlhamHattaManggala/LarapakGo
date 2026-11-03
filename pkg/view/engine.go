package view

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Engine parses and renders Rikat .nview and .html templates.
type Engine struct {
	viewsDir string
}

// New creates a new Rikat Template Engine pointing to the views directory.
func New(viewsDir string) *Engine {
	if viewsDir == "" {
		viewsDir = "resources/views"
	}
	return &Engine{viewsDir: viewsDir}
}

// RenderTemplate parses a template file with given layout inheritance and data.
func (e *Engine) RenderTemplate(viewName string, data map[string]any) (string, error) {
	filePath := e.resolveViewPath(viewName)
	rawBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("cithakan ora ditemokake: %s (%w)", filePath, err)
	}

	content := string(rawBytes)
	return e.RenderString(content, data)
}

// RenderString parses template syntax from string content.
func (e *Engine) RenderString(content string, data map[string]any) (string, error) {
	layoutPattern := regexp.MustCompile(`@layout\s+["']([^"']+)["']`)
	layoutMatch := layoutPattern.FindStringSubmatch(content)

	var layoutContent string
	if len(layoutMatch) > 1 {
		layoutPath := e.resolveViewPath(layoutMatch[1])
		if bytes, err := os.ReadFile(layoutPath); err == nil {
			layoutContent = string(bytes)
		}
		// Strip @layout directive
		content = layoutPattern.ReplaceAllString(content, "")
	}

	// Extract sections: @section "name" ... @endsection
	sectionPattern := regexp.MustCompile(`(?s)@section\s+["']([^"']+)["'](.*?)@endsection`)
	sections := make(map[string]string)
	matches := sectionPattern.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		sections[m[1]] = strings.TrimSpace(m[2])
	}

	// If layout exists, replace @yield "name" with extracted section content
	output := content
	if layoutContent != "" {
		output = layoutContent
		yieldPattern := regexp.MustCompile(`@yield\s+["']([^"']+)["']`)
		output = yieldPattern.ReplaceAllStringFunc(output, func(match string) string {
			sub := yieldPattern.FindStringSubmatch(match)
			if len(sub) > 1 {
				if secVal, ok := sections[sub[1]]; ok {
					return secVal
				}
			}
			return ""
		})
	}

	// Handle @include "partial"
	includePattern := regexp.MustCompile(`@include\s+["']([^"']+)["']`)
	output = includePattern.ReplaceAllStringFunc(output, func(match string) string {
		sub := includePattern.FindStringSubmatch(match)
		if len(sub) > 1 {
			incPath := e.resolveViewPath(sub[1])
			if bytes, err := os.ReadFile(incPath); err == nil {
				return string(bytes)
			}
		}
		return ""
	})

	// Handle @csrf directive
	csrfToken := ""
	if tok, ok := data["_csrf"].(string); ok {
		csrfToken = tok
	} else if tok, ok := data["csrf_token"].(string); ok {
		csrfToken = tok
	}
	csrfTag := fmt.Sprintf(`<input type="hidden" name="_csrf" value="%s">`, html.EscapeString(csrfToken))
	output = strings.ReplaceAll(output, "@csrf", csrfTag)

	// Replace raw unescaped interpolations: {!! variable !!}
	rawPattern := regexp.MustCompile(`\{!!\s*([a-zA-Z0-9_.]+)\s*!!\}`)
	output = rawPattern.ReplaceAllStringFunc(output, func(match string) string {
		sub := rawPattern.FindStringSubmatch(match)
		if len(sub) > 1 {
			val := resolveDataValue(data, sub[1])
			return fmt.Sprintf("%v", val)
		}
		return match
	})

	// Replace safe escaped interpolations: {{ variable }}
	escapedPattern := regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.]+)\s*\}\}`)
	output = escapedPattern.ReplaceAllStringFunc(output, func(match string) string {
		sub := escapedPattern.FindStringSubmatch(match)
		if len(sub) > 1 {
			val := resolveDataValue(data, sub[1])
			return html.EscapeString(fmt.Sprintf("%v", val))
		}
		return match
	})

	return output, nil
}

func (e *Engine) resolveViewPath(name string) string {
	if !strings.HasSuffix(name, ".nview") && !strings.HasSuffix(name, ".html") {
		// check if .nview exists first, else .html
		nviewPath := filepath.Join(e.viewsDir, name+".nview")
		if _, err := os.Stat(nviewPath); err == nil {
			return nviewPath
		}
		return filepath.Join(e.viewsDir, name+".html")
	}
	return filepath.Join(e.viewsDir, name)
}

func resolveDataValue(data map[string]any, path string) any {
	parts := strings.Split(path, ".")
	var current any = data

	for _, p := range parts {
		if m, ok := current.(map[string]any); ok {
			current = m[p]
		} else {
			return ""
		}
	}
	if current == nil {
		return ""
	}
	return current
}
