/*
 * Shea Odland, Von Castro, Eric Wedemire
 * CMPT315
 * Group Project: Codenames
 */

package main

import (
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
)

type tempHandler struct {
	temp *template.Template
}

// main spins up the mux router, database connection, defines the handler
// functions, logging middleware, and finally the server
func main() {
	//using gorilla mux
	router := mux.NewRouter()

	//database connection with const values
	database = dbConnect()
	//close DB on panic
	defer database.Close()

	// tmp := &tempHandler{temp: template.Must(template.ParseFiles("dist/game.tmpl"))}
	template := &tempHandler{temp: template.Must(template.ParseFiles("dist/templates/game.tmpl"))}
	//API Handlers ------------------------------------------------------------
	// Creates new game by setting up websocket
	router.HandleFunc("/api/v1/games", createGame).
		Methods(http.MethodPost)

	// router.HandleFunc("/games/{postId:[0-9]+")

	router.PathPrefix("/games/{id}").
		Methods(http.MethodGet).
		HandlerFunc(template.passTemplate)

	// Default
	router.HandleFunc("/api/v1/", defaultHandle)
	//-------------------------------------------------------------------------

	//Non-API Handlers --------------------------------------------------------
	router.HandleFunc("/games", newSocketConnection).
		Queries("id", "{gameID:[a-zA-Z0-9]+}").
		Methods(http.MethodGet)

	router.Path("/notfound").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./dist/notfound.html")
	})

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./dist/")))
	//-------------------------------------------------------------------------

	//start server and setting logging flags
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	log.Fatal(http.ListenAndServe(FULLHOST, loggingMiddleware(router)))
}

//defaultHandle will handle any API request not made to a valid endpoint by
// throwing a HTTP 400 code
func defaultHandle(writer http.ResponseWriter, request *http.Request) {
	log.Println("Not valid route")
	http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}
