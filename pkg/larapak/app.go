package larapak

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// App is the primary engine instance for a LarapakGo web application.
type App struct {
	*Router
	Name    string
	Version string
	Port    string
}

// New creates a new LarapakGo application with standard middlewares.
func New() *App {
	app := &App{
		Router:  NewRouter(),
		Name:    "LarapakGo Application",
		Version: "1.2.2",
		Port:    "8080",
	}

	// Default global middlewares
	app.Use(Recover())
	app.Use(Logger())

	return app
}

// ServeHTTP implements http.Handler interface for native Go compatibility.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(w, r)

	route, params := a.match(r.Method, r.URL.Path)
	if route == nil {
		if err := a.notFound(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	ctx.Params = params

	// Build handler chain with route middlewares
	handler := route.handler
	for i := len(route.handlers) - 1; i >= 0; i-- {
		handler = route.handlers[i](handler)
	}

	if err := handler(ctx); err != nil {
		log.Printf("[LarapakGo Error] %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Run starts the HTTP web server on the given address.
func (a *App) Run(addr ...string) error {
	address := ":8080"
	if len(addr) > 0 && addr[0] != "" {
		address = addr[0]
	} else if envPort := os.Getenv("PORT"); envPort != "" {
		address = ":" + envPort
	}

	fmt.Printf(`
   __                                __   _____ 
  / /   ____ __________ _____  ____ _/ /__/ ___/ 
 / /   / __ '/ ___/ __ '/ __ \/ __ '/ //_/ / __  
/ /___/ /_/ / /  / /_/ / /_/ / /_/ / ,< / /_/ /  
/_____/\__,_/_/   \__,_/ .___/\__,_/_/|_|\____/   
                      /_/                         
  LarapakGo Web Framework - Versi %s
  Mlayu cepet ing http://localhost%s
  [Basa: Go (Golang) • 100%% Standard Library]
`, a.Version, address)

	return http.ListenAndServe(address, a)
}

