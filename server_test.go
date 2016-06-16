package smtpd_test

import (
	"fmt"
	"net/smtp"
	"testing"
	"time"

	"github.com/mailproto/smtpd"
)

type MessageRecorder struct {
	Messages []*smtpd.Message
}

func (m *MessageRecorder) Record(msg *smtpd.Message) error {
	m.Messages = append(m.Messages, msg)
	return nil
}

func TestSMTPServer(t *testing.T) {

	recorder := &MessageRecorder{}
	server := smtpd.NewServer(recorder.Record)
	go server.ListenAndServe("localhost:0")
	defer server.Close()

	WaitUntilAlive(server)

	// Connect to the remote SMTP server.
	c, err := smtp.Dial(server.Address())
	if err != nil {
		t.Errorf("Should be able to dial localhost: %v", err)
	}

	// Set the sender and recipient first
	if err := c.Mail("sender@example.org"); err != nil {
		t.Errorf("Should be able to set a sender: %v", err)
	}
	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("Should be able to set a RCPT: %v", err)
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Error creating the data body: %v", err)
	}
	_, err = fmt.Fprintf(wc, `To: sender@example.org
From: recipient@example.net
Content-Type: text/html

This is the email body`)
	if err != nil {
		t.Errorf("Error writing email: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Error(err)
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		t.Errorf("Server wouldn't accept QUIT: %v", err)
	}

}

func TestSMTPServerTimeout(t *testing.T) {

	recorder := &MessageRecorder{}
	server := smtpd.NewServer(recorder.Record)

	// Set some really short timeouts
	server.ReadTimeout = time.Millisecond * 1
	server.WriteTimeout = time.Millisecond * 1

	go server.ListenAndServe("localhost:0")
	defer server.Close()

	WaitUntilAlive(server)

	// Connect to the remote SMTP server.
	c, err := smtp.Dial(server.Address())
	if err != nil {
		t.Errorf("Should be able to dial localhost: %v", err)
	}

	// Sleep for twice the timeout
	time.Sleep(time.Millisecond * 20)

	// Set the sender and recipient first
	if err := c.Hello("sender@example.org"); err == nil {
		t.Errorf("Should have gotten a timeout from the upstream server")
	}

}
