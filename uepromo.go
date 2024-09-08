package imapgoprebuilt

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
)

// get promo code from ubereats with multiple emails
func (n *ImapOpts) getUberEatsPromo() (map[string]map[string]string, error) {
	// Connect to the server
	c, err := client.DialTLS(n.Imap.Imap, nil)
	if err != nil {
		return nil, errors.New("could not connect to mail server")
	}
	defer c.Logout()

	// handle login
	if n.CatchallPass == "" {
		if err := c.Login(n.ReceiverEmail, n.ReceiverPass); err != nil {
			return nil, fmt.Errorf("login and password are incorrect: %s:%s - %s", n.ReceiverEmail, n.ReceiverPass, err.Error())
		}
	} else {
		if err := c.Login(n.CatchallEmail, n.CatchallPass); err != nil {
			return nil, fmt.Errorf("login and password are incorrect: %s:%s - %s", n.CatchallEmail, n.CatchallPass, err.Error())
		}
	}

	// now we grab our mails
	var boxes []string
	mailboxes := make(chan *imap.MailboxInfo, 5)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	for m := range mailboxes {
		if m.Name == "[Gmail]/Important" {
			continue
		}
		boxes = append(boxes, m.Name)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("login and password are incorrect: %s:%s - %s", n.CatchallEmail, n.CatchallPass, err.Error())
	}

	for _, box := range boxes {
		// Select INBOX
		mbox, err := c.Select(box, false)
		if err != nil {
			continue
		}

		// Get the last message
		if mbox.Messages == 0 {
			continue
		}

		var to, from uint32
		if mbox.Messages > 30 {
			from = mbox.Messages
			to = mbox.Messages - 30
		} else {
			from = mbox.Messages
			to = 0
		}

		seqSet := new(imap.SeqSet)
		seqSet.AddRange(from, to)

		// Get the whole message body
		var section imap.BodySectionName
		items := []imap.FetchItem{section.FetchItem()}

		messages := make(chan *imap.Message, 8)

		go func() {
			c.Fetch(seqSet, items, messages)
		}()

		var address, fromaddress, mailsubject string

		for msg := range messages {

			// If the message is null or if the activation email was found then skip the email
			if msg == nil {
				continue
			}

			r := msg.GetBody(&section)
			if r == nil {
				continue
			}

			// Create a new mail reader
			mr, err := mail.CreateReader(r)
			if err != nil {
				continue
			}

			// Print some info about the message
			header := mr.Header

			if subject, err := header.Subject(); err == nil {
				mailsubject = strings.ToLower(subject)
			}

			if !strings.Contains(mailsubject, "$") {
				continue
			}

			if from, err := header.AddressList("From"); err == nil {
				fromaddress = from[0].String()
			}

			if !strings.Contains(strings.ToLower(fromaddress), "uber@uber.com") {
				continue
			}

			if to, err := header.AddressList("To"); err == nil {
				if len(to) == 0 {
					continue
				}
				address = strings.Trim(to[0].String(), "<>")
			}

			if _, exists := n.ReceiverEmails[strings.ToLower(strings.Trim(address, "<>"))]; exists {
				// if the email is in the map, we can now start reading the email
				for {
					p, err := mr.NextPart()
					if err == io.EOF {
						break
					} else if err != nil {
						break
					}

					switch p.Header.(type) {
					case *mail.InlineHeader:
						// This is the message's text (can be plain-text or HTML)
						b, _ := io.ReadAll(p.Body)
						pattern := `(?mU)promo code (\S*)\.`
						re := regexp.MustCompile(pattern)
						match := re.FindStringSubmatch(string(b))
						if strings.Contains(string(b), "promo") && len(match) > 1 {
							value := match[1]
							n.ReceiverEmails[address] = map[string]string{
								"promoType": mailsubject,
								"promoCode": value,
							}
						}
					default:
						continue
					}
				}
			}
		}
	}

	return n.ReceiverEmails, nil
}
