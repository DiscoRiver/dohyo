GODOC

```
package dohyo // import "."

Package dohyo provides a wrapper for Sumo Logic Search Job API tasks.

Documentation for the API can be found here:
https://help.sumologic.com/APIs/Search-Job-API/About-the-Search-Job-API

CONSTANTS

const (
        PATH_SEPARATOR = "/"
)

TYPES

type SearchJobMessageRaw struct {
        MessageTime string `json:"_messagetime"`
        Host        string `json:"_sourcehost"`
        Type        string `json:"_sourcename"`
        Log         string `json:"_raw"`
}

type SearchJobMessages struct {
        Message SearchJobMessageRaw `json:"map"`
}
    SearchJobMessages contains messages returned from a Sumo Logic Search Job.

type SearchJobQuery struct {
        Query    string `json:"query"`
        From     string `json:"from"`
        To       string `json:"to"`
        TimeZone string `json:"timeZone"`
}
    SearchJobQuery contains Sumo Logic Search Job parameters as described in the
    API documentation:
    https://help.sumologic.com/APIs/Search-Job-API/About-the-Search-Job-API

type SearchJobState struct {
        ID           string `json:"id"`
        State        string `json:"state"`
        MessageCount int    `json:"messageCount"`
        RecordCount  int    `json:"recordCount"`
}
    SearchJobData contains information about a current job. Return values are
    specified in the API documentation:
    https://help.sumologic.com/APIs/Search-Job-API/About-the-Search-Job-API

type SumoLogicAuthModel struct {
        AccessID  string
        AccessKey string
}

func (a *SumoLogicAuthModel) BasicAuthHeader(r *http.Request)

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
    SumoObject contains data for a particular Sumo Logic session

func (o *SumoObject) DeleteSearchJob() error

func (o *SumoObject) GenerateAndPutAuthModel(id, key string)
    GenerateAndPutAuthModel creates and populates the SumoObject.Auth value with
    the provides Sumo Logic AccessID & AccessKey

func (o *SumoObject) SearchJob() error

func (o *SumoObject) SearchJobMessages(query map[string]string) error
    SearchJobMessages populates the SearchJobMessage struct value for a
    SumoObject. The query parameter should be a json map containing the offset,
    and limit. TODO: Test how paging behaves with struct values. Does it
    overwrite, or append?

func (o *SumoObject) SearchJobRecords(query map[string]string) (map[string]interface{}, error)
    SearchJobRecords returns a map[string]interface{} type containing the
    requested records. The query parameter should be a json map containing the
    offset, and limit. TODO: Implement struct to handle records, instead of
    returning a map. Reason is to use a pointer we can unmarshal JSON to, like
    how we handle messages above.

func (o *SumoObject) SearchJobStatus() error

```