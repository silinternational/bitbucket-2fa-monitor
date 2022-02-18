package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

const bitbucketResponse = `{
  "values": [
    {
      "user": {
		"nickname": "example",
		"display_name": "Negative Test",
		"has_2fa_enabled": null
      }
    },
    {
      "user": {
		"nickname": "example",
		"display_name": "Positive Test",
		"has_2fa_enabled": true
      }
    }
  ]
}`

func TestGetNon2svWorkspaceMembers(t *testing.T) {
	assert := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprint(w, bitbucketResponse)
	}))

	var members bitbucketMembers

	exBytes := []byte(bitbucketResponse)
	err := json.Unmarshal(exBytes, &members)
	assert.NoError(err, "error unmarshalling fixtures")
	want := []bitbucketMember{members.Values[:1][0].User}

	var api bitbucketAPI
	api.BaseURL = server.URL
	got, err := api.getNon2svWorkspaceMembers()
	assert.NoError(err)
	assert.Equal(want, got, "bad struct results")
}
