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

func TestLocalDBSetAndGet(t *testing.T) {
	testTimeStamp := TimeStamp{
		TimeValue: time.Now(),
	}

	testDB := NewDataLayer()
	go testDB.StartDataLayer()

	err := testDB.Set(&testTimeStamp)

	if err != nil {
		t.Errorf("Test DB should be initialized correctly. Error: %s", err)
	}

	dbTimeStamp, err := testDB.Get()
	if err != nil {
		t.Errorf("Test DB should be initialized correctly. Error: %s", err)
	}

	if testTimeStamp != dbTimeStamp {
		t.Error("Expected that testTimeStamp: ", testTimeStamp, " should be equal to fbYimeStamp: ", dbTimeStamp)
	}

}
