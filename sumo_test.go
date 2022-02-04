package dohyo

import (
	"fmt"
	"testing"
	"time"
)

var (
	sumoObj = SumoObject{
		Auth:           auth,
		HostURL:        "https://api.us2.sumologic.com/api/v1",
		QueryURL:       "/search/jobs",
		Headers:        nil,
		SearchJobQuery: Job,
	}

	// Add credentials for testing.
	auth = &SumoLogicAuthModel{
		AccessID:  "",
		AccessKey: "",
	}

	// Edit these values as necessary
	Job = &SearchJobQuery{
		Query:    "\"Query\" | count _sourceCategory",
		From:     "2020-03-01T00:00:00",
		To:       "2020-03-29T00:00:00",
		TimeZone: "GB",
	}
)

func TestSearchJob(t *testing.T) {
	err := sumoObj.SearchJob()
	if err != nil {
		t.Error(err)
	} else {
		t.Log("search job ID returned: ", sumoObj.SearchJobState.ID)
	}
}

func TestSearchJobStatus(t *testing.T) {
	// Search Job ID will be different in the test log, don't be confused with the result from TestSearchJob
	err := sumoObj.SearchJob()
	if err != nil {
		t.Error(err)
		t.Skip("Remaining test logic skipped after failure to prevent nil pointer exception")
	}

	err = sumoObj.SearchJobStatus()
	if err != nil {
		t.Error(err)
		t.Skip("Remaining test logic skipped after failure to prevent nil pointer exception")
	} else {
		t.Log("search job status recieved: ", sumoObj.SearchJobState)
	}
}

func TestSearchJobMessages(t *testing.T) {
	// Search Job ID will be different in the test log, don't be confused with the result from TestSearchJob
	err := sumoObj.SearchJob()
	if err != nil {
		t.Error(err)
		t.Skip("Remaining test logic skipped after failure to prevent nil pointer exception")
	}

	for {
		err = sumoObj.SearchJobStatus()
		if err != nil {
			t.Error(err)
			t.Skip("Remaining test logic skipped after failure to prevent nil pointer exception")
		} else {
			if sumoObj.SearchJobState.State == "GATHERING RESULTS" {
				time.Sleep(time.Second * 5)
			} else if sumoObj.SearchJobState.State == "DONE GATHERING RESULTS" {
				t.Log(fmt.Sprintf("gathering reported complete: %d messages obtained", sumoObj.SearchJobState.MessageCount))
				break
			}
		}
	}

	var limit = "10000"
	var offset = "0"

	query := map[string]string{"limit": limit, "offset": offset}

	err = sumoObj.SearchJobMessages(query)
	if err != nil {
		t.Error(err)
		t.Skip("Remaining test logic skipped after failure to prevent nil pointer exception")
	}

	t.Log(fmt.Sprintf("Number of messages processed: %d", len(*sumoObj.SearchJobMessage)))
}

func TestSearchJobRecords(t *testing.T) {
	// Search Job ID will be different in the test log, don't be confused with the result from TestSearchJob
	err := sumoObj.SearchJob()
	if err != nil {
		t.Error(err)
		t.Skip("Remaining test logic skipped after failure to prevent nil pointer exception")
	}

	for {
		err = sumoObj.SearchJobStatus()
		if err != nil {
			t.Error(err)
			t.Skip("Remaining test logic skipped after failure to prevent nil pointer exception")
		} else {
			if sumoObj.SearchJobState.State == "GATHERING RESULTS" {
				time.Sleep(time.Second * 5)
			} else if sumoObj.SearchJobState.State == "DONE GATHERING RESULTS" {
				if sumoObj.SearchJobState.RecordCount == 0 {
					t.Skip(fmt.Sprintf("no records recieved, got %d messages", sumoObj.SearchJobState.MessageCount))
				} else {
					t.Log(fmt.Sprintf("gathering reported complete: %d records obtained", sumoObj.SearchJobState.RecordCount))
					break
				}
			}
		}
	}

	var limit = "10000"
	var offset = "0"

	query := map[string]string{"limit": limit, "offset": offset}

	records, err := sumoObj.SearchJobRecords(query)
	if err != nil {
		t.Error(err)
		t.Skip("Remaining test logic skipped after failure to prevent nil pointer exception")
	}

	t.Log(fmt.Sprintf("Number of records processed: %d", len(records)))
}

func TestDeleteJob(t *testing.T) {
	err := sumoObj.SearchJob()
	if err != nil {
		t.Error(err)
		t.Skip("Remaining test logic skipped after failure to prevent nil pointer exception")
	}

	err = sumoObj.DeleteSearchJob()
	if err != nil {
		t.Error(err)
	}
}
