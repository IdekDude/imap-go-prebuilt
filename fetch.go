package imapgoprebuilt

import (
	"time"
)

// fetch the email information with the prebuilt functions
func (n *ImapOpts) FetchEmail() (string, error) {
	var message string
	var err error
	switch n.Site {
	case Nike:
		for i := 1; i < n.MaxChecks; i++ {
			message, err = n.getNikeLoginCode()
			if err == nil {
				break
			}
			time.Sleep(5 * time.Second)
		}
	case TicketMaster:
		for i := 1; i < n.MaxChecks; i++ {
			message, err = n.getTicketMasterMFA()
			if err == nil {
				break
			}
			time.Sleep(5 * time.Second)
		}
	case Walmart:
		for i := 1; i < n.MaxChecks; i++ {
			message, err = n.getWalmartMFA()
			if err == nil {
				break
			}
			time.Sleep(5 * time.Second)
		}
	case Footsites:
		for i := 1; i < n.MaxChecks; i++ {
			message, err = n.getFLXActivationLink()
			if err == nil {
				break
			}
			time.Sleep(5 * time.Second)
		}
	case UberEats:
		for i := 1; i < n.MaxChecks; i++ {
			message, err = n.getUberEatsCode()
			if err == nil {
				break
			}
			time.Sleep(5 * time.Second)
		}
	}
	return message, err
}

// fetch the email information with the prebuilt functions for multiple accounts
func (n *ImapOpts) FetchEmailForMultipleAccounts() (map[string]map[string]string, error) {
	var messages map[string]map[string]string
	var err error
	switch n.Site {
	case UberEatsPromo:
		for i := 1; i < n.MaxChecks; i++ {
			messages, err = n.getUberEatsPromo()
			if err == nil {
				break
			}
			time.Sleep(5 * time.Second)
		}
	}
	return messages, err
}
