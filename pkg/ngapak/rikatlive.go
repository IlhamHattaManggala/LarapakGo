package ngapak

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/IlhamHattaManggala/larapak-go/pkg/paths"
	"github.com/IlhamHattaManggala/larapak-go/pkg/rikatlive"
	"github.com/IlhamHattaManggala/larapak-go/pkg/view"
)

// NgapakComponent implements rikatlive.DynamicComponent for components written in NgapakIn (.ngpk).
type NgapakComponent struct {
	Name       string
	FilePath   string
	Engine     *Engine
	ViewEngine *view.Engine
	Env        *Environment
	stateKeys  map[string]bool
}

// NewNgapakComponent parses and initializes an .ngpk component file.
func NewNgapakComponent(name, filePath string, eng *Engine, vEngine *view.Engine) (*NgapakComponent, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("gagal maca berkas komponen %s: %w", filePath, err)
	}

	lexer := NewLexer(string(bytes), filePath)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("token error neng %s: %w", filePath, err)
	}

	parser := NewParser(tokens, filePath)
	ast, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("parse error neng %s: %w", filePath, err)
	}

	compEnv := NewEnvironment(eng.Evaluator.Globals)
	nc := &NgapakComponent{
		Name:       name,
		FilePath:   filePath,
		Engine:     eng,
		ViewEngine: vEngine,
		Env:        compEnv,
		stateKeys:  make(map[string]bool),
	}

	// Execute component file to initialize state and register functions
	_, err = eng.Evaluator.executeBlock(ast.Statements, compEnv)
	if err != nil {
		return nil, fmt.Errorf("eksekusi komponen %s kliru: %w", filePath, err)
	}

	// Capture all non-function variables as initial reactive state
	for k, v := range compEnv.Values() {
		if _, isFn := v.(*Function); !isFn {
			nc.stateKeys[k] = true
		}
	}

	return nc, nil
}

// GetState returns the current reactive state map of the component.
func (nc *NgapakComponent) GetState() map[string]any {
	state := make(map[string]any)
	for k := range nc.stateKeys {
		if val, ok := nc.Env.Get(k); ok {
			state[k] = val
		}
	}
	return state
}

// SetState updates the component environment with values from client snapshot/updates.
func (nc *NgapakComponent) SetState(state map[string]any) error {
	for k, v := range state {
		nc.Env.Set(k, v)
		nc.stateKeys[k] = true
	}
	return nil
}

// CallAction invokes a method defined in the .ngpk component (e.g. gawe tambah() ... rampung).
func (nc *NgapakComponent) CallAction(action string, args []any) error {
	// 1. Direct match
	val, ok := nc.Env.Get(action)
	if !ok {
		// 2. Case-insensitive match
		for k, v := range nc.Env.Values() {
			if strings.EqualFold(k, action) {
				val = v
				ok = true
				break
			}
		}
	}

	if !ok {
		return fmt.Errorf("metode/aksi '%s' ora ditemokna neng komponen '%s'", action, nc.Name)
	}

	_, err := nc.Engine.Evaluator.CallCallable(val, args)
	return err
}

// Render renders the component's HTML with its active state.
func (nc *NgapakComponent) Render() string {
	state := nc.GetState()

	// 1. If component explicitly defines a render() function
	if rVal, ok := nc.Env.Get("render"); ok {
		res, err := nc.Engine.Evaluator.CallCallable(rVal, []any{})
		if err == nil {
			if strVal, isStr := res.(string); isStr {
				// If string contains HTML tags, return directly
				if strings.Contains(strVal, "<") && strings.Contains(strVal, ">") {
					return strVal
				}
				// Otherwise, treat as view name
				if nc.ViewEngine != nil {
					if html, err := nc.ViewEngine.RenderTemplate(strVal, state); err == nil {
						return html
					}
				}
			}
		}
	}

	// 2. Look up matching template: "live/" + lowercase(name) (e.g. live/counter.nview)
	lowerName := strings.ToLower(nc.Name)
	if nc.ViewEngine != nil {
		if html, err := nc.ViewEngine.RenderTemplate("live/"+lowerName, state); err == nil {
			return html
		}
		if html, err := nc.ViewEngine.RenderTemplate(lowerName, state); err == nil {
			return html
		}
	}

	// 3. Fallback clean representation
	var stateItems []string
	for k, v := range state {
		stateItems = append(stateItems, fmt.Sprintf("<div><strong>%s:</strong> %v</div>", k, v))
	}
	return fmt.Sprintf(`<div class="rikat-component p-4 border rounded bg-white shadow-sm">
	<h4 class="font-bold text-gray-800 mb-2">%s</h4>
	%s
</div>`, nc.Name, strings.Join(stateItems, "\n"))
}

// RegisterNgapakComponentLoader connects the NgapakIn script engine with RikatLive.
func RegisterNgapakComponentLoader(baseDir string, eng *Engine, vEngine *view.Engine) {
	if baseDir == "" {
		baseDir = "."
	}

	rikatlive.RegisterLoader(func(name string) (rikatlive.Component, error) {
		p := paths.GetPaths(baseDir)
		lowerName := strings.ToLower(name)

		candidates := []string{
			filepath.Join(p.Live, name+".ngpk"),
			filepath.Join(p.Live, lowerName+".ngpk"),
			filepath.Join(baseDir, "jeroan", "Urip", name+".ngpk"),
			filepath.Join(baseDir, "jeroan", "Urip", lowerName+".ngpk"),
			filepath.Join(baseDir, "aplikasi", "Live", name+".ngpk"),
			filepath.Join(baseDir, "aplikasi", "Live", lowerName+".ngpk"),
			filepath.Join(baseDir, "app", "Live", name+".ngpk"),
			filepath.Join(baseDir, "app", "Live", lowerName+".ngpk"),
			filepath.Join(p.Views, "live", lowerName+".ngpk"),
		}

		for _, cand := range candidates {
			if _, err := os.Stat(cand); err == nil {
				return NewNgapakComponent(name, cand, eng, vEngine)
			}
		}

		return nil, fmt.Errorf("komponen .ngpk nggo '%s' ora ditemokna neng folder Live", name)
	})
}

