package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const countPerPage = 500

var workspaceMembersURLPath = "/2.0/workspaces/{workspace}/members?fields=size,values.user.display_name,values.user.nickname,values.user.has_2fa_enabled"

type bitbucketMember struct {
	DisplayName   string `json:"display_name"`
	Nickname      string `json:"nickname"`
	Has2faEnabled *bool  `json:"has_2fa_enabled"`
}

type bitbucketMembers struct {
	Values []struct {
		User bitbucketMember `json:"user"`
	} `json:"values"`
	Size int `json:"size"`
}

func (members *bitbucketMembers) getNon2svMembers() []bitbucketMember {
	var non2svMembers []bitbucketMember

	for _, value := range members.Values {
		if value.User.Has2faEnabled == nil || *value.User.Has2faEnabled == false {
			non2svMembers = append(non2svMembers, value.User)
		}
	}

	return non2svMembers
}

type bitbucketAPI struct {
	BaseURL     string `json:"APIBaseURL"`
	Username    string `json:"APIUsername"`
	AppPassword string `json:"APIAppPassword"`
	Workspace   string `json:"APIWorkspace"`
}

func (api *bitbucketAPI) callAPI(urlPath string, queryParams map[string]string) ([]byte, error) {
	var err error
	var req *http.Request

	url := api.BaseURL + urlPath

	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing http request: %s", err)
	}

	req.SetBasicAuth(api.Username, api.AppPassword)
	req.Header.Set("Accept", "application/json")

	// Add query parameters
	q := req.URL.Query()
	for key, val := range queryParams {
		q.Add(key, val)
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Do(req)
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error making http request: %s", err)
	} else if resp.StatusCode >= 300 {
		err := fmt.Errorf("API returned an error. URL: %s, Code: %v, Status: %s Body: %s",
			url, resp.StatusCode, resp.Status, bodyBytes)
		return nil, err
	}

	return bodyBytes, nil
}

func (api *bitbucketAPI) getWorkspaceMembersPage(pageNum int) (*bitbucketMembers, error) {
	queryParams := map[string]string{
		"per_page": strconv.Itoa(countPerPage),
		"page":     strconv.Itoa(pageNum),
	}

	// Make http call
	bodyBytes, err := api.callAPI(workspaceMembersURLPath, queryParams)
	if err != nil {
		return nil, err
	}

	var pageMembers bitbucketMembers
	if err := json.Unmarshal(bodyBytes, &pageMembers); err != nil {
		return nil, fmt.Errorf("error decoding response json for workspace members: %s", err)
	}

	return &pageMembers, nil
}

func (api *bitbucketAPI) getNon2svWorkspaceMembers() ([]bitbucketMember, error) {
	var allMembers []bitbucketMember

	for i := 1; ; i++ {
		members, err := api.getWorkspaceMembersPage(i)
		if err != nil {
			err = fmt.Errorf("error fetching page %v ... %s", i, err)
			return nil, err
		}

		allMembers = append(allMembers, members.getNon2svMembers()...)

		if members.Size < countPerPage {
			break
		}
	}

	return allMembers, nil
}
