package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// db is the storage solution of this web service. In our case it is an "in memory" solution
var DB DataLayer

// TimeStamp contains a time.Time value called TimeValue. This struct helps to work with the json parsing in the communication
// between the clients and the service.
type TimeStamp struct {
	TimeValue time.Time
}

// String gives back the underlying result of the of the TimeValue's String() function
func (t *TimeStamp) String() string {
	return t.TimeValue.String()
}

func (t *TimeStamp) ConvertToUnix() int64 {
	return t.TimeValue.UTC().Unix()
}

func (t *TimeStamp) SetFromUnix(ux int64) {
	t.TimeValue = time.Unix(ux, 0).UTC()
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
		resp: make(chan time.Time, 1),
	}

	db.reads <- read
	t.TimeValue = <-read.resp

	return t, nil
}

// setTimeHandler sets the user provided timestamp in the DataLayer to the user provided value that is extracted from the post request.
func handleSetTime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request!", http.StatusBadRequest)
		return
	}

	var timeStamp TimeStamp
	ux, err := strconv.ParseInt(strings.TrimSpace(string(body)), 10, 64)
	if err != nil {
		http.Error(w, "Invalid unix timestamp!", http.StatusBadRequest)
		return
	}
	timeStamp.SetFromUnix(ux)

	err = DB.Set(&timeStamp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

// getTimeHandler sends back the stored time in json text format
func handleGetTime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	timeStamp, err := DB.Get()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := []byte(strconv.FormatInt(timeStamp.ConvertToUnix(), 10))

	w.Header().Set("Content-Type", "text/plain")
	w.Write(data)
}

func StartWebService() {
	DB = NewDataLayer()
	go DB.StartDataLayer()

	time.Sleep(100 * time.Millisecond)

	http.HandleFunc("/getTime", handleGetTime)
	http.HandleFunc("/setTime", handleSetTime)

	go http.ListenAndServe(":8080", nil)

	time.Sleep(200 * time.Millisecond)
}

func RunClient() error {
	client := &http.Client{Timeout: 2 * time.Second}
	timestamp := TimeStamp{TimeValue: time.Now()}

	// Store the timestamp on the web service
	data := strings.NewReader(strconv.FormatInt(timestamp.ConvertToUnix(), 10))
	resp, err := client.Post("http://127.0.0.1:8080/setTime", "text/plain", data)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("setTime failed: %s", resp.Status)
	}

	// Get the stored timestamp
	resp2, err := client.Get("http://127.0.0.1:8080/getTime")
	if err != nil {
		log.Fatalf("Request failed: %s", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		log.Fatalf("getTime failed: %s", resp2.Status)
	}

	body, err := io.ReadAll(resp2.Body)
	if err != nil {
		log.Fatal(err)
	}

	var stored TimeStamp
	received, err := strconv.ParseInt(strings.TrimSpace(string(body)), 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	stored.SetFromUnix(received)

	fmt.Println("Received value: ", stored.String())
	return nil
}

func main() {
	StartWebService()

	err := RunClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
