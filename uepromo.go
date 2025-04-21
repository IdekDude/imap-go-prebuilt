package imapgoprebuilt

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
)

var (
	promoCodeRe    = regexp.MustCompile(`(?is)<strong[^>]*>\s*([A-Za-z0-9]+)\s*</strong>`)
	validUntilRe   = regexp.MustCompile(`(?i)Valid until (.*?)\.`)
	validForDaysRe = regexp.MustCompile(`(?i)valid for (\d+) days`)
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
					body := string(b)

					promoCode, err := extractPromoCode(body)
					if err != nil {
						continue
					}

					sentDate, _ := header.Date()

					promoExpiration, err := extractPromoExpiration(body, sentDate)
					if err != nil {
						continue
					}

					promo := map[string]string{
						"promoType":       subject,
						"promoCode":       promoCode,
						"promoExpiration": promoExpiration,
					}
					results[address] = append(results[address], promo)
				}
			}
		}
	}

	return results, nil
}

func extractPromoCode(body string) (string, error) {
	matches := promoCodeRe.FindStringSubmatch(body)
	if len(matches) < 2 {
		return "", errors.New("promo code not found")
	}
	return matches[1], nil
}

func extractPromoExpiration(body string, sentDate time.Time) (string, error) {
	// 1) “Valid until ….”
	if m := validUntilRe.FindStringSubmatch(body); len(m) > 1 {
		raw := m[1]
		// try parsing against common formats
		layouts := []string{
			"01/02/2006",      // MM/DD/YYYY
			"Jan 2, 2006",     // e.g. Jan 2, 2006
			"January 2, 2006", // e.g. January 2, 2006
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, raw); err == nil {
				return t.Format("01/02/2006"), nil
			}
		}
		// if we couldn’t parse, return the raw
		return raw, nil
	}

	// 2) “valid for N days”
	if m := validForDaysRe.FindStringSubmatch(body); len(m) > 1 {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			return "", fmt.Errorf("invalid days count %q: %w", m[1], err)
		}
		exp := sentDate.AddDate(0, 0, n)
		return exp.Format("01/02/2006"), nil
	}

	return "", errors.New("promo expiration not found")
}
