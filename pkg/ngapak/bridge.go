package ngapak

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/IlhamHattaManggala/larapak-go/pkg/database"
	"github.com/IlhamHattaManggala/larapak-go/pkg/factory"
	"github.com/IlhamHattaManggala/larapak-go/pkg/larapak"
	"github.com/IlhamHattaManggala/larapak-go/pkg/mail"
	"github.com/IlhamHattaManggala/larapak-go/pkg/paths"
	"github.com/IlhamHattaManggala/larapak-go/pkg/rikatlive"
	"github.com/IlhamHattaManggala/larapak-go/pkg/security"
	"github.com/IlhamHattaManggala/larapak-go/pkg/storage"
	"github.com/IlhamHattaManggala/larapak-go/pkg/studio"
	"github.com/IlhamHattaManggala/larapak-go/pkg/validation"
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
	// Auto-mount RikatLive realtime component routes and Ngapak script loader
	rikatlive.RegisterRoutes(app)
	RegisterNgapakComponentLoader(eng.baseDir, eng, eng.ViewEngine)

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
		"wates": func(args []any) (any, error) {
			maxReq := 60
			windowSec := 60
			if len(args) > 0 {
				if m, ok := toInt(args[0]); ok {
					maxReq = m
				}
			}
			if len(args) > 1 {
				if w, ok := toInt(args[1]); ok {
					windowSec = w
				}
			}
			app.Use(larapak.Throttle(maxReq, time.Duration(windowSec)*time.Second))
			return nil, nil
		},
		"throttle": func(args []any) (any, error) {
			maxReq := 60
			windowSec := 60
			if len(args) > 0 {
				if m, ok := toInt(args[0]); ok {
					maxReq = m
				}
			}
			if len(args) > 1 {
				if w, ok := toInt(args[1]); ok {
					windowSec = w
				}
			}
			app.Use(larapak.Throttle(maxReq, time.Duration(windowSec)*time.Second))
			return nil, nil
		},
		"cache": func(args []any) (any, error) {
			sec := 60
			if len(args) > 0 {
				if s, ok := toInt(args[0]); ok {
					sec = s
				}
			}
			app.Use(larapak.CacheResponse(time.Duration(sec) * time.Second))
			return nil, nil
		},
		"simpen_cache": func(args []any) (any, error) {
			sec := 60
			if len(args) > 0 {
				if s, ok := toInt(args[0]); ok {
					sec = s
				}
			}
			app.Use(larapak.CacheResponse(time.Duration(sec) * time.Second))
			return nil, nil
		},
	}

	eng.Evaluator.Globals.Set("rute", ruteObj)

	// Bind 'db' global object for fluent database queries in .ngpk scripts
	dbObj := map[string]any{
		"tabel": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("db.tabel mbutuhna jeneng tabel")
			}
			tableName := fmt.Sprintf("%v", args[0])
			conn, err := database.GetManager().Connection()
			if err != nil {
				return nil, err
			}
			return createQueryBuilderBridge(conn.Table(tableName)), nil
		},
		"table": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("db.table requires table name")
			}
			tableName := fmt.Sprintf("%v", args[0])
			conn, err := database.GetManager().Connection()
			if err != nil {
				return nil, err
			}
			return createQueryBuilderBridge(conn.Table(tableName)), nil
		},
		"sambungan": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("db.sambungan mbutuhna jeneng koneksi")
			}
			connName := fmt.Sprintf("%v", args[0])
			conn, err := database.GetManager().Connection(connName)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"tabel": func(tArgs []any) (any, error) {
					if len(tArgs) == 0 {
						return nil, fmt.Errorf("mbutuhna jeneng tabel")
					}
					return createQueryBuilderBridge(conn.Table(fmt.Sprintf("%v", tArgs[0]))), nil
				},
				"table": func(tArgs []any) (any, error) {
					if len(tArgs) == 0 {
						return nil, fmt.Errorf("requires table name")
					}
					return createQueryBuilderBridge(conn.Table(fmt.Sprintf("%v", tArgs[0]))), nil
				},
			}, nil
		},
		"connection": func(args []any) (any, error) {
			connName := "default"
			if len(args) > 0 {
				connName = fmt.Sprintf("%v", args[0])
			}
			conn, err := database.GetManager().Connection(connName)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"table": func(tArgs []any) (any, error) {
					if len(tArgs) == 0 {
						return nil, fmt.Errorf("requires table name")
					}
					return createQueryBuilderBridge(conn.Table(fmt.Sprintf("%v", tArgs[0]))), nil
				},
				"tabel": func(tArgs []any) (any, error) {
					if len(tArgs) == 0 {
						return nil, fmt.Errorf("mbutuhna jeneng tabel")
					}
					return createQueryBuilderBridge(conn.Table(fmt.Sprintf("%v", tArgs[0]))), nil
				},
			}, nil
		},
		"duwe_akeh": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("db.duwe_akeh mbutuhna record induk lan jeneng tabel relasi")
			}
			rec, ok := args[0].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("record kudu awujud map")
			}
			relTable := fmt.Sprintf("%v", args[1])
			fk := ""
			pk := "id"
			if len(args) > 2 && args[2] != nil {
				fk = fmt.Sprintf("%v", args[2])
			}
			if len(args) > 3 && args[3] != nil {
				pk = fmt.Sprintf("%v", args[3])
			}
			conn, err := database.GetManager().Connection()
			if err != nil {
				return nil, err
			}
			qb, err := database.HasMany(conn, "", rec, relTable, fk, pk)
			if err != nil {
				return nil, err
			}
			return createQueryBuilderBridge(qb), nil
		},
		"has_many": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("db.has_many requires parent record and relation table name")
			}
			rec, ok := args[0].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("record must be a map")
			}
			relTable := fmt.Sprintf("%v", args[1])
			fk := ""
			pk := "id"
			if len(args) > 2 && args[2] != nil {
				fk = fmt.Sprintf("%v", args[2])
			}
			if len(args) > 3 && args[3] != nil {
				pk = fmt.Sprintf("%v", args[3])
			}
			conn, err := database.GetManager().Connection()
			if err != nil {
				return nil, err
			}
			qb, err := database.HasMany(conn, "", rec, relTable, fk, pk)
			if err != nil {
				return nil, err
			}
			return createQueryBuilderBridge(qb), nil
		},
		"melu": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("db.melu mbutuhna record anak lan jeneng tabel induk")
			}
			rec, ok := args[0].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("record kudu awujud map")
			}
			parentTable := fmt.Sprintf("%v", args[1])
			fk := ""
			ownerKey := "id"
			if len(args) > 2 && args[2] != nil {
				fk = fmt.Sprintf("%v", args[2])
			}
			if len(args) > 3 && args[3] != nil {
				ownerKey = fmt.Sprintf("%v", args[3])
			}
			conn, err := database.GetManager().Connection()
			if err != nil {
				return nil, err
			}
			qb, err := database.BelongsTo(conn, rec, parentTable, fk, ownerKey)
			if err != nil {
				return nil, err
			}
			return createQueryBuilderBridge(qb), nil
		},
		"belongs_to": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("db.belongs_to requires child record and parent table name")
			}
			rec, ok := args[0].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("record must be a map")
			}
			parentTable := fmt.Sprintf("%v", args[1])
			fk := ""
			ownerKey := "id"
			if len(args) > 2 && args[2] != nil {
				fk = fmt.Sprintf("%v", args[2])
			}
			if len(args) > 3 && args[3] != nil {
				ownerKey = fmt.Sprintf("%v", args[3])
			}
			conn, err := database.GetManager().Connection()
			if err != nil {
				return nil, err
			}
			qb, err := database.BelongsTo(conn, rec, parentTable, fk, ownerKey)
			if err != nil {
				return nil, err
			}
			return createQueryBuilderBridge(qb), nil
		},
		"duwe_siji": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("db.duwe_siji mbutuhna record induk lan jeneng tabel relasi")
			}
			rec, ok := args[0].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("record kudu awujud map")
			}
			relTable := fmt.Sprintf("%v", args[1])
			fk := ""
			pk := "id"
			if len(args) > 2 && args[2] != nil {
				fk = fmt.Sprintf("%v", args[2])
			}
			if len(args) > 3 && args[3] != nil {
				pk = fmt.Sprintf("%v", args[3])
			}
			conn, err := database.GetManager().Connection()
			if err != nil {
				return nil, err
			}
			qb, err := database.HasOne(conn, "", rec, relTable, fk, pk)
			if err != nil {
				return nil, err
			}
			return createQueryBuilderBridge(qb), nil
		},
		"has_one": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("db.has_one requires parent record and relation table name")
			}
			rec, ok := args[0].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("record must be a map")
			}
			relTable := fmt.Sprintf("%v", args[1])
			fk := ""
			pk := "id"
			if len(args) > 2 && args[2] != nil {
				fk = fmt.Sprintf("%v", args[2])
			}
			if len(args) > 3 && args[3] != nil {
				pk = fmt.Sprintf("%v", args[3])
			}
			conn, err := database.GetManager().Connection()
			if err != nil {
				return nil, err
			}
			qb, err := database.HasOne(conn, "", rec, relTable, fk, pk)
			if err != nil {
				return nil, err
			}
			return createQueryBuilderBridge(qb), nil
		},
	}
	eng.Evaluator.Globals.Set("db", dbObj)

	// Bind 'pabrik' and 'factory' global objects for fake data generation & seeding
	faker := factory.Default()
	pabrikObj := map[string]any{
		"jeneng": func(args []any) (any, error) {
			return faker.Jeneng(), nil
		},
		"name": func(args []any) (any, error) {
			return faker.Name(), nil
		},
		"email": func(args []any) (any, error) {
			if len(args) > 0 {
				return faker.Email(fmt.Sprintf("%v", args[0])), nil
			}
			return faker.Email(), nil
		},
		"hp": func(args []any) (any, error) {
			return faker.HP(), nil
		},
		"phone": func(args []any) (any, error) {
			return faker.Phone(), nil
		},
		"alamat": func(args []any) (any, error) {
			return faker.Alamat(), nil
		},
		"address": func(args []any) (any, error) {
			return faker.Address(), nil
		},
		"kutha": func(args []any) (any, error) {
			return faker.Kutha(), nil
		},
		"city": func(args []any) (any, error) {
			return faker.City(), nil
		},
		"teks": func(args []any) (any, error) {
			words := 10
			if len(args) > 0 {
				if w, ok := toInt(args[0]); ok {
					words = w
				}
			}
			return faker.Teks(words), nil
		},
		"text": func(args []any) (any, error) {
			words := 10
			if len(args) > 0 {
				if w, ok := toInt(args[0]); ok {
					words = w
				}
			}
			return faker.Text(words), nil
		},
		"angka": func(args []any) (any, error) {
			minVal := 1
			maxVal := 100
			if len(args) > 0 {
				if m, ok := toInt(args[0]); ok {
					minVal = m
				}
			}
			if len(args) > 1 {
				if m, ok := toInt(args[1]); ok {
					maxVal = m
				}
			}
			return faker.Angka(minVal, maxVal), nil
		},
		"number": func(args []any) (any, error) {
			minVal := 1
			maxVal := 100
			if len(args) > 0 {
				if m, ok := toInt(args[0]); ok {
					minVal = m
				}
			}
			if len(args) > 1 {
				if m, ok := toInt(args[1]); ok {
					maxVal = m
				}
			}
			return faker.Number(minVal, maxVal), nil
		},
		"rega": func(args []any) (any, error) {
			minVal := 5000
			maxVal := 100000
			if len(args) > 0 {
				if m, ok := toInt(args[0]); ok {
					minVal = m
				}
			}
			if len(args) > 1 {
				if m, ok := toInt(args[1]); ok {
					maxVal = m
				}
			}
			return faker.Rega(minVal, maxVal), nil
		},
		"price": func(args []any) (any, error) {
			minVal := 5000
			maxVal := 100000
			if len(args) > 0 {
				if m, ok := toInt(args[0]); ok {
					minVal = m
				}
			}
			if len(args) > 1 {
				if m, ok := toInt(args[1]); ok {
					maxVal = m
				}
			}
			return faker.Price(minVal, maxVal), nil
		},
		"bener_salah": func(args []any) (any, error) {
			return faker.BenerSalah(), nil
		},
		"boolean": func(args []any) (any, error) {
			return faker.Boolean(), nil
		},
		"tanggal": func(args []any) (any, error) {
			return faker.Tanggal(), nil
		},
		"date": func(args []any) (any, error) {
			return faker.Date(), nil
		},
		"pilihan": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, nil
			}
			if list, ok := args[0].([]any); ok {
				return faker.Pilihan(list), nil
			}
			return faker.Pilihan(args), nil
		},
		"choice": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, nil
			}
			if list, ok := args[0].([]any); ok {
				return faker.Choice(list), nil
			}
			return faker.Choice(args), nil
		},
		"uuid": func(args []any) (any, error) {
			return faker.UUID(), nil
		},
		"id_unik": func(args []any) (any, error) {
			return faker.UUID(), nil
		},
		"gawe": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("pabrik.gawe mbutuhna jeneng tabel lan jumlah/generator")
			}
			tableName := fmt.Sprintf("%v", args[0])
			count := 1
			var genFn any
			if c, ok := toInt(args[1]); ok {
				count = c
				if len(args) > 2 {
					genFn = args[2]
				}
			} else {
				genFn = args[1]
			}
			conn, err := database.GetManager().Connection()
			if err != nil {
				return nil, err
			}
			var created []any
			for i := 0; i < count; i++ {
				var itemData map[string]any
				if fn, ok := genFn.(*Function); ok {
					res, err := eng.Evaluator.CallCallable(fn, []any{i + 1})
					if err != nil {
						return nil, err
					}
					if m, ok := res.(map[string]any); ok {
						itemData = m
					}
				} else if dMap, ok := genFn.(map[string]any); ok {
					itemData = dMap
				}
				if itemData != nil {
					_, err := conn.Table(tableName).Insert(itemData)
					if err != nil {
						return nil, err
					}
					created = append(created, itemData)
				}
			}
			return created, nil
		},
		"create": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("factory.create requires table name and count/generator")
			}
			tableName := fmt.Sprintf("%v", args[0])
			count := 1
			var genFn any
			if c, ok := toInt(args[1]); ok {
				count = c
				if len(args) > 2 {
					genFn = args[2]
				}
			} else {
				genFn = args[1]
			}
			conn, err := database.GetManager().Connection()
			if err != nil {
				return nil, err
			}
			var created []any
			for i := 0; i < count; i++ {
				var itemData map[string]any
				if fn, ok := genFn.(*Function); ok {
					res, err := eng.Evaluator.CallCallable(fn, []any{i + 1})
					if err != nil {
						return nil, err
					}
					if m, ok := res.(map[string]any); ok {
						itemData = m
					}
				} else if dMap, ok := genFn.(map[string]any); ok {
					itemData = dMap
				}
				if itemData != nil {
					_, err := conn.Table(tableName).Insert(itemData)
					if err != nil {
						return nil, err
					}
					created = append(created, itemData)
				}
			}
			return created, nil
		},
	}
	eng.Evaluator.Globals.Set("pabrik", pabrikObj)
	eng.Evaluator.Globals.Set("factory", pabrikObj)

	// Bind 'sandi' and 'hash' global objects for password hashing
	sandiObj := map[string]any{
		"gawe": func(args []any) (any, error) {
			if len(args) == 0 {
				return "", fmt.Errorf("sandi.gawe mbutuhna password")
			}
			return security.HashPassword(fmt.Sprintf("%v", args[0]))
		},
		"make": func(args []any) (any, error) {
			if len(args) == 0 {
				return "", fmt.Errorf("hash.make requires password")
			}
			return security.HashPassword(fmt.Sprintf("%v", args[0]))
		},
		"cocok": func(args []any) (any, error) {
			if len(args) < 2 {
				return false, fmt.Errorf("sandi.cocok mbutuhna password lan hash")
			}
			return security.VerifyPassword(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1])), nil
		},
		"check": func(args []any) (any, error) {
			if len(args) < 2 {
				return false, fmt.Errorf("hash.check requires password and hash")
			}
			return security.VerifyPassword(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1])), nil
		},
	}
	eng.Evaluator.Globals.Set("sandi", sandiObj)
	eng.Evaluator.Globals.Set("hash", sandiObj)

	// Bind 'env' and 'lingkungan' helper
	envFn := func(args []any) (any, error) {
		if len(args) == 0 {
			return "", nil
		}
		key := fmt.Sprintf("%v", args[0])
		val := os.Getenv(key)
		if val == "" && len(args) > 1 {
			return args[1], nil
		}
		return val, nil
	}
	eng.Evaluator.Globals.Set("env", envFn)
	eng.Evaluator.Globals.Set("lingkungan", envFn)

	// Bind 'http' native standard library client
	httpClient := &http.Client{Timeout: 30 * time.Second}
	httpObj := map[string]any{
		"get": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("http.get mbutuhna URL")
			}
			urlStr := fmt.Sprintf("%v", args[0])
			req, err := http.NewRequest("GET", urlStr, nil)
			if err != nil {
				return nil, err
			}
			if len(args) > 1 {
				if hMap, ok := args[1].(map[string]any); ok {
					for k, v := range hMap {
						req.Header.Set(k, fmt.Sprintf("%v", v))
					}
				}
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			bodyStr := string(b)
			var jsonData any
			_ = json.Unmarshal(b, &jsonData)
			return map[string]any{
				"status": resp.StatusCode,
				"ok":     resp.StatusCode >= 200 && resp.StatusCode < 300,
				"body":   bodyStr,
				"json":   jsonData,
			}, nil
		},
		"post": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("http.post mbutuhna URL")
			}
			urlStr := fmt.Sprintf("%v", args[0])
			var bodyReader io.Reader
			contentType := "application/json"
			if len(args) > 1 && args[1] != nil {
				switch data := args[1].(type) {
				case string:
					bodyReader = strings.NewReader(data)
					contentType = "application/x-www-form-urlencoded"
				default:
					jsonBytes, err := json.Marshal(data)
					if err != nil {
						return nil, err
					}
					bodyReader = bytes.NewBuffer(jsonBytes)
				}
			}
			req, err := http.NewRequest("POST", urlStr, bodyReader)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", contentType)
			if len(args) > 2 {
				if hMap, ok := args[2].(map[string]any); ok {
					for k, v := range hMap {
						req.Header.Set(k, fmt.Sprintf("%v", v))
					}
				}
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			bodyStr := string(b)
			var jsonData any
			_ = json.Unmarshal(b, &jsonData)
			return map[string]any{
				"status": resp.StatusCode,
				"ok":     resp.StatusCode >= 200 && resp.StatusCode < 300,
				"body":   bodyStr,
				"json":   jsonData,
			}, nil
		},
	}
	eng.Evaluator.Globals.Set("http", httpObj)

	// Bind 'validasi' and 'validate'
	validateObj := map[string]any{
		"make": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("validate.make requires data and rules")
			}
			dMap, _ := args[0].(map[string]any)
			rMap := make(map[string]string)
			if rawRMap, ok := args[1].(map[string]any); ok {
				for k, val := range rawRMap {
					rMap[k] = fmt.Sprintf("%v", val)
				}
			}
			return createValidatorBridge(dMap, rMap), nil
		},
		"gawe": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("validasi.gawe mbutuhna data lan aturan")
			}
			dMap, _ := args[0].(map[string]any)
			rMap := make(map[string]string)
			if rawRMap, ok := args[1].(map[string]any); ok {
				for k, val := range rawRMap {
					rMap[k] = fmt.Sprintf("%v", val)
				}
			}
			return createValidatorBridge(dMap, rMap), nil
		},
	}
	eng.Evaluator.Globals.Set("validasi", validateObj)
	eng.Evaluator.Globals.Set("validate", validateObj)

	// Bind 'email' and 'mail' global objects for sending emails
	mailObj := map[string]any{
		"kirim": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("email.kirim mbutuhna data panampa")
			}
			var m mail.Message
			if dMap, ok := args[0].(map[string]any); ok {
				if toVal, exists := dMap["marang"]; exists {
					m.To = parseRecipients(toVal)
				} else if toVal, exists := dMap["to"]; exists {
					m.To = parseRecipients(toVal)
				}
				if sub, exists := dMap["subyek"]; exists {
					m.Subject = fmt.Sprintf("%v", sub)
				} else if sub, exists := dMap["subject"]; exists {
					m.Subject = fmt.Sprintf("%v", sub)
				}
				if body, exists := dMap["isi"]; exists {
					m.Body = fmt.Sprintf("%v", body)
				} else if body, exists := dMap["body"]; exists {
					m.Body = fmt.Sprintf("%v", body)
				}
				if vName, exists := dMap["sawangan"]; exists {
					var vData map[string]any
					if vd, ok := dMap["data"].(map[string]any); ok {
						vData = vd
					}
					rendered, err := eng.ViewEngine.RenderTemplate(fmt.Sprintf("%v", vName), vData)
					if err == nil {
						m.Body = rendered
						m.HTML = true
					}
				} else if vName, exists := dMap["view"]; exists {
					var vData map[string]any
					if vd, ok := dMap["data"].(map[string]any); ok {
						vData = vd
					}
					rendered, err := eng.ViewEngine.RenderTemplate(fmt.Sprintf("%v", vName), vData)
					if err == nil {
						m.Body = rendered
						m.HTML = true
					}
				}
				if from, exists := dMap["saka"]; exists {
					m.From = fmt.Sprintf("%v", from)
				} else if from, exists := dMap["from"]; exists {
					m.From = fmt.Sprintf("%v", from)
				}
			}
			err := mail.Send(m)
			if err != nil {
				return map[string]any{"ok": false, "error": err.Error()}, err
			}
			return map[string]any{"ok": true, "status": "sent"}, nil
		},
		"send": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("mail.send requires recipient data")
			}
			var m mail.Message
			if dMap, ok := args[0].(map[string]any); ok {
				if toVal, exists := dMap["to"]; exists {
					m.To = parseRecipients(toVal)
				} else if toVal, exists := dMap["marang"]; exists {
					m.To = parseRecipients(toVal)
				}
				if sub, exists := dMap["subject"]; exists {
					m.Subject = fmt.Sprintf("%v", sub)
				} else if sub, exists := dMap["subyek"]; exists {
					m.Subject = fmt.Sprintf("%v", sub)
				}
				if body, exists := dMap["body"]; exists {
					m.Body = fmt.Sprintf("%v", body)
				} else if body, exists := dMap["isi"]; exists {
					m.Body = fmt.Sprintf("%v", body)
				}
				if vName, exists := dMap["view"]; exists {
					var vData map[string]any
					if vd, ok := dMap["data"].(map[string]any); ok {
						vData = vd
					}
					rendered, err := eng.ViewEngine.RenderTemplate(fmt.Sprintf("%v", vName), vData)
					if err == nil {
						m.Body = rendered
						m.HTML = true
					}
				}
			}
			err := mail.Send(m)
			if err != nil {
				return map[string]any{"ok": false, "error": err.Error()}, err
			}
			return map[string]any{"ok": true, "status": "sent"}, nil
		},
	}
	eng.Evaluator.Globals.Set("email", mailObj)
	eng.Evaluator.Globals.Set("mail", mailObj)

	// Bind 'storage' and 'lumbung' global objects
	disk := storage.Default()
	storageObj := map[string]any{
		"simpen": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("storage.simpen mbutuhna path karo isi")
			}
			path := fmt.Sprintf("%v", args[0])
			content := fmt.Sprintf("%v", args[1])
			err := disk.Put(path, []byte(content))
			return err == nil, err
		},
		"put": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("storage.put requires path and content")
			}
			path := fmt.Sprintf("%v", args[0])
			content := fmt.Sprintf("%v", args[1])
			err := disk.Put(path, []byte(content))
			return err == nil, err
		},
		"jupuk": func(args []any) (any, error) {
			if len(args) == 0 {
				return "", fmt.Errorf("storage.jupuk mbutuhna path")
			}
			b, err := disk.Get(fmt.Sprintf("%v", args[0]))
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
		"get": func(args []any) (any, error) {
			if len(args) == 0 {
				return "", fmt.Errorf("storage.get requires path")
			}
			b, err := disk.Get(fmt.Sprintf("%v", args[0]))
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
		"ana": func(args []any) (any, error) {
			if len(args) == 0 {
				return false, nil
			}
			return disk.Exists(fmt.Sprintf("%v", args[0])), nil
		},
		"exists": func(args []any) (any, error) {
			if len(args) == 0 {
				return false, nil
			}
			return disk.Exists(fmt.Sprintf("%v", args[0])), nil
		},
		"busak": func(args []any) (any, error) {
			if len(args) == 0 {
				return false, nil
			}
			err := disk.Delete(fmt.Sprintf("%v", args[0]))
			return err == nil, err
		},
		"delete": func(args []any) (any, error) {
			if len(args) == 0 {
				return false, nil
			}
			err := disk.Delete(fmt.Sprintf("%v", args[0]))
			return err == nil, err
		},
		"dalan": func(args []any) (any, error) {
			if len(args) == 0 {
				return "", nil
			}
			return disk.URL(fmt.Sprintf("%v", args[0])), nil
		},
		"url": func(args []any) (any, error) {
			if len(args) == 0 {
				return "", nil
			}
			return disk.URL(fmt.Sprintf("%v", args[0])), nil
		},
	}
	eng.Evaluator.Globals.Set("storage", storageObj)
	eng.Evaluator.Globals.Set("lumbung", storageObj)
}

func createValidatorBridge(data map[string]any, rules map[string]string) map[string]any {
	v := validation.Make(data, rules)
	var vBridge map[string]any
	vBridge = map[string]any{
		"passes": func(args []any) (any, error) {
			return v.Passes(), nil
		},
		"bener": func(args []any) (any, error) {
			return v.Passes(), nil
		},
		"fails": func(args []any) (any, error) {
			return v.Fails(), nil
		},
		"gagal": func(args []any) (any, error) {
			return v.Fails(), nil
		},
		"errors": func(args []any) (any, error) {
			res := make(map[string]any)
			for f, errs := range v.Errors() {
				res[f] = errs
			}
			return res, nil
		},
		"kliru": func(args []any) (any, error) {
			res := make(map[string]any)
			for f, errs := range v.Errors() {
				res[f] = errs
			}
			return res, nil
		},
		"first": func(args []any) (any, error) {
			if len(args) == 0 {
				return "", nil
			}
			return v.First(fmt.Sprintf("%v", args[0])), nil
		},
		"firsts": func(args []any) (any, error) {
			res := make(map[string]any)
			for f, msg := range v.AllFirstErrors() {
				res[f] = msg
			}
			return res, nil
		},
	}
	return vBridge
}

func wrapRecord(row map[string]any, tableName string) map[string]any {
	if row == nil {
		return nil
	}
	wrapped := make(map[string]any, len(row)+9)
	for k, v := range row {
		wrapped[k] = v
	}
	wrapped["_tabel"] = tableName

	conn, _ := database.GetManager().Connection()

	// duwe_akeh / has_many
	hasManyFn := func(args []any) (any, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("duwe_akeh mbutuhna jeneng tabel relasi")
		}
		relTable := fmt.Sprintf("%v", args[0])
		fk := ""
		pk := "id"
		if len(args) > 1 && args[1] != nil {
			fk = fmt.Sprintf("%v", args[1])
		}
		if len(args) > 2 && args[2] != nil {
			pk = fmt.Sprintf("%v", args[2])
		}
		qb, err := database.HasMany(conn, tableName, wrapped, relTable, fk, pk)
		if err != nil {
			return nil, err
		}
		return createQueryBuilderBridge(qb), nil
	}
	wrapped["duwe_akeh"] = hasManyFn
	wrapped["has_many"] = hasManyFn

	// duwe_siji / has_one
	hasOneFn := func(args []any) (any, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("duwe_siji mbutuhna jeneng tabel relasi")
		}
		relTable := fmt.Sprintf("%v", args[0])
		fk := ""
		pk := "id"
		if len(args) > 1 && args[1] != nil {
			fk = fmt.Sprintf("%v", args[1])
		}
		if len(args) > 2 && args[2] != nil {
			pk = fmt.Sprintf("%v", args[2])
		}
		qb, err := database.HasOne(conn, tableName, wrapped, relTable, fk, pk)
		if err != nil {
			return nil, err
		}
		return createQueryBuilderBridge(qb), nil
	}
	wrapped["duwe_siji"] = hasOneFn
	wrapped["has_one"] = hasOneFn

	// melu / belongs_to
	belongsToFn := func(args []any) (any, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("melu mbutuhna jeneng tabel induk")
		}
		parentTable := fmt.Sprintf("%v", args[0])
		fk := ""
		ownerKey := "id"
		if len(args) > 1 && args[1] != nil {
			fk = fmt.Sprintf("%v", args[1])
		}
		if len(args) > 2 && args[2] != nil {
			ownerKey = fmt.Sprintf("%v", args[2])
		}
		qb, err := database.BelongsTo(conn, wrapped, parentTable, fk, ownerKey)
		if err != nil {
			return nil, err
		}
		return createQueryBuilderBridge(qb), nil
	}
	wrapped["melu"] = belongsToFn
	wrapped["belongs_to"] = belongsToFn

	return wrapped
}

func createQueryBuilderBridge(qb *database.QueryBuilder) map[string]any {
	var bridge map[string]any
	bridge = map[string]any{
		"pilih": func(args []any) (any, error) {
			var cols []string
			for _, a := range args {
				cols = append(cols, fmt.Sprintf("%v", a))
			}
			qb.Select(cols...)
			return bridge, nil
		},
		"select": func(args []any) (any, error) {
			var cols []string
			for _, a := range args {
				cols = append(cols, fmt.Sprintf("%v", a))
			}
			qb.Select(cols...)
			return bridge, nil
		},
		"neng": func(args []any) (any, error) {
			if len(args) >= 3 {
				qb.Where(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1]), args[2])
			} else if len(args) == 2 {
				qb.Where(fmt.Sprintf("%v", args[0]), "=", args[1])
			}
			return bridge, nil
		},
		"where": func(args []any) (any, error) {
			if len(args) >= 3 {
				qb.Where(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1]), args[2])
			} else if len(args) == 2 {
				qb.Where(fmt.Sprintf("%v", args[0]), "=", args[1])
			}
			return bridge, nil
		},
		"ngendi": func(args []any) (any, error) {
			if len(args) >= 3 {
				qb.Where(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1]), args[2])
			} else if len(args) == 2 {
				qb.Where(fmt.Sprintf("%v", args[0]), "=", args[1])
			}
			return bridge, nil
		},
		"urut": func(args []any) (any, error) {
			if len(args) > 1 {
				qb.OrderBy(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1]))
			} else if len(args) == 1 {
				qb.OrderBy(fmt.Sprintf("%v", args[0]), "ASC")
			}
			return bridge, nil
		},
		"orderBy": func(args []any) (any, error) {
			if len(args) > 1 {
				qb.OrderBy(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1]))
			} else if len(args) == 1 {
				qb.OrderBy(fmt.Sprintf("%v", args[0]), "ASC")
			}
			return bridge, nil
		},
		"wates": func(args []any) (any, error) {
			if len(args) > 0 {
				if n, ok := toInt(args[0]); ok {
					qb.Limit(n)
				}
			}
			return bridge, nil
		},
		"limit": func(args []any) (any, error) {
			if len(args) > 0 {
				if n, ok := toInt(args[0]); ok {
					qb.Limit(n)
				}
			}
			return bridge, nil
		},
		"jupuk": func(args []any) (any, error) {
			res, err := qb.Get()
			if err != nil {
				return []any{}, nil
			}
			var list []any
			for _, r := range res {
				list = append(list, wrapRecord(r, qb.TableName()))
			}
			return list, nil
		},
		"get": func(args []any) (any, error) {
			res, err := qb.Get()
			if err != nil {
				return []any{}, nil
			}
			var list []any
			for _, r := range res {
				list = append(list, wrapRecord(r, qb.TableName()))
			}
			return list, nil
		},
		"kabeh": func(args []any) (any, error) {
			res, err := qb.Get()
			if err != nil {
				return []any{}, nil
			}
			var list []any
			for _, r := range res {
				list = append(list, wrapRecord(r, qb.TableName()))
			}
			return list, nil
		},
		"all": func(args []any) (any, error) {
			res, err := qb.Get()
			if err != nil {
				return []any{}, nil
			}
			var list []any
			for _, r := range res {
				list = append(list, wrapRecord(r, qb.TableName()))
			}
			return list, nil
		},
		"siji": func(args []any) (any, error) {
			row, err := qb.First()
			if err != nil || row == nil {
				return nil, nil
			}
			return wrapRecord(row, qb.TableName()), nil
		},
		"first": func(args []any) (any, error) {
			row, err := qb.First()
			if err != nil || row == nil {
				return nil, nil
			}
			return wrapRecord(row, qb.TableName()), nil
		},
		"golek": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("golek mbutuhna id")
			}
			row, err := qb.Where("id", "=", args[0]).First()
			if err != nil || row == nil {
				return nil, nil
			}
			return wrapRecord(row, qb.TableName()), nil
		},
		"find": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("find requires id")
			}
			row, err := qb.Where("id", "=", args[0]).First()
			if err != nil || row == nil {
				return nil, nil
			}
			return wrapRecord(row, qb.TableName()), nil
		},
		"cacah": func(args []any) (any, error) {
			return qb.Count()
		},
		"count": func(args []any) (any, error) {
			return qb.Count()
		},
		"tambah": func(args []any) (any, error) {
			if len(args) == 0 {
				return int64(0), fmt.Errorf("tambah mbutuhna data dictionary")
			}
			if dMap, ok := args[0].(map[string]any); ok {
				return qb.Insert(dMap)
			}
			return int64(0), fmt.Errorf("data kudu wujud dictionary")
		},
		"insert": func(args []any) (any, error) {
			if len(args) == 0 {
				return int64(0), fmt.Errorf("insert requires data map")
			}
			if dMap, ok := args[0].(map[string]any); ok {
				return qb.Insert(dMap)
			}
			return int64(0), fmt.Errorf("data must be a map")
		},
		"owahi": func(args []any) (any, error) {
			if len(args) == 0 {
				return int64(0), fmt.Errorf("owahi mbutuhna data dictionary")
			}
			if dMap, ok := args[0].(map[string]any); ok {
				return qb.Update(dMap)
			}
			return int64(0), fmt.Errorf("data kudu wujud dictionary")
		},
		"nganyari": func(args []any) (any, error) {
			if len(args) == 0 {
				return int64(0), fmt.Errorf("nganyari mbutuhna data dictionary")
			}
			if dMap, ok := args[0].(map[string]any); ok {
				return qb.Update(dMap)
			}
			return int64(0), fmt.Errorf("data kudu wujud dictionary")
		},
		"update": func(args []any) (any, error) {
			if len(args) == 0 {
				return int64(0), fmt.Errorf("update requires data map")
			}
			if dMap, ok := args[0].(map[string]any); ok {
				return qb.Update(dMap)
			}
			return int64(0), fmt.Errorf("data must be a map")
		},
		"busak": func(args []any) (any, error) {
			return qb.Delete()
		},
		"mbusak": func(args []any) (any, error) {
			return qb.Delete()
		},
		"delete": func(args []any) (any, error) {
			return qb.Delete()
		},
		"toSql": func(args []any) (any, error) {
			sqlStr, _ := qb.ToSelectSQL()
			return sqlStr, nil
		},
		"kaca": func(args []any) (any, error) {
			page := 1
			perPage := 15
			if len(args) > 0 {
				if p, ok := toInt(args[0]); ok {
					page = p
				}
			}
			if len(args) > 1 {
				if pp, ok := toInt(args[1]); ok {
					perPage = pp
				}
			}
			pag, err := qb.Paginate(page, perPage)
			if err != nil {
				return nil, err
			}
			if rawList, ok := pag["data"].([]map[string]any); ok {
				var wrappedList []any
				for _, r := range rawList {
					wrappedList = append(wrappedList, wrapRecord(r, qb.TableName()))
				}
				pag["data"] = wrappedList
			}
			return pag, nil
		},
		"paginate": func(args []any) (any, error) {
			page := 1
			perPage := 15
			if len(args) > 0 {
				if p, ok := toInt(args[0]); ok {
					page = p
				}
			}
			if len(args) > 1 {
				if pp, ok := toInt(args[1]); ok {
					perPage = pp
				}
			}
			pag, err := qb.Paginate(page, perPage)
			if err != nil {
				return nil, err
			}
			if rawList, ok := pag["data"].([]map[string]any); ok {
				var wrappedList []any
				for _, r := range rawList {
					wrappedList = append(wrappedList, wrapRecord(r, qb.TableName()))
				}
				pag["data"] = wrappedList
			}
			return pag, nil
		},
	}
	return bridge
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

	var routeBridge map[string]any
	routeBridge = map[string]any{
		"wates": func(bArgs []any) (any, error) {
			maxReq := 60
			windowSec := 60
			if len(bArgs) > 0 {
				if m, ok := toInt(bArgs[0]); ok {
					maxReq = m
				}
			}
			if len(bArgs) > 1 {
				if w, ok := toInt(bArgs[1]); ok {
					windowSec = w
				}
			}
			app.AddMiddlewareToLastRoute(larapak.Throttle(maxReq, time.Duration(windowSec)*time.Second))
			return routeBridge, nil
		},
		"throttle": func(bArgs []any) (any, error) {
			maxReq := 60
			windowSec := 60
			if len(bArgs) > 0 {
				if m, ok := toInt(bArgs[0]); ok {
					maxReq = m
				}
			}
			if len(bArgs) > 1 {
				if w, ok := toInt(bArgs[1]); ok {
					windowSec = w
				}
			}
			app.AddMiddlewareToLastRoute(larapak.Throttle(maxReq, time.Duration(windowSec)*time.Second))
			return routeBridge, nil
		},
		"cache": func(bArgs []any) (any, error) {
			sec := 60
			if len(bArgs) > 0 {
				if s, ok := toInt(bArgs[0]); ok {
					sec = s
				}
			}
			app.AddMiddlewareToLastRoute(larapak.CacheResponse(time.Duration(sec)*time.Second))
			return routeBridge, nil
		},
		"simpen_cache": func(bArgs []any) (any, error) {
			sec := 60
			if len(bArgs) > 0 {
				if s, ok := toInt(bArgs[0]); ok {
					sec = s
				}
			}
			app.AddMiddlewareToLastRoute(larapak.CacheResponse(time.Duration(sec)*time.Second))
			return routeBridge, nil
		},
	}

	return routeBridge, nil
}

func parseRecipients(val any) []string {
	if s, ok := val.(string); ok {
		return []string{s}
	}
	if list, ok := val.([]any); ok {
		var res []string
		for _, item := range list {
			res = append(res, fmt.Sprintf("%v", item))
		}
		return res
	}
	if list, ok := val.([]string); ok {
		return list
	}
	return nil
}

func (eng *Engine) createContextBridge(c *larapak.Context) (map[string]any, map[string]any) {
	// req bridge
	params := make(map[string]any)
	for k, v := range c.Params {
		params[k] = v
	}

	sess := c.Session()
	sesiObj := map[string]any{
		"simpen": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("sesi.simpen mbutuhna key lan value")
			}
			sess.Set(fmt.Sprintf("%v", args[0]), args[1])
			return nil, nil
		},
		"set": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("session.set requires key and value")
			}
			sess.Set(fmt.Sprintf("%v", args[0]), args[1])
			return nil, nil
		},
		"jupuk": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, nil
			}
			val, ok := sess.Get(fmt.Sprintf("%v", args[0]))
			if !ok && len(args) > 1 {
				return args[1], nil
			}
			return val, nil
		},
		"get": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, nil
			}
			val, ok := sess.Get(fmt.Sprintf("%v", args[0]))
			if !ok && len(args) > 1 {
				return args[1], nil
			}
			return val, nil
		},
		"busak": func(args []any) (any, error) {
			if len(args) > 0 {
				sess.Delete(fmt.Sprintf("%v", args[0]))
			}
			return nil, nil
		},
		"delete": func(args []any) (any, error) {
			if len(args) > 0 {
				sess.Delete(fmt.Sprintf("%v", args[0]))
			}
			return nil, nil
		},
		"kilat": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("sesi.kilat mbutuhna key lan pesan")
			}
			sess.Flash(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1]))
			return nil, nil
		},
		"flash": func(args []any) (any, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("session.flash requires key and message")
			}
			sess.Flash(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1]))
			return nil, nil
		},
		"kilat_jupuk": func(args []any) (any, error) {
			if len(args) == 0 {
				return "", nil
			}
			msg := sess.GetFlash(fmt.Sprintf("%v", args[0]))
			if msg == "" && len(args) > 1 {
				return fmt.Sprintf("%v", args[1]), nil
			}
			return msg, nil
		},
		"get_flash": func(args []any) (any, error) {
			if len(args) == 0 {
				return "", nil
			}
			msg := sess.GetFlash(fmt.Sprintf("%v", args[0]))
			if msg == "" && len(args) > 1 {
				return fmt.Sprintf("%v", args[1]), nil
			}
			return msg, nil
		},
		"kabeh": func(args []any) (any, error) {
			return sess.All(), nil
		},
		"all": func(args []any) (any, error) {
			return sess.All(), nil
		},
	}
	eng.Evaluator.Globals.Set("sesi", sesiObj)
	eng.Evaluator.Globals.Set("session", sesiObj)

	req := map[string]any{
		"params":  params,
		"method":  c.Method(),
		"path":    c.Path(),
		"sesi":    sesiObj,
		"session": sesiObj,
		"query": func(args []any) (any, error) {
			if len(args) == 0 {
				return "", nil
			}
			key := fmt.Sprintf("%v", args[0])
			return c.Query(key), nil
		},
		"input": func(args []any) (any, error) {
			if len(args) == 0 {
				return "", nil
			}
			key := fmt.Sprintf("%v", args[0])
			all := c.AllInputs()
			if val, ok := all[key]; ok {
				return val, nil
			}
			if len(args) > 1 {
				return args[1], nil
			}
			return "", nil
		},
		"all": func(args []any) (any, error) {
			return c.AllInputs(), nil
		},
		"kabeh": func(args []any) (any, error) {
			return c.AllInputs(), nil
		},
		"validate": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("req.validate requires rules")
			}
			rMap := make(map[string]string)
			if rawRMap, ok := args[0].(map[string]any); ok {
				for k, val := range rawRMap {
					rMap[k] = fmt.Sprintf("%v", val)
				}
			}
			return createValidatorBridge(c.AllInputs(), rMap), nil
		},
		"validasi": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("req.validasi mbutuhna aturan validasi")
			}
			rMap := make(map[string]string)
			if rawRMap, ok := args[0].(map[string]any); ok {
				for k, val := range rawRMap {
					rMap[k] = fmt.Sprintf("%v", val)
				}
			}
			return createValidatorBridge(c.AllInputs(), rMap), nil
		},
		"old": func(args []any) (any, error) {
			if len(args) == 0 {
				return c.AllOld(), nil
			}
			key := fmt.Sprintf("%v", args[0])
			def := ""
			if len(args) > 1 {
				def = fmt.Sprintf("%v", args[1])
			}
			return c.Old(key, def), nil
		},
		"lawas": func(args []any) (any, error) {
			if len(args) == 0 {
				return c.AllOld(), nil
			}
			key := fmt.Sprintf("%v", args[0])
			def := ""
			if len(args) > 1 {
				def = fmt.Sprintf("%v", args[1])
			}
			return c.Old(key, def), nil
		},
		"file": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("req.file mbutuhna jeneng field")
			}
			key := fmt.Sprintf("%v", args[0])
			fh, err := c.FormFile(key)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"name":   fh.Filename,
				"jeneng": fh.Filename,
				"size":   fh.Size,
				"ukuran": fh.Size,
				"type":   fh.Header.Get("Content-Type"),
				"tipe":   fh.Header.Get("Content-Type"),
				"save": func(sArgs []any) (any, error) {
					if len(sArgs) == 0 {
						return false, fmt.Errorf("file.save mbutuhna path tujuan")
					}
					dst := fmt.Sprintf("%v", sArgs[0])
					err := c.SaveUploadedFile(fh, dst)
					return err == nil, err
				},
				"simpen": func(sArgs []any) (any, error) {
					if len(sArgs) == 0 {
						return false, fmt.Errorf("file.simpen mbutuhna path tujuan")
					}
					dst := fmt.Sprintf("%v", sArgs[0])
					err := c.SaveUploadedFile(fh, dst)
					return err == nil, err
				},
				"simpen_neng": func(sArgs []any) (any, error) {
					subDir := ""
					if len(sArgs) > 0 {
						subDir = fmt.Sprintf("%v", sArgs[0])
					}
					src, err := fh.Open()
					if err != nil {
						return "", err
					}
					defer src.Close()
					content, err := io.ReadAll(src)
					if err != nil {
						return "", err
					}
					return storage.Default().SaveUploaded(content, fh.Filename, subDir)
				},
				"save_to": func(sArgs []any) (any, error) {
					subDir := ""
					if len(sArgs) > 0 {
						subDir = fmt.Sprintf("%v", sArgs[0])
					}
					src, err := fh.Open()
					if err != nil {
						return "", err
					}
					defer src.Close()
					content, err := io.ReadAll(src)
					if err != nil {
						return "", err
					}
					return storage.Default().SaveUploaded(content, fh.Filename, subDir)
				},
			}, nil
		},
	}
	req["isi"] = req["input"]
	req["berkas"] = req["file"]

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
			if _, hasFlash := data["flash"]; !hasFlash {
				flashes := sess.AllFlashes()
				data["flash"] = flashes
				for k, v := range flashes {
					data["flash_"+k] = v
				}
			}
			if _, hasSess := data["session"]; !hasSess {
				data["session"] = sess.All()
			}
			if _, hasOld := data["old"]; !hasOld {
				oldData := c.AllOld()
				data["old"] = oldData
				data["lawas"] = oldData
			}
			if _, hasErrors := data["errors"]; !hasErrors {
				rawErr := sess.GetFlash("_errors")
				if rawErr != "" {
					var errMap map[string]any
					if json.Unmarshal([]byte(rawErr), &errMap) == nil {
						data["errors"] = errMap
						data["kliru"] = errMap
					}
				}
			}
			if _, hasCsrf := data["_csrf"]; !hasCsrf {
				if csrfVal, ok := c.Get("_csrf"); ok {
					data["_csrf"] = csrfVal
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
		"resource": func(args []any) (any, error) {
			return eng.handleResourceResponse(c, args)
		},
		"sumber": func(args []any) (any, error) {
			return eng.handleResourceResponse(c, args)
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
		"redirect": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("res.redirect mbutuhna target URL")
			}
			targetURL := fmt.Sprintf("%v", args[0])
			code := http.StatusFound
			if len(args) > 1 {
				if cCode, ok := toInt(args[1]); ok {
					code = cCode
				}
			}
			return nil, c.Redirect(code, targetURL)
		},
		"alurna": func(args []any) (any, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("res.alurna mbutuhna target URL")
			}
			targetURL := fmt.Sprintf("%v", args[0])
			code := http.StatusFound
			if len(args) > 1 {
				if cCode, ok := toInt(args[1]); ok {
					code = cCode
				}
			}
			return nil, c.Redirect(code, targetURL)
		},
		"with_input": func(args []any) (any, error) {
			c.FlashInputs()
			return res, nil
		},
		"karo_input": func(args []any) (any, error) {
			c.FlashInputs()
			return res, nil
		},
		"with_errors": func(args []any) (any, error) {
			if len(args) > 0 {
				if b, err := json.Marshal(args[0]); err == nil {
					sess.Flash("_errors", string(b))
				}
			}
			return res, nil
		},
		"karo_kliru": func(args []any) (any, error) {
			if len(args) > 0 {
				if b, err := json.Marshal(args[0]); err == nil {
					sess.Flash("_errors", string(b))
				}
			}
			return res, nil
		},
		"back": func(args []any) (any, error) {
			fallback := "/"
			if len(args) > 0 {
				fallback = fmt.Sprintf("%v", args[0])
			}
			return nil, c.Back(fallback)
		},
		"balek": func(args []any) (any, error) {
			fallback := "/"
			if len(args) > 0 {
				fallback = fmt.Sprintf("%v", args[0])
			}
			return nil, c.Back(fallback)
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

func (eng *Engine) handleResourceResponse(c *larapak.Context, args []any) (any, error) {
	if len(args) == 0 {
		return nil, c.JSON(http.StatusOK, map[string]any{"data": nil})
	}

	data := args[0]
	var transformer any
	if len(args) > 1 {
		transformer = args[1]
	}

	var extraMeta map[string]any
	if len(args) > 2 {
		if em, ok := args[2].(map[string]any); ok {
			extraMeta = em
		}
	}

	// Helper to transform a single item
	transformItem := func(item any) (any, error) {
		if transformer == nil {
			return item, nil
		}
		// If transformer is a closure function
		if fn, ok := transformer.(*Function); ok {
			return eng.Evaluator.CallCallable(fn, []any{item})
		}
		// If transformer is string name: "UserResource"
		if resName, ok := transformer.(string); ok {
			fn, err := eng.loadResourceTransformer(resName)
			if err != nil {
				return item, nil // fallback to untransformed item
			}
			return eng.Evaluator.CallCallable(fn, []any{item})
		}
		return item, nil
	}

	// 1. Check if data is a pagination result map (contains "data" and "total")
	if pMap, ok := data.(map[string]any); ok {
		if rawItems, hasData := pMap["data"]; hasData {
			if _, hasTotal := pMap["total"]; hasTotal {
				var transformedList []any
				if list, ok := rawItems.([]any); ok {
					for _, it := range list {
						tr, err := transformItem(it)
						if err != nil {
							return nil, err
						}
						transformedList = append(transformedList, tr)
					}
				} else if list, ok := rawItems.([]map[string]any); ok {
					for _, it := range list {
						tr, err := transformItem(it)
						if err != nil {
							return nil, err
						}
						transformedList = append(transformedList, tr)
					}
				}
				meta := map[string]any{
					"total":        pMap["total"],
					"per_page":     pMap["per_page"],
					"current_page": pMap["current_page"],
					"last_page":    pMap["last_page"],
				}
				if extraMeta != nil {
					for k, v := range extraMeta {
						meta[k] = v
					}
				}
				return nil, c.JSON(http.StatusOK, map[string]any{
					"data": transformedList,
					"meta": meta,
				})
			}
		}
	}

	// 2. Check if data is a slice/list of items
	if list, ok := data.([]any); ok {
		var transformedList []any
		for _, it := range list {
			tr, err := transformItem(it)
			if err != nil {
				return nil, err
			}
			transformedList = append(transformedList, tr)
		}
		resp := map[string]any{"data": transformedList}
		if extraMeta != nil {
			resp["meta"] = extraMeta
		}
		return nil, c.JSON(http.StatusOK, resp)
	}
	if list, ok := data.([]map[string]any); ok {
		var transformedList []any
		for _, it := range list {
			tr, err := transformItem(it)
			if err != nil {
				return nil, err
			}
			transformedList = append(transformedList, tr)
		}
		resp := map[string]any{"data": transformedList}
		if extraMeta != nil {
			resp["meta"] = extraMeta
		}
		return nil, c.JSON(http.StatusOK, resp)
	}

	// 3. Single record (map or scalar)
	single, err := transformItem(data)
	if err != nil {
		return nil, err
	}
	resp := map[string]any{"data": single}
	if extraMeta != nil {
		resp["meta"] = extraMeta
	}
	return nil, c.JSON(http.StatusOK, resp)
}

func (eng *Engine) loadResourceTransformer(resName string) (any, error) {
	if !strings.HasSuffix(resName, "Resource") && !strings.HasSuffix(resName, "resource") {
		resName += "Resource"
	}
	candidates := []string{
		filepath.Join(eng.baseDir, "app", "Resources", resName+".ngpk"),
		filepath.Join(eng.baseDir, "app", "Http", "Resources", resName+".ngpk"),
		filepath.Join(eng.baseDir, "jeroan", "Sumber", resName+".ngpk"),
		filepath.Join(eng.baseDir, "aplikasi", "Sumber", resName+".ngpk"),
		filepath.Join(eng.baseDir, "resources", resName+".ngpk"),
	}

	var foundPath string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			foundPath = p
			break
		}
	}

	if foundPath == "" {
		return nil, fmt.Errorf("berkas resource transformer %s.ngpk ora ditemokna", resName)
	}

	content, err := os.ReadFile(foundPath)
	if err != nil {
		return nil, err
	}

	resEnv := NewEnvironment(eng.Evaluator.Globals)
	lexer := NewLexer(string(content), foundPath)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}
	parser := NewParser(tokens, foundPath)
	ast, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	_, err = eng.Evaluator.executeBlock(ast.Statements, resEnv)
	if err != nil {
		return nil, err
	}

	if fnVal, ok := resEnv.Get("transform"); ok {
		return fnVal, nil
	}
	if fnVal, ok := resEnv.Get("ngowahi"); ok {
		return fnVal, nil
	}
	if fnVal, ok := resEnv.Get("toArray"); ok {
		return fnVal, nil
	}

	return nil, fmt.Errorf("fungsi 'transform' utawa 'ngowahi' ora ditemokna neng %s.ngpk", resName)
}

