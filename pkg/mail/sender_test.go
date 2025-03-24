package mail

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jordan-wright/email"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hankimmy/PtmrBackend/pkg/util"
)

type MockGmailSender struct {
	lastEmail *email.Email
	sendError error
}

func (m *MockGmailSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
	inlineFiles map[string]string,
) error {
	// Create a new email message
	e := email.NewEmail()
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	// Mock attaching files
	for _, file := range attachFiles {
		e.Attach(bytes.NewReader([]byte("mock content")), file, "text/plain")
	}

	// Mock inline files
	for cid, filePath := range inlineFiles {
		e.Attach(bytes.NewReader([]byte("mock inline content")), filePath, "image/png")
		// Set mock headers
		lastAttachment := e.Attachments[len(e.Attachments)-1]
		lastAttachment.Header.Set("Content-ID", "<"+cid+">")
		lastAttachment.Header.Set("Content-Disposition", "inline")
	}

	// Store the last email object for validation
	m.lastEmail = e

	// Return mock error if set
	return m.sendError
}

func TestSendEmailWithInlineAttachments(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "Test Email"
	htmlContent := `
		<html>
			<body>
				<p>Hello, this is a test email!</p>
				<img src="cid:logo" alt="logo">
			</body>
		</html>
	`
	to := []string{"hankim0572@gmail.com"}
	workingDir, err := os.Getwd()
	fmt.Println(workingDir)
	assert.NoError(t, err)
	inlineFiles := map[string]string{
		"logo": filepath.Join(workingDir, "assets", "images", "logo.png"),
	}

	// Act
	err = sender.SendEmail(
		subject,
		htmlContent,
		to,
		nil,
		nil,
		nil,
		inlineFiles,
	)
	assert.NoError(t, err)
}
