package imapgoprebuilt

type ImapOpts struct {
	Imap          *EmailOpts
	Site          string
	ReceiverEmail string
	ReceiverPass  string
	CatchallEmail string
	CatchallPass  string
	MaxChecks     int

	// Multiple Accounts only
	ReceiverEmails map[string]map[string]string
}

type EmailOpts struct {
	Email string
	Imap  string
}
