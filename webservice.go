package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"gopkg.in/couchbase/gocb.v1"
)

type PersonalInfo struct {
	Name      Name      `json:"name"`
	Addresses []Address `json:"addresses"`
	Phones    []Phone   `json:"phones"`
	Emails    []Email   `json:"emails"`
	Payments  []Payment `json:"payments"`
}

type Address struct {
	AddressId    int    `json:"addressId"`
	AddressType  string `json:"addressType"`
	Preferred    bool   `json:"preferred"`
	AddressLine1 string `json:"addressLine1"`
	AddressLine2 string `json:"addressLine2"`
	AddressLine3 string `json:"addressLine3"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	PostalCode   string `json:"postalCode"`
	Company      string `json:"company"`
}

type Phone struct {
	PhoneId        int    `json:"phoneId"`
	PhoneType      string `json:"phoneType"`
	Preferred      bool   `json:"preferred"`
	PhoneNumber    string `json:"phoneNumber"`
	PhoneExtension string `json:"phoneExtension"`
}

type Email struct {
	EmailId      int    `json:"emailId"`
	EmailAddress string `json:"emailaddress"`
	Preferred    bool   `json:"preferred"`
}

type Payment struct {
	PaymentId  int    `json:"paymentId"`
	CardNumber string `json:"cardNumber"`
	CardCode   string `json:"cardCode"`
	ExpireDate string `json:"expireDate"`
	Preferred  bool   `json:"preferred"`
}

type Name struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	MiddleInit string `json:"middleInit"`
	Title      string `json:"title"`
}
type Guest struct {
	GuestId      int          `json:"guestId"`
	PersonalInfo PersonalInfo `json:"personalInfo"`
}

func main() {
	//these handle the routes, likely need ot be in a speerate function when
	// this is all said and done
	router := mux.NewRouter()

	//Lets route to the right methods ...
	router.Path("/guests").Queries("firstName", "", "lastName", "").Methods("GET").HandlerFunc(guestGetSearchHandler)
	router.HandleFunc("/guests", guestGetHandler).Methods("GET")
	router.HandleFunc("/guests/{guestId}", guestGetByIdHandler).Methods("GET")

	//hey, look, a web server
	log.Fatal(http.ListenAndServe(":8181", router))
}

func guestGetSearchHandler(w http.ResponseWriter, r *http.Request) {
	lastName := r.URL.Query().Get("lastName")
	if guests, ok := returnGuestsSearch(lastName); ok {
		outJson, error := json.Marshal(guests)
		if error != nil {
			log.Println(error.Error())
			http.Error(w, error.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, string(outJson))
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Guests not found")
	}

}

func guestGetByIdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//this gets the input form the web service, spefically thre URI
	vars := mux.Vars(r)
	guestId := vars["guestId"]

	log.Println("Request for:", guestId)

	if guest, ok := connectCouch(guestId); ok {
		outJson, error := json.Marshal(guest)
		if error != nil {
			log.Println(error.Error())
			http.Error(w, error.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, string(outJson))
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Guest not found")
	}
}

func connectCouch(id string) (interface{}, bool) {
	log.Println("in the connect..")
	key := "guest::" + id

	myCluster, _ := gocb.Connect("couchbase://127.0.0.1")
	myBucket, _ := myCluster.OpenBucket("guest", "")

	var guest map[string]interface{}
	cas, _ := myBucket.Get(key, &guest)

	log.Println(cas)

	return guest, (guest != nil)
}

func guestGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if guests, ok := returnGuestsView(); ok {
		outJson, error := json.Marshal(guests)
		if error != nil {
			log.Println(error.Error())
			http.Error(w, error.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, string(outJson))
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Guests not found")
	}
}

func returnGuestsView() (interface{}, bool) {
	log.Println("in the all connect..")

	myCluster, _ := gocb.Connect("couchbase://127.0.0.1")
	myBucket, _ := myCluster.OpenBucket("guest", "")
	myQuery := gocb.NewViewQuery("list", "list")

	rows, err := myBucket.ExecuteViewQuery(myQuery)
	var guests = map[string]Name{}

	// so this is kinda messy.  Basically, we get a ViewResults type back from the gocb.v1
	// the typing gets a bit confusing, but essentially the slice is not a slice of string but a slice
	// of empty interfaces
	var row map[string]interface{}
	for rows.Next(&row) {
		slice := reflect.ValueOf(row["value"])
		var person = Name{FirstName: (slice.Index(1).Interface()).(string), LastName: (slice.Index(0).Interface()).(string), MiddleInit: (slice.Index(2).Interface()).(string)}
		index := reflect.ValueOf(row["key"])
		guests[index.String()] = person
	}

	return guests, (err == nil)
}

func returnGuestsSearch(name string) (interface{}, bool) {
	myCluster, _ := gocb.Connect("couchbase://127.0.0.1")
	myBucket, _ := myCluster.OpenBucket("guest", "")
	myQuery := gocb.NewN1qlQuery("SELECT guestId, personalInfo FROM guest WHERE personalInfo.name.lastName=$1")
	var myParams []interface{}
	myParams = append(myParams, name)

	fmt.Println("name:", name)

	rows, err := myBucket.ExecuteN1qlQuery(myQuery, myParams)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("after search query")
	//var guests []map[string]interface{}
	var guests []Guest

	//var row map[string]interface{}
	var row Guest
	for rows.Next(&row) {
		fmt.Println("row:", row)
		//try the json decode
		guests = append(guests, row)
		bytes, err := json.Marshal(row)
		fmt.Println("err: ", err)
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		fmt.Printf("bytes: %s", bytes)
	}

	_ = rows.Close()

	return guests, (err == nil)
}
