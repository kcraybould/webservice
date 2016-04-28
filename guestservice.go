package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kcraybould/guesthandler"
)

func main() {
	//these handle the routes, likely need to be in a seperate function when
	// this is all said and done
	router := mux.NewRouter()

	//Lets route to the right methods ...
	router.Path("/guests").Queries("firstName", "", "lastName", "").Methods("GET").HandlerFunc(guesthandler.GuestGetSearchNameHandler)
	router.Path("/guests").Queries("lastName", "").Methods("GET").HandlerFunc(guesthandler.GuestGetSearchLnameHandler)
	router.Path("/guests").Queries("emailAddress", "").Methods("GET").HandlerFunc(guesthandler.GuestGetSearchEmailHandler)
	router.HandleFunc("/guests", guesthandler.GuestGetListHandler).Methods("GET")
	router.HandleFunc("/guests/{guestId}", guesthandler.GuestGetByIdHandler).Methods("GET")

	//hey, look, a web server
	log.Fatal(http.ListenAndServe(":8181", router))
}
