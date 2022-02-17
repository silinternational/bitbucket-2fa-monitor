package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
)

const (
	envSep = ","

	envAPIBaseURL     = "API_BASE_URL"
	envAPIUsername    = "API_USERNAME"
	envAPIAppPassword = "API_APP_PASSWORD"
	envAPIWorkspace   = "API_WORKSPACE"

	envSesCharset         = "SES_CHARSET"
	envSesReturnToAddress = "SES_RETURN_TO_ADDRESS"
	envSesRecipientEmails = "SES_RECIPIENT_EMAILS"
)

var debug = getEnvironmentVariableAsBool("DEBUG")

type appContext struct {
	API    bitbucketAPI `json:"BitbucketAPI"`
	Mailer mail         `json:"MailConfig"`
}

func (app *appContext) init() {
	app.API = bitbucketAPI{
		BaseURL:     getRequiredEnvironmentVariable(envAPIBaseURL),
		Workspace:   getRequiredEnvironmentVariable(envAPIWorkspace),
		Username:    getRequiredEnvironmentVariable(envAPIUsername),
		AppPassword: getRequiredEnvironmentVariable(envAPIAppPassword),
	}
	if !debug {
		app.Mailer = mail{
			CharSet:         getRequiredEnvironmentVariable(envSesCharset),
			ReturnToAddr:    getRequiredEnvironmentVariable(envSesReturnToAddress),
			RecipientEmails: getRequiredEnvironmentVariableAsSlice(envSesRecipientEmails, envSep),
		}
	}

	workspaceMembersURLPath = strings.ReplaceAll(workspaceMembersURLPath, "{workspace}", app.API.Workspace)
}

func getEnvironmentVariableAsBool(key string) bool {
	value := os.Getenv(key)
	result, err := strconv.ParseBool(value)
	if err != nil {
		result = false
	}
	return result
}

func getRequiredEnvironmentVariable(key string) string {
	value := os.Getenv(key)
	if value == "" {
		os.Stderr.WriteString("Required value missing for environment variable: " + key + "\n")
		os.Exit(1)
	}

	return value
}

func getRequiredEnvironmentVariableAsSlice(key string, sep string) []string {
	value := getRequiredEnvironmentVariable(key)
	return strings.Split(value, sep)
}

func handler(app appContext) error {
	app.init()

	members, err := app.API.getNon2svWorkspaceMembers()
	if err != nil {
		return err
	}

	if len(members) > 0 {
		app.Mailer.SubjectText = fmt.Sprintf("%d %s members do not have 2SV enabled", len(members), app.API.Workspace)
		var message string
		for _, member := range members {
			message += member.DisplayName + " - " + member.Nickname + "\n"
		}

		if debug {
			fmt.Printf("%v:\n", app.Mailer.SubjectText)
			fmt.Println(message)
		} else {
			app.Mailer.sendEmail(message)
		}
	}

	return nil
}

func main() {
	if debug {
		var app appContext
		if err := handler(app); err != nil {
			panic("error calling handler ... " + err.Error())
		}

		fmt.Printf("Success!\n")
	} else {
		lambda.Start(handler)
	}
}
