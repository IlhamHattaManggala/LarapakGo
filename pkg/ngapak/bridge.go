package ngapak

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/IlhamHattaManggala/larapak-go/pkg/larapak"
	"github.com/IlhamHattaManggala/larapak-go/pkg/paths"
	"github.com/IlhamHattaManggala/larapak-go/pkg/studio"
	"github.com/IlhamHattaManggala/larapak-go/pkg/view"
)

// Engine manages execution of .ngpk scripts and web route integration.
type Engine struct {
	Evaluator  *Evaluator
	ViewEngine *view.Engine
	baseDir    string
}

// NewEngine creates a new Ngapak Engine.
func NewEngine(baseDir string, vEngine *view.Engine) *Engine {
	if baseDir == "" {
		baseDir = "."
	}
	if vEngine == nil {
		vEngine = view.New(filepath.Join(baseDir, "resources", "views"))
	}
	return &Engine{
		Evaluator:  NewEvaluator(),
		ViewEngine: vEngine,
		baseDir:    baseDir,
	}
}

// RunFile reads and executes an .ngpk script file.
func (eng *Engine) RunFile(filePath string) (any, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("gagal maca berkas .ngpk: %w", err)
	}
	return eng.RunSource(string(bytes), filePath)
}

// RunSource parses and executes Ngapak source code.
func (eng *Engine) RunSource(source, filename string) (any, error) {
	lexer := NewLexer(source, filename)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}

	parser := NewParser(tokens, filename)
	ast, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	return eng.Evaluator.ExecuteProgram(ast)
}

// BindRoutes binds the 'rute' global object and registers handlers with the Larapak App.
func (eng *Engine) BindRoutes(app *larapak.App) {
	eng.BindRoutesWithStudio(app, nil)
}

// BindRoutesWithStudio binds the 'rute' global object and registers routes with Larapak and Studio.
func (eng *Engine) BindRoutesWithStudio(app *larapak.App, st *studio.Studio) {
	ruteObj := map[string]any{
		"get": func(args []any) (any, error) {
			return eng.registerRoute(app, st, "GET", args)
		},
		"post": func(args []any) (any, error) {
			return eng.registerRoute(app, st, "POST", args)
		},
		"put": func(args []any) (any, error) {
			return eng.registerRoute(app, st, "PUT", args)
		},
		"delete": func(args []any) (any, error) {
			return eng.registerRoute(app, st, "DELETE", args)
		},
		"patch": func(args []any) (any, error) {
			return eng.registerRoute(app, st, "PATCH", args)
		},
	}

	eng.Evaluator.Globals.Set("rute", ruteObj)
}

func (eng *Engine) registerRoute(app *larapak.App, st *studio.Studio, method string, args []any) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("rute.%s mbutuhna path karo handler", strings.ToLower(method))
	}

	path := fmt.Sprintf("%v", args[0])
	action := args[1]

	// Convert route path pattern from {param} to :param if needed
	routePath := path
	if strings.Contains(routePath, "{") {
		// convert {slug} to :slug
		parts := strings.Split(routePath, "/")
		for i, p := range parts {
			if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
				parts[i] = ":" + strings.Trim(p, "{}")
			}
		}
		routePath = strings.Join(parts, "/")
	}

	targetName := "NgapakIn Closure"
	if actionStr, ok := action.(string); ok {
		targetName = actionStr
	}
	if st != nil {
		st.RegisterRoute(method, routePath, targetName)
	}

	handler := func(c *larapak.Context) error {
		reqObj, resObj := eng.createContextBridge(c)

		// 1. If action is a Function defined in Ngapak (gawe(req, res) ... rampung)
		if fn, ok := action.(*Function); ok {
			_, err := eng.Evaluator.CallCallable(fn, []any{reqObj, resObj})
			return err
		}

		// 2. If action is a string: "HomeController@index"
		if actionStr, ok := action.(string); ok {
			parts := strings.Split(actionStr, "@")
			if len(parts) == 2 {
				ctrlName := parts[0]
				methodName := parts[1]
				return eng.invokeControllerMethod(ctrlName, methodName, reqObj, resObj)
			}
		}

		return fmt.Errorf("aksi rute ora dingerteni: %v", action)
	}

	switch method {
	case "GET":
		app.GET(routePath, handler)
	case "POST":
		app.POST(routePath, handler)
	case "PUT":
		app.PUT(routePath, handler)
	case "DELETE":
		app.DELETE(routePath, handler)
	case "PATCH":
		app.PATCH(routePath, handler)
	}

	return nil, nil
}

func (eng *Engine) createContextBridge(c *larapak.Context) (map[string]any, map[string]any) {
	// req bridge
	params := make(map[string]any)
	for k, v := range c.Params {
		params[k] = v
	}

	req := map[string]any{
		"params": params,
		"method": c.Method(),
		"path":   c.Path(),
		"query": func(args []any) (any, error) {
			if len(args) == 0 {
				return "", nil
			}
			key := fmt.Sprintf("%v", args[0])
			return c.Query(key), nil
		},
	}

	// res bridge
	var res map[string]any
	res = map[string]any{
		"view": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("res.view mbutuhna jeneng view")
			}
			viewName := fmt.Sprintf("%v", args[0])
			data := make(map[string]any)
			if len(args) > 1 {
				if dMap, ok := args[1].(map[string]any); ok {
					data = dMap
				}
			}
			htmlContent, err := eng.ViewEngine.RenderTemplate(viewName, data)
			if err != nil {
				return nil, err
			}
			return nil, c.HTML(http.StatusOK, htmlContent)
		},
		"json": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, c.JSON(http.StatusOK, map[string]any{})
			}
			return nil, c.JSON(http.StatusOK, args[0])
		},
		"html": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, c.HTML(http.StatusOK, "")
			}
			return nil, c.HTML(http.StatusOK, fmt.Sprintf("%v", args[0]))
		},
		"string": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, c.String(http.StatusOK, "")
			}
			return nil, c.String(http.StatusOK, "%v", args[0])
		},
		"status": func(args []any) (any, error) {
			if len(args) > 0 {
				if code, ok := toInt(args[0]); ok {
					c.Status(code)
				}
			}
			return res, nil
		},
	}

	return req, res
}

func (eng *Engine) invokeControllerMethod(ctrlName, methodName string, req, res map[string]any) error {
	// Look for controller .ngpk file in known directory structures
	pPaths := paths.GetPaths(eng.baseDir)
	candidates := []string{
		filepath.Join(pPaths.Controllers, ctrlName+".ngpk"),
		filepath.Join(eng.baseDir, "jeroan", "Pangatur", ctrlName+".ngpk"),
		filepath.Join(eng.baseDir, "aplikasi", "Pengendali", ctrlName+".ngpk"),
		filepath.Join(eng.baseDir, "app", "Controllers", ctrlName+".ngpk"),
		filepath.Join(eng.baseDir, "app", "Http", "Controllers", ctrlName+".ngpk"),
	}

	var foundPath string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			foundPath = p
			break
		}
	}

	if foundPath == "" {
		return fmt.Errorf("berkas controller %s.ngpk ora ditemokna neng folder controllers", ctrlName)
	}

	// Execute controller file in an isolated environment that inherits Globals
	ctrlEnv := NewEnvironment(eng.Evaluator.Globals)
	bytes, err := os.ReadFile(foundPath)
	if err != nil {
		return err
	}

	lexer := NewLexer(string(bytes), foundPath)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return err
	}
	parser := NewParser(tokens, foundPath)
	ast, err := parser.Parse()
	if err != nil {
		return err
	}

	// Load controller definitions into ctrlEnv
	_, err = eng.Evaluator.executeBlock(ast.Statements, ctrlEnv)
	if err != nil {
		return err
	}

	// Look up methodName
	fnVal, ok := ctrlEnv.Get(methodName)
	if !ok {
		return fmt.Errorf("metode %q ora ditemokna neng %s.ngpk", methodName, ctrlName)
	}

	_, err = eng.Evaluator.CallCallable(fnVal, []any{req, res})
	return err
}

