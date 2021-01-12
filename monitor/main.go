package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const countPerPage = 500

var workspaceMembersURLPath = "2.0/users/{workspace}/members"

const (
	envAPIBaseURL     = "API_BASE_URL"
	envAPIUsername    = "API_USERNAME"
	envAPIAppPassword = "API_APP_PASSWORD"
	envAPIWorkspace   = "API_WORKSPACE"
	envAWSS3Bucket    = "AWS_S3_BUCKET"
)

type bitbucketMember struct {
	DisplayName   string `json:"display_name"`
	Nickname      string `json:"nickname"`
	Has2faEnabled *bool  `json:"has_2fa_enabled"`
}

type bitbucketMembers struct {
	Values []bitbucketMember `json:"values"`
	Size   int               `json:"size"`
}

func (members *bitbucketMembers) getNon2svMembers() []bitbucketMember {
	var non2svMembers []bitbucketMember

	for _, member := range members.Values {
		if member.Has2faEnabled == nil {
			non2svMembers = append(non2svMembers, member)
		}
	}

	return non2svMembers
}

type lambdaConfig struct {
	APIBaseURL     string `json:"APIBaseURL"`
	APIUsername    string `json:"APIUsername"`
	APIAppPassword string `json:"APIAppPassword"`
	APIWorkspace   string `json:"APIWorkspace"`
	AWSS3Bucket    string `json:"AWSS3Bucket"`
	AWSS3Filename  string `json:"AWSS3FileName"`
}

func (c *lambdaConfig) init() error {
	if err := getRequiredString(envAPIBaseURL, &c.APIBaseURL); err != nil {
		return err
	}

	if err := getRequiredString(envAPIWorkspace, &c.APIWorkspace); err != nil {
		return err
	}
	workspaceMembersURLPath = strings.ReplaceAll(workspaceMembersURLPath, "{workspace}", c.APIWorkspace)

	if err := getRequiredString(envAPIUsername, &c.APIUsername); err != nil {
		return err
	}

	if err := getRequiredString(envAPIAppPassword, &c.APIAppPassword); err != nil {
		return err
	}

	if err := getRequiredString(envAWSS3Bucket, &c.AWSS3Bucket); err != nil {
		return err
	}

	return nil
}

func getRequiredString(envKey string, configEntry *string) error {
	if *configEntry != "" {
		return nil
	}

	value := os.Getenv(envKey)
	if value == "" {
		return fmt.Errorf("required value missing for environment variable %s", envKey)
	}
	*configEntry = value

	return nil
}

func callAPI(urlPath string, config lambdaConfig, queryParams map[string]string) (*http.Response, error) {
	var err error
	var req *http.Request

	url := config.APIBaseURL + "/" + urlPath

	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing http request: %s", err)
	}

	req.SetBasicAuth(config.APIUsername, config.APIAppPassword)
	req.Header.Set("Accept", "application/json")

	// Add query parameters
	q := req.URL.Query()
	for key, val := range queryParams {
		q.Add(key, val)
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error making http request: %s", err)
	} else if resp.StatusCode >= 300 {
		err := fmt.Errorf("API returned an error. URL: %s, Code: %v, Status: %s Body: %s",
			url, resp.StatusCode, resp.Status, resp.Body)
		return nil, err
	}

	return resp, nil
}

func getWorkspaceMembersPage(pageNum int, config lambdaConfig) (*bitbucketMembers, error) {
	queryParams := map[string]string{
		"per_page": strconv.Itoa(countPerPage),
		"page":     strconv.Itoa(pageNum),
	}

	// Make http call
	resp, err := callAPI(workspaceMembersURLPath, config, queryParams)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	var pageMembers bitbucketMembers
	// fmt.Println("\n", string(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &pageMembers); err != nil {
		return nil, fmt.Errorf("error decoding response json for workspace members: %s", err)
	}

	return &pageMembers, nil
}

func getNon2svWorkspaceMembers(config lambdaConfig) ([]bitbucketMember, error) {
	var allMembers []bitbucketMember

	for i := 1; ; i++ {
		members, err := getWorkspaceMembersPage(i, config)
		if err != nil {
			fmt.Println("\n", err)

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

func handler(config lambdaConfig) error {
	if err := config.init(); err != nil {
		return err
	}

	members, err := getNon2svWorkspaceMembers(config)
	if err != nil {
		return err
	}

	if len(members) > 0 {
		subject := fmt.Sprintf("%d %s members do not have 2SV enabled", len(members), config.APIWorkspace)
		var message string
		for _, member := range members {
			message += member.DisplayName + " - " + member.Nickname + "\n"
		}

		// TODO: Add SES mailing capabilities
		// var c MailConfig
		// SendEmail(c, message)

		fmt.Printf("%v:\n", subject)
		fmt.Println(message)
	}

	return nil
}

func manualRun() {
	var config lambdaConfig
	if err := config.init(); err != nil {
		panic("error initializing config ... " + err.Error())
	}

	if err := handler(config); err != nil {
		panic("error calling handler ... " + err.Error())
	}

	fmt.Printf("Success!\n")
}

func main() {
	//lambda.Start(handler)
	manualRun()

	// 	foo := `{"pagelen":50,"values":[{"display_name":"foo","has_2fa_enabled":null,"links":{"hooks":{"href":"foo"},"self":{"href":"foo"},"repositories":{"href":"foo"},"html":{"href":"foo"},"avatar":{"href":"foo"},"snippets":{"href":"foo"}},"nickname":"foo","zoneinfo":null,"account_id":"foo","department":null,"created_on":"foo","is_staff":false,"location":null,"account_status":"foo","organization":null,"job_title":"foo","type":"foo","properties":{},"uuid":"foo"}],"page":1,"size":27}`
	// 	bar := []byte(foo)
	// 	d, e := json.Unmarshal(bar, BitbucketMembers)
}
