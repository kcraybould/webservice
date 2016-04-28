package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	//these handle the routes, likely need to be in a seperate function when
	// this is all said and done
	router := mux.NewRouter()

	//Lets route to the right methods ...
	router.HandleFunc("/subscriptions", subhandler.SubPostHandler).Methods("POST")
	router.HandleFunc("/subscriptions", subhandler.SubPutHandler).Methods("PUT")
	router.HandleFunc("/subscriptions/{emailAddress}", subhandler.SubGetByIdHandler).Methods("GET")
	router.HandleFunc("/subscriptions/unsubscribe", subhandler.SubUnsubscribeHandler).Methods("POST")

	//hey, look, a web server
	log.Fatal(http.ListenAndServe(":8181", router))
}
