To install
`go get github.com/Aspect-af/imap-go-prebuilt`

Use case - retrieve nike login code (if you don't add catchall it will use receiver details, if you add catchall information, it will use catchall and retrieve for the receiver)
```go
	imapOpts := &imapgoprebuilt.ImapOpts{
		Imap:          imapgoprebuilt.Gmail,
		Site:          imapgoprebuilt.Nike,
		ReceiverEmail: "EMAIL@gmail.com",
		ReceiverPass:  "APP_PASSWORD",
		CatchallEmail: "",
		CatchallPass:  "",
		MaxChecks:     5,
	}
	if code, err := imapOpts.FetchEmail(); err != nil {
		log.Println(err)
	} else {
		log.Println(code)
	}
```