package imapgoprebuilt

import (
	"log"
	"testing"
)

func TestUberEatsPromo(t *testing.T) {
	imapOpts := &ImapOpts{
		Imap:          Gmail,
		Site:          UberEatsPromo,
		ReceiverEmail: "",
		ReceiverPass:  "",
		CatchallEmail: "",
		CatchallPass:  "",
		MaxChecks:     5,
		Days:          4,

		ReceiverEmails: map[string]map[string]string{},
	}
	promos, err := imapOpts.FetchEmailForMultipleAccounts()
	if err != nil {
		t.Fatalf("Failed to fetch UberEats promo: %v", err)
	}
	for email, list := range promos {
		log.Printf("Email: %s\n", email)
		for _, promo := range list {
			log.Printf(" %s - %s\n", promo["promoCode"], promo["promoType"])
		}
	}
}
