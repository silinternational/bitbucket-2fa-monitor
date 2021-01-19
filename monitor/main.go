package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
)

const (
	envAPIBaseURL     = "API_BASE_URL"
	envAPIUsername    = "API_USERNAME"
	envAPIAppPassword = "API_APP_PASSWORD"
	envAPIWorkspace   = "API_WORKSPACE"

	envAwsRegion          = "AWS_REGION"
	envAwsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	envAwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	envSesCharset         = "SES_CHARSET"
	envSesReturnToAddress = "SES_RETURN_TO_ADDRESS"
	envSesRecipientEmails = "SES_RECIPIENT_EMAILS"
)

type lambdaConfig struct {
	BitbucketAPI bitbucketAPI `json:"BitbucketAPI"`
	MailConfig   mail         `json:"MailConfig"`
	Debug        bool         `json:"Debug"`
}

func (c *lambdaConfig) init() error {
	if err := getRequiredString(envAPIBaseURL, &c.BitbucketAPI.BaseURL); err != nil {
		return err
	}

	if err := getRequiredString(envAPIWorkspace, &c.BitbucketAPI.Workspace); err != nil {
		return err
	}
	workspaceMembersURLPath = strings.ReplaceAll(workspaceMembersURLPath, "{workspace}", c.BitbucketAPI.Workspace)

	if err := getRequiredString(envAPIUsername, &c.BitbucketAPI.Username); err != nil {
		return err
	}

	if err := getRequiredString(envAPIAppPassword, &c.BitbucketAPI.AppPassword); err != nil {
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

func handler(config lambdaConfig) error {
	if err := config.init(); err != nil {
		return err
	}

	members, err := config.BitbucketAPI.getNon2svWorkspaceMembers()
	if err != nil {
		return err
	}

	if len(members) > 0 {
		config.MailConfig.SubjectText = fmt.Sprintf("%d %s members do not have 2SV enabled", len(members), config.BitbucketAPI.Workspace)
		var message string
		for _, member := range members {
			message += member.DisplayName + " - " + member.Nickname + "\n"
		}

		if config.Debug {
			fmt.Printf("%v:\n", config.MailConfig.SubjectText)
			fmt.Println(message)
		} else {
			config.MailConfig.sendEmail(message)
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
	lambda.Start(handler)
	//manualRun()
}
