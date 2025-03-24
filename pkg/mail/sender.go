package mail

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"

	"github.com/jordan-wright/email"
)

const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

type EmailSender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
		inlineFiles map[string]string, // cid -> filepath for inline images
	) error
}

type GmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

func NewGmailSender(name string, fromEmailAddress string, fromEmailPassword string) EmailSender {
	return &GmailSender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}

func (sender *GmailSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
	inlineFiles map[string]string,
) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.fromEmailAddress)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	// Attach regular files
	for _, f := range attachFiles {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", f, err)
		}
	}

	// Attach inline files (images with cid)
	for cid, filePath := range inlineFiles {
		// Read the file content
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read inline file %s: %w", filePath, err)
		}

		// Get the file extension (needed to set MIME type)
		ext := filepath.Ext(filePath)
		var mimeType string
		switch ext {
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		default:
			mimeType = "application/octet-stream"
		}

		// Attach file as inline using bytes.Reader and manually set MIME type
		_, err = e.Attach(bytes.NewReader(fileData), filepath.Base(filePath), mimeType)
		if err != nil {
			return fmt.Errorf("failed to attach inline file %s: %w", filePath, err)
		}

		// Create the Content-ID reference in the inline attachment (works with default Attach method)
		lastAttachment := e.Attachments[len(e.Attachments)-1]
		lastAttachment.Header.Set("Content-ID", fmt.Sprintf("<%s>", cid))
		lastAttachment.Header.Set("Content-Disposition", "inline")
		lastAttachment.Header.Set("Content-Transfer-Encoding", "base64")
	}

	// Set up SMTP authentication and send the email
	smtpAuth := smtp.PlainAuth("", sender.fromEmailAddress, sender.fromEmailPassword, smtpAuthAddress)
	return e.Send(smtpServerAddress, smtpAuth)
}
