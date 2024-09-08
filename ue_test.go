package imapgoprebuilt

import (
	"log"
	"testing"
)

func TestUberEats(t *testing.T) {
	imapOpts := &ImapOpts{
		Imap:          ICloud,
		Site:          UberEats,
		ReceiverEmail: "",
		ReceiverPass:  "",
		CatchallEmail: "",
		CatchallPass:  "",
		MaxChecks:     5,
	}
	code, err := imapOpts.FetchEmail()
	log.Println(code, err)
}
