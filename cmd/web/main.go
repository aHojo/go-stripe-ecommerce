package main

import (
	"flag"
	"fmt"
	"github.com/ahojo/go-stripe-ecommerce/internal/driver"
	"github.com/ahojo/go-stripe-ecommerce/internal/models"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"
const cssVersion = "1"

type config struct {
	port int
	env  string
	api  string
	db   struct {
		dsn string
	}
	stripe struct {
		secret string
		key    string
	}
}

type application struct {
	config        config
	infoLog       *log.Logger
	errorLog      *log.Logger
	templateCache map[string]*template.Template
	version       string
	DB models.DBModel
}

func (app *application) serve() error {
	serve := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	app.infoLog.Printf("Starting server in %s mode port %d", app.config.env, app.config.port)
	return serve.ListenAndServe()
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8080, "listen port")
	flag.StringVar(&cfg.env, "env", "development", "application environment")
	flag.StringVar(&cfg.api, "api", "http://localhost:4001", "api url")
	flag.StringVar(&cfg.db.dsn, "dsn", "root:password@tcp(localhost:3306)/widgets?parseTime=true&tls=false", "DSN")

	flag.Parse()

	cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.secret = os.Getenv("STRIPE_SECRET")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	conn, err := driver.OpenDB(cfg.db.dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer conn.Close()
	tc := make(map[string]*template.Template)

	app := &application{
		config:        cfg,
		infoLog:       infoLog,
		errorLog:      errorLog,
		templateCache: tc,
		version:       version,
		DB: models.DBModel{DB: conn},
	}

	err = app.serve()
	if err != nil {
		errorLog.Fatal(err)
	}

}
