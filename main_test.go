package main

import (
	"testing"
	"time"
)

func TestTimeStampString(t *testing.T) {
	var testTime time.Time
	var timeStamp TimeStamp

	if testTime != timeStamp.TimeValue {
		t.Errorf("Expected zero value (%s) and got value: %s", testTime.String(), timeStamp.String())
	}

	if testTime.String() != timeStamp.String() {
		t.Errorf("Expected that testTime's string value (%s) is equal to the String value of timeStamp: %s", testTime.String(), timeStamp.String())
	}

	now := time.Now()
	testTime = now
	timeStamp.TimeValue = now

	if testTime.String() != timeStamp.String() {
		t.Errorf("Expected that testTime's string value (%s) is equal to the String value of timeStamp: %s", testTime.String(), timeStamp.String())
	}
}

func TestTimeStampUnixConversion(t *testing.T) {
	timeStamp := TimeStamp{
		TimeValue: time.Now(),
	}

	testUnix := timeStamp.TimeValue.UTC().Unix()
	resultUnix := timeStamp.ConvertToUnix()

	if resultUnix != testUnix {
		t.Errorf("Expected that testUnix (%v) is equal to result_unix: %v", testUnix, resultUnix)
	}

	timeStamp.SetFromUnix(resultUnix)
	testTime := time.Unix(testUnix, 0).UTC()

	if testTime.String() != timeStamp.String() {
		t.Errorf("Expected that testTime's string value (%s) is equal to the String value of timeStamp: %s", testTime.String(), timeStamp.String())
	}
}

func TestNewDataLayer(t *testing.T) {
	testDB := NewDataLayer()
	go testDB.StartDataLayer()
	timeStamp, err := testDB.Get()

	if err != nil {
		t.Errorf("NewDataLayer should initialize the DB so we should get a timeStamp value without errors")
	}

	testTimeStamp := TimeStamp{}

	if timeStamp != testTimeStamp {
		t.Error("Expected that testTimeStamp: ", testTimeStamp, " should be equal to timeStamp: ", timeStamp)
	}
}

func TestLocalDB(t *testing.T) {
	var testDB DataLayer = &LocalDB{}

	dbTimeStamp, err := testDB.Get()
	if err == nil {
		t.Errorf("Test DB should throw an error cause it is not initialized. Error: %s", err)
	}

	testTimeStamp := TimeStamp{
		TimeValue: time.Now(),
	}

	err = testDB.Set(&testTimeStamp)
	if err == nil {
		t.Errorf("Test DB should throw an error cause it is not initialized. Error: %s", err)
	}

	err = testDB.Set(nil)
	if err == nil {
		t.Errorf("Test DB should throw an error cause the given time stamp is nil pointer. Error: %s", err)
	}

	testDB = NewDataLayer()
	go testDB.StartDataLayer()

	err = testDB.Set(&testTimeStamp)

	if err != nil {
		t.Errorf("Test DB should be initialized correctly. Error: %s", err)
	}

	dbTimeStamp, err = testDB.Get()
	if err != nil {
		t.Errorf("Test DB should be initialized correctly. Error: %s", err)
	}

	if testTimeStamp != dbTimeStamp {
		t.Error("Expected that testTimeStamp: ", testTimeStamp, " should be equal to fbYimeStamp: ", dbTimeStamp)
	}

}

func TestService(t *testing.T) {
	StartWebService()

	err := RunClient()
	if err != nil {
		t.Errorf("This run should work without errors. Error: %s", err)
	}
}
