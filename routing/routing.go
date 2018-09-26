package routing

import (
	"context"
	"log"
	"maus/together-go/database"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

var emailRegExString string = "" +
	"^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+" +
	"@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?" +
	"(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"

var prefix string
var db database.Db
var emailRegEx *regexp.Regexp = regexp.MustCompile(emailRegExString)
var ctx context.Context
var cancel context.CancelFunc

func InitRouting(prefix_ string, db_ database.Db) http.Handler {
	prefix = prefix_
	db = db_

	// set router
	router := mux.NewRouter()

	// setup routes
	router.HandleFunc(prefix+"/login", handleLogin).Methods("POST")
	router.HandleFunc(prefix+"/start", secure(handleStart)).Methods("GET")
	router.HandleFunc(prefix+"/check", handleCheck).Methods("GET")

	// setup wb hub
	ctx, cancel = context.WithCancel(context.Background())

	hub := newHub()
	go hub.run()

	router.HandleFunc(prefix+"/ws", func(rw http.ResponseWriter, req *http.Request) {
		log.Println("handleWs called")
		handleWs(hub, rw, req)
	}).Methods("GET")

	return &myHandler{router}
}

func CloseRouting() {
	cancel()

	<-ctx.Done()
}

type myHandler struct {
	router *mux.Router
}

func (m *myHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
	}
	rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	rw.Header().Set(
		"Access-Control-Allow-Headers",
		"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	)
	rw.Header().Set("Access-Control-Allow-Credentials", "true")

	if req.Method == "OPTIONS" {
		return
	}

	m.router.ServeHTTP(rw, req)
}
