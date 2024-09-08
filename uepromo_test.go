package imapgoprebuilt

import (
	"log"
	"testing"
)

func TestUberEatsPromo(t *testing.T) {
	imapOpts := &ImapOpts{
		Imap:          ICloud,
		Site:          UberEatsPromo,
		ReceiverEmail: "",
		ReceiverPass:  "",
		CatchallEmail: "",
		CatchallPass:  "",
		MaxChecks:     5,

		ReceiverEmails: map[string]map[string]string{},
	}
	code, err := imapOpts.FetchEmailForMultipleAccounts()
	log.Println(code, err)
}
