/*
Package dohyo provides a wrapper for Sumo Logic Search Job API tasks.

Documentation for the API can be found here: https://help.sumologic.com/APIs/Search-Job-API/About-the-Search-Job-API
*/
package dohyo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	pathSeparator = "/"
)

// SumoObject contains data for a particular Sumo Logic session
type SumoObject struct {
	// Auth is a SumoLogicAuthModel, and contains the AccessID and AccessKey for
	// an authorized user.
	Auth *SumoLogicAuthModel
	// HostURL is the Sumo Logic host
	HostURL string
	// QueryURL is the Sumo Logic API endpoint
	QueryURL string
	// Headers are additional headers. These are applied to any query performed
	// using this SumoObject pointer.
	Headers          map[string]string
	SearchJobQuery   *SearchJobQuery
	SearchJobState   *SearchJobState
	SearchJobMessage *[]SearchJobMessages `json:"messages"`
}

// SearchJobData contains information about a current job. Return values are
// specified in the API documentation: https://help.sumologic.com/APIs/Search-Job-API/About-the-Search-Job-API
type SearchJobState struct {
	ID           string `json:"id"`
	State        string `json:"state"`
	MessageCount int    `json:"messageCount"`
	RecordCount  int    `json:"recordCount"`
}

// SearchJobQuery contains Sumo Logic Search Job parameters as described
// in the API documentation: https://help.sumologic.com/APIs/Search-Job-API/About-the-Search-Job-API
type SearchJobQuery struct {
	Query    string `json:"query"`
	From     string `json:"from"`
	To       string `json:"to"`
	TimeZone string `json:"timeZone"`
}

// SearchJobMessages contains messages returned from a Sumo Logic Search Job.
type SearchJobMessages struct {
	Message SearchJobMessageRaw `json:"map"`
}

type SearchJobMessageRaw struct {
	MessageTime string `json:"_messagetime"`
	Host        string `json:"_sourcehost"`
	Type        string `json:"_sourcename"`
	Log         string `json:"_raw"`
}

// GenerateAndPutAuthModel creates and populates the SumoObject.Auth value with the provides Sumo Logic AccessID & AccessKey
func (o *SumoObject) GenerateAndPutAuthModel(id, key string) {
	o.Auth = &SumoLogicAuthModel{
		AccessID:  id,
		AccessKey: key,
	}
}

// get returns the http.Response from a completed GET request. Response parsing should be performed by the caller.
func (o *SumoObject) get(url string, body io.Reader, query map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, body)
	if err != nil {
		return nil, fmt.Errorf("unable to create http GET request: %s", err)
	}

	o.encodeQuery(req, query)
	o.Auth.BasicAuthHeader(req)
	req.Header.Add("Content-Type", "application/json")

	// Note: this returns both *http.Response and error.
	return runRequest(req)
}

// post returns the http.Response from a completed POST request. Response parsing should be performed by the caller.
func (o *SumoObject) post(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("unable to create http POST request: %s", err)
	}

	o.Auth.BasicAuthHeader(req)
	req.Header.Add("Content-Type", "application/json")

	// Note: this returns both *http.Response and error.
	return runRequest(req)
}

// delete returns the http.Response from a completed DELETE request. Response parsing should be performed by the caller.
func (o *SumoObject) delete(url string, query map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create http DELETE request: %s", err)
	}

	o.Auth.BasicAuthHeader(req)
	o.encodeQuery(req, query)

	// Note: this returns both *http.Response and error.
	return runRequest(req)
}

func runRequest(r *http.Request) (*http.Response, error) {
	client := http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("unable to execute http %s request: %s", r.Method, err)
	}

	return resp, err
}

// encodeQuery populates the request query based on a string map.
func (o *SumoObject) encodeQuery(r *http.Request, query map[string]string) {
	q := r.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	r.URL.RawQuery = q.Encode()
}

func (o *SumoObject) getRequestURL() string {
	return fmt.Sprintf("%s%s", o.HostURL, o.QueryURL)
}

// SearchJob executes the SumoLogic search query. As this is executed remotely, use *SumoObject.SearchJobStatus to monitor status for updates.
func (o *SumoObject) SearchJob() error {
	requestBody, err := json.Marshal(o.SearchJobQuery)
	if err != nil {
		return fmt.Errorf("could not marshal SearchJobQuery: %s", err)
	}

	response, err := o.post(fmt.Sprintf("%s%s", o.HostURL, o.QueryURL), bytes.NewReader(requestBody))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var bdy []byte
	if response.StatusCode == 202 {
		bdy, _ = ioutil.ReadAll(response.Body)
	} else {
		return fmt.Errorf("%s", response.Status)
	}

	err = json.Unmarshal(bdy, &o.SearchJobState)
	if err != nil {
		return fmt.Errorf("could not unmarshal response body: %s", err)
	}

	return nil
}

/*
SearchJobStatus retrieves the current status of the job, and populates SumoObject.SearchJobStatus. State will be "GATHERING RESULTS" while the search is active, and "DONE GATHERING RESULTS" when the search is complete and messages can be retrieved.

An example of the type of logic to retrieve status for an executed job might be something similar to this;

	for {
		err = SumoObj.SearchJobStatus()
		if err != nil {
			// handle error
		} else {
			if SumoObj.SearchJobState.State == "GATHERING RESULTS" {
				if SumoObj.SearchJobState.MessageCount != 0 {
					// log number of messages currently found
				}
                // Pace these checks
				time.Sleep(time.Second * 3)
			} else if SumoObj.SearchJobState.State == "DONE GATHERING RESULTS" {
				if SumoObj.SearchJobState.MessageCount != 0 {
					// report total messages found
					break
				} else {
              		fmt.Println("No log entries found.")

					// Delete search job
					if err := SumoObj.DeleteSearchJob(); err != nil {
						// handle error
					}
					os.Exit(0)
				}
			}
		}
	}
*/
func (o *SumoObject) SearchJobStatus() error {
	response, err := o.get(o.getRequestURL()+pathSeparator+o.SearchJobState.ID, nil, nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(responseBody, &o.SearchJobState)
	if err != nil {
		return fmt.Errorf("could not unmarshal response body: %s", err)
	}

	return nil
}

/*
SearchJobMessages populates the SearchJobMessage struct value for a SumoObject. Paging results overwrites previously obtained messages, so existing messages should be processed/handled before retrieving additional messages with an offset. The query parameter should be a json map containing the
offset, and limit. Here is an example for how this might be used;

	var limit = "1000"
	var offset = 0
	var written = 0

	for {
		query := map[string]string{"limit": limit, "offset": fmt.Sprintf("%d", offset)}

		// This overwrites existing messages in the struct.
		err := SumoObj.SearchJobMessages(query)
		if err != nil {
			return err
		}

		err = someMessageHandler(output_file) // handle existing messages
		if err != nil {
			return err
		}
		// Report what was written.
		written += len(*SumoObj.SearchJobMessage)

		// Continue if there are more messages to receive.
		if written < SumoObj.SearchJobState.MessageCount {
			offset += 1000
		} else {
			break
		}
	}
*/
func (o *SumoObject) SearchJobMessages(query map[string]string) error {
	response, err := o.get(o.getRequestURL()+pathSeparator+o.SearchJobState.ID+pathSeparator+"messages", nil, query)
	if err != nil {
		return err
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(responseBody, &o)
	if err != nil {
		return fmt.Errorf("could not unmarshal response body: %s", err)
	}

	return nil
}

// SearchJobRecords returns a map[string]interface{} type containing the requested records. The query parameter should be a json map containing the
// offset, and limit.
func (o *SumoObject) SearchJobRecords(query map[string]string) (map[string]interface{}, error) {
	response, err := o.get(o.getRequestURL()+pathSeparator+o.SearchJobState.ID+pathSeparator+"records", nil, query)
	if err != nil {
		return nil, err
	}

	var responseBody map[string]interface{}
	json.NewDecoder(response.Body).Decode(&responseBody)

	if response.StatusCode == 400 {
		return nil, fmt.Errorf("%s", responseBody["code"])
	}

	return responseBody, nil
}

func (o *SumoObject) DeleteSearchJob() error {
	response, err := o.delete(o.getRequestURL()+pathSeparator+o.SearchJobState.ID, nil)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("could not delete search job: %s", response.Status)
	}

	return nil
}
