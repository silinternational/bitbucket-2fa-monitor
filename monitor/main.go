package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

const countPerPage = 500

var workspaceMembersURLPath = "/2.0/users/{workspace}/members"

const (
	envAPIBaseURL         = "API_BASE_URL"
	envAPIUsername        = "API_USERNAME"
	envAPIAppPassword     = "API_APP_PASSWORD"
	envAPIWorkspace       = "API_WORKSPACE"
	envAwsRegion          = "AWS_REGION"
	envAwsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	envAwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	envSesCharset         = "SES_CHARSET"
	envSesReturnToAddress = "SES_RETURN_TO_ADDRESS"
	envSesRecipientEmails = "SES_RECIPIENT_EMAILS"
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
	APIBaseURL     string     `json:"APIBaseURL"`
	APIUsername    string     `json:"APIUsername"`
	APIAppPassword string     `json:"APIAppPassword"`
	APIWorkspace   string     `json:"APIWorkspace"`
	MailConfig     mailConfig `json:"MailConfig"`
	SESRecipients  string     `json:"SESRecipients"`
	Debug          bool       `json:"Debug"`
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

	if err := getRequiredString(envSesCharset, &c.MailConfig.CharSet); err != nil {
		return err
	}

	if err := getRequiredString(envSesReturnToAddress, &c.MailConfig.ReturnToAddr); err != nil {
		return err
	}

	var emails string
	if err := getRequiredString(envSesRecipientEmails, &emails); err != nil {
		return err
	}
	c.MailConfig.RecipientEmails = strings.Split(emails, " ")

	if err := getRequiredString(envAwsRegion, &c.MailConfig.AWSRegion); err != nil {
		return err
	}

	if err := getRequiredString(envAwsAccessKeyID, &c.MailConfig.AWSAccessKeyID); err != nil {
		return err
	}

	if err := getRequiredString(envAwsSecretAccessKey, &c.MailConfig.AWSSecretAccessKey); err != nil {
		return err
	}

	var err error
	debug := os.Getenv("DEBUG")
	c.Debug, err = strconv.ParseBool(debug)
	if err != nil {
		c.Debug = false
	}

	return nil
}

type mailConfig struct {
	AWSRegion          string
	CharSet            string
	ReturnToAddr       string
	SubjectText        string
	RecipientEmails    []string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
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

	url := config.APIBaseURL + urlPath

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

func sendEmail(config mailConfig, body string) {
	charSet := config.CharSet

	subject := config.SubjectText
	subjContent := ses.Content{
		Charset: &charSet,
		Data:    &subject,
	}

	msgContent := ses.Content{
		Charset: &charSet,
		Data:    &body,
	}

	msgBody := ses.Body{
		Text: &msgContent,
	}

	emailMsg := ses.Message{}
	emailMsg.SetSubject(&subjContent)
	emailMsg.SetBody(&msgBody)

	// Only report the last email error
	lastError := ""
	badRecipients := []string{}

	// Send emails to one recipient at a time to avoid one bad email sabotaging it all
	for _, address := range config.RecipientEmails {
		err := sendAnEmail(emailMsg, address, config)
		if err != nil {
			lastError = err.Error()
			badRecipients = append(badRecipients, address)
		}
	}

	if lastError != "" {
		addresses := strings.Join(badRecipients, ", ")
		log.Printf("Error sending Bitbucket 2FA monitor email from %s to: %s\n %s",
			config.ReturnToAddr, addresses, lastError)
	}
}

func sendAnEmail(emailMsg ses.Message, recipient string, config mailConfig) error {
	recipients := []*string{&recipient}

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: recipients,
		},
		Message: &emailMsg,
		Source:  aws.String(config.ReturnToAddr),
	}

	cfg := &aws.Config{Region: aws.String(config.AWSRegion)}
	if config.AWSAccessKeyID != "" && config.AWSSecretAccessKey != "" {
		cfg.Credentials = credentials.NewStaticCredentials(config.AWSAccessKeyID, config.AWSSecretAccessKey, "")
	}
	sess, err := session.NewSession(cfg)
	if err != nil {
		return fmt.Errorf("error creating AWS session: %s", err)
	}

	svc := ses.New(sess)
	result, err := svc.SendEmail(input)
	if err != nil {
		return fmt.Errorf("error sending email, result: %s, error: %s", result, err)
	}
	log.Printf("alert message sent to %s, message ID: %s", recipient, *result.MessageId)
	return nil
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
		config.MailConfig.SubjectText = fmt.Sprintf("%d %s members do not have 2SV enabled", len(members), config.APIWorkspace)
		var message string
		for _, member := range members {
			message += member.DisplayName + " - " + member.Nickname + "\n"
		}

		if config.Debug {
			fmt.Printf("%v:\n", config.MailConfig.SubjectText)
			fmt.Println(message)
		} else {
			sendEmail(config.MailConfig, message)
		}
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
}
