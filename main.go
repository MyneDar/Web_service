package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

// db is the storage solution of this web service. In our case it is an "in memory" solution
var db DataLayer

// TimeStamp contains a time.Time value called TimeValue. This struct helps to work with the json parsing in the communication
// between the clients and the service.
type TimeStamp struct {
	TimeValue time.Time `json:"time"`
}

// String gives back the underlying result of the of the TimeValue's String() function
func (t *TimeStamp) String() string {
	return t.TimeValue.String()
}

// DataLayer interface helps to implement a maintainable solution for the underlying data layer communication
type DataLayer interface {
	Set(t *TimeStamp) error
	Get() (TimeStamp, error)
	StartDataLayer()
	DBIsInitialized() bool
}

// NewDataLayer is a function that helps safely initialize our db with using a currently implemented DB solution.
func NewDataLayer() DataLayer {
	t := time.Time{}
	db := LocalDB{
		timeStamp: &t,
		reads:     make(chan readOp),
		writes:    make(chan writeOp),
	}
	return &db
}

type readOp struct {
	resp chan time.Time
}
type writeOp struct {
	val  time.Time
	resp chan error
}

// LocalDB is the in memory implementation of the DataLayer interface
type LocalDB struct {
	timeStamp *time.Time
	reads     chan readOp
	writes    chan writeOp
	//Should somehow make concurrency of this safe with the use of channels
}

// StartDataLayer is an experimental function to use stateful goroutines
func (db *LocalDB) StartDataLayer() {
	for {
		select {
		case read := <-db.reads:
			read.resp <- *db.timeStamp
		case write := <-db.writes:
			*db.timeStamp = write.val
			write.resp <- nil
		}
	}
}

func (db *LocalDB) DBIsInitialized() bool {
	if db.timeStamp == nil {
		return false
	}
	return true
}

// Set helps to set the db.timeStamp to the given time value
func (db *LocalDB) Set(t *TimeStamp) error {
	if t == nil {
		return errors.New("DB should not set uninitialized Timestamps")
	}

	if !db.DBIsInitialized() {
		return errors.New("DB is not initialized")
	}

	write := writeOp{
		val:  t.TimeValue,
		resp: make(chan error, 1),
	}

	db.writes <- write
	err := <-write.resp

	if err != nil {
		return err
	}

	return nil
}

// Get gives back the currently stored timeStamp form our db
func (db *LocalDB) Get() (TimeStamp, error) {
	var t TimeStamp
	if !db.DBIsInitialized() {
		return t, errors.New("DB is not initialized")
	}

	read := readOp{
		resp: make(chan time.Time),
	}

	db.reads <- read
	t.TimeValue = <-read.resp

	return t, nil
}

// setTimeHandler sets the user provided timestamp in the ... database/service to the user provided value that is extracted from the post request.
func handleSetTime(w http.ResponseWriter, r *http.Request) {
	var timeStamp TimeStamp
	err := json.NewDecoder(r.Body).Decode(&timeStamp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Handling setTime request: New time value: %s", timeStamp.String())
	err = db.Set(&timeStamp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

// getTimeHandler sends back the stored time in json text format
func handleGetTime(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling getTime request.")
	timeStamp, err := db.Get()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(timeStamp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// This function handles the client side communication toward our service. It sets a time in our service after that it gets the value and
// check if what is the value it get and writes it into the logs
func runClient() {

}

func main() {
	fmt.Println("Use it for testing")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	db = NewDataLayer()
	go db.StartDataLayer()

	http.HandleFunc("/getTime", handleGetTime)
	http.HandleFunc("/setTime", handleSetTime)

	log.Println("Starting HTTP server at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
