package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

type mail struct {
	AWSRegion       string
	CharSet         string
	ReturnToAddr    string
	SubjectText     string
	RecipientEmails []string
}

func (m *mail) sendEmail(body string) {
	charSet := m.CharSet

	subject := m.SubjectText
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
	for _, address := range m.RecipientEmails {
		err := m.sendAnEmail(emailMsg, address)
		if err != nil {
			lastError = err.Error()
			badRecipients = append(badRecipients, address)
		}
	}

	if lastError != "" {
		addresses := strings.Join(badRecipients, ", ")
		log.Printf("Error sending Bitbucket 2FA monitor email from %s to: %s\n %s",
			m.ReturnToAddr, addresses, lastError)
	}
}

func (m *mail) sendAnEmail(emailMsg ses.Message, recipient string) error {
	recipients := []*string{&recipient}

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: recipients,
		},
		Message: &emailMsg,
		Source:  aws.String(m.ReturnToAddr),
	}

	sess, err := session.NewSession()
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
