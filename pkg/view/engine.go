package view

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/IlhamHattaManggala/larapak-go/pkg/rikatlive"
)

// Engine parses and renders Rikat .nview and .html templates.
type Engine struct {
	viewsDir     string
	cacheEnabled bool
	cacheMu      sync.RWMutex
	cache        map[string]string
}

// New creates a new Rikat Template Engine pointing to the views directory.
func New(viewsDir string) *Engine {
	if viewsDir == "" {
		viewsDir = "resources/views"
	}
	cacheEnabled := strings.ToLower(os.Getenv("APP_ENV")) == "production"
	return &Engine{
		viewsDir:     viewsDir,
		cacheEnabled: cacheEnabled,
		cache:        make(map[string]string),
	}
}

// EnableCache explicitly enables or disables in-memory template caching.
func (e *Engine) EnableCache(enable bool) {
	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()
	e.cacheEnabled = enable
	if !enable {
		e.cache = make(map[string]string)
	}
}

// ClearCache flushes the in-memory template cache.
func (e *Engine) ClearCache() {
	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()
	e.cache = make(map[string]string)
}

// RenderTemplate parses a template file with given layout inheritance and data.
func (e *Engine) RenderTemplate(viewName string, data map[string]any) (string, error) {
	var content string

	e.cacheMu.RLock()
	cached, isCached := e.cache[viewName]
	e.cacheMu.RUnlock()

	if isCached && e.cacheEnabled {
		content = cached
	} else {
		filePath := e.resolveViewPath(viewName)
		rawBytes, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("cithakan ora ditemokna: %s (%w)", filePath, err)
		}
		content = string(rawBytes)
		if e.cacheEnabled {
			e.cacheMu.Lock()
			e.cache[viewName] = content
			e.cacheMu.Unlock()
		}
	}

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

	// Extract sections: @section "name" ... @endsection or @rampung
	sectionPattern := regexp.MustCompile(`(?s)@section\s+["']([^"']+)["'](.*?)(?:@endsection|@rampung)`)
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

	// Handle @flash("key") ... @endflash
	flashPattern := regexp.MustCompile(`(?s)@flash\s*\(\s*["']([^"']+)["']\s*\)(.*?)(?:@endflash|@rampung)`)
	output = flashPattern.ReplaceAllStringFunc(output, func(match string) string {
		sub := flashPattern.FindStringSubmatch(match)
		if len(sub) > 2 {
			key := sub[1]
			body := sub[2]
			var msg string
			if flashMap, ok := data["flash"].(map[string]string); ok {
				msg = flashMap[key]
			} else if flashMap, ok := data["flash"].(map[string]any); ok {
				if v, ok := flashMap[key]; ok {
					msg = fmt.Sprintf("%v", v)
				}
			} else if val, ok := data[key]; ok {
				msg = fmt.Sprintf("%v", val)
			}
			if msg != "" {
				res := strings.ReplaceAll(body, "{{ message }}", html.EscapeString(msg))
				res = strings.ReplaceAll(res, "{{ pesan }}", html.EscapeString(msg))
				return res
			}
		}
		return ""
	})

	// Handle @session("key") ... @endsession
	sessionPattern := regexp.MustCompile(`(?s)@session\s*\(\s*["']([^"']+)["']\s*\)(.*?)(?:@endsession|@rampung)`)
	output = sessionPattern.ReplaceAllStringFunc(output, func(match string) string {
		sub := sessionPattern.FindStringSubmatch(match)
		if len(sub) > 2 {
			key := sub[1]
			body := sub[2]
			var valStr string
			if sessMap, ok := data["session"].(map[string]any); ok {
				if v, ok := sessMap[key]; ok {
					valStr = fmt.Sprintf("%v", v)
				}
			} else if val, ok := data[key]; ok {
				valStr = fmt.Sprintf("%v", val)
			}
			if valStr != "" {
				res := strings.ReplaceAll(body, "{{ value }}", html.EscapeString(valStr))
				res = strings.ReplaceAll(res, "{{ nilai }}", html.EscapeString(valStr))
				return res
			}
		}
		return ""
	})

	// Handle RikatLive directive: @rikatlive("ComponentName") or @live("ComponentName")
	livePattern := regexp.MustCompile(`@(?:rikatlive|live)\s*\(\s*["']([^"']+)["']\s*\)`)
	output = livePattern.ReplaceAllStringFunc(output, func(match string) string {
		sub := livePattern.FindStringSubmatch(match)
		if len(sub) > 1 {
			rendered, err := rikatlive.RenderInitial(sub[1])
			if err == nil {
				return rendered
			}
			return fmt.Sprintf("<!-- RikatLive Error: %v -->", err)
		}
		return match
	})

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

// HasTemplate checks whether a template file exists in the views directory.
func (e *Engine) HasTemplate(name string) bool {
	filePath := e.resolveViewPath(name)
	_, err := os.Stat(filePath)
	return err == nil
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
