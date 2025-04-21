package imapgoprebuilt

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
)

// get promo code from ubereats with optional days filtering and optional receiver emails filtering
func (n *ImapOpts) getUberEatsPromo() (map[string][]map[string]string, error) {
	c, err := client.DialTLS(n.Imap.Imap, nil)
	if err != nil {
		return nil, errors.New("could not connect to mail server")
	}
	defer c.Logout()

	if n.CatchallPass == "" {
		if err := c.Login(n.ReceiverEmail, n.ReceiverPass); err != nil {
			return nil, fmt.Errorf("login failed: %v", err)
		}
	} else {
		if err := c.Login(n.CatchallEmail, n.CatchallPass); err != nil {
			return nil, fmt.Errorf("login failed: %v", err)
		}
	}

	criteria := &imap.SearchCriteria{}
	if n.Days > 0 {
		criteria.Since = time.Now().AddDate(0, 0, -n.Days)
	}

	mailboxes := []string{"INBOX"}
	results := make(map[string][]map[string]string)

	for _, box := range mailboxes {
		_, err := c.Select(box, false)
		if err != nil {
			continue
		}

		ids, err := c.Search(criteria)
		if err != nil || len(ids) == 0 {
			continue
		}

		seqSet := new(imap.SeqSet)
		seqSet.AddNum(ids...)

		section := &imap.BodySectionName{}
		items := []imap.FetchItem{section.FetchItem()}

		messages := make(chan *imap.Message, 8)
		go func() {
			c.Fetch(seqSet, items, messages)
		}()

		for msg := range messages {
			if msg == nil {
				continue
			}

			r := msg.GetBody(section)
			if r == nil {
				continue
			}

			mr, err := mail.CreateReader(r)
			if err != nil {
				continue
			}

			header := mr.Header
			subject, err := header.Subject()
			if err != nil || !strings.Contains(strings.ToLower(subject), "$") {
				continue
			}

			from, err := header.AddressList("From")
			if err != nil || len(from) == 0 || !strings.Contains(strings.ToLower(from[0].Address), "uber@uber.com") {
				continue
			}

			to, err := header.AddressList("To")
			if err != nil || len(to) == 0 {
				continue
			}

			address := strings.Trim(to[0].Address, "<>")

			if len(n.ReceiverEmails) > 0 {
				if _, exists := n.ReceiverEmails[strings.ToLower(address)]; !exists {
					continue
				}
			}

			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					break
				}

				if _, ok := p.Header.(*mail.InlineHeader); ok {
					b, _ := io.ReadAll(p.Body)
					pattern := `(?is)<strong[^>]*>\s*([A-Za-z0-9]+)\s*</strong>`
					re := regexp.MustCompile(pattern)
					match := re.FindStringSubmatch(string(b))
					if len(match) > 1 {
						promo := map[string]string{
							"promoType": subject,
							"promoCode": match[1],
						}
						results[address] = append(results[address], promo)
					}
				}
			}
		}
	}

	return results, nil
}
