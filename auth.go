package dohyo

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

type SumoLogicAuthModel struct {
	AccessID string
	AccessKey string
}

func (a *SumoLogicAuthModel) BasicAuthHeader(r *http.Request) {
	r.Header.Add("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(a.AccessID + ":" + a.AccessKey))))
}

