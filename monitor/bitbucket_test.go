package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

const bitbucketResponse = `{
	"values": [
		{
			"display_name": "Negative Test",
			"nickname": "example",
			"has_2fa_enabled": null
		},
		{
			"display_name": "Positive Test",
			"nickname": "example",
			"has_2fa_enabled": true
		}
	],
	"size": 2
}`

func TestGetNon2svWorkspaceMembers(t *testing.T) {
	assert := require.New(t)
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	handler := func(w http.ResponseWriter, req *http.Request) {
		jsonBytes, err := json.Marshal(bitbucketResponse)
		if err != nil {
			t.Errorf("Unable to marshal fixture results, error: %s", err.Error())
			t.FailNow()
		}

		w.WriteHeader(200)
		w.Header().Set("content-type", "application/json")

		s, _ := strconv.Unquote(string(jsonBytes))
		_, _ = fmt.Fprintf(w, s)
	}

	mux.HandleFunc(workspaceMembersURLPath, handler)

	var members bitbucketMembers

	exBytes := []byte(bitbucketResponse)
	err := json.Unmarshal(exBytes, &members)
	assert.NoError(err, "error unmarshalling fixtures")
	want := members.Values[:1]

	var api bitbucketAPI
	api.BaseURL = server.URL
	got, err := api.getNon2svWorkspaceMembers()
	assert.NoError(err)
	assert.Equal(want, got, "bad struct results")
}
