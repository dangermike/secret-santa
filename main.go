package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"text/template"

	log "github.com/sirupsen/logrus"
	"gopkg.in/jpoehls/gophermail.v0"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Version = appVersion
	app.Usage = "Secret Santa Emailer!"
	app.Flags = appFlags
	app.Before = appBefore
	app.Action = appMain
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "error: %w", err)
		os.Exit(1)
	}
}

func appBefore(c *cli.Context) error {
	if !c.Bool("dry-run") {
		if !c.IsSet("from-address") {
			return errors.New("Parameter 'from-address' required except in dry run")
		}
	}
	return nil
}

func appMain(c *cli.Context) error {
	people, err := loadPeople(c.String("source-file"))
	if err != nil {
		log.WithError(err).Fatal("Decode error")
		return err
	}

	// copy the people into a "from" slice and a "to" slice
	froms := make([]Recipient, len(people))
	tos := make([]Recipient, len(people))
	copy(froms, people)
	copy(tos, people)

	for _, r := range froms {
		log.Infof("recipient: %s <%s> (%v)", r.Name, r.Address.Address, r.Extras)
	}

	log.Info("Shuffling...")
	rand.Shuffle(len(tos), func(i, j int) {
		tos[i], tos[j] = tos[j], tos[i]
	})

	var valid bool

	for !valid {
		log.Info("validating...")
		valid = true
		for i := 0; i < len(tos); i++ {
			if tos[i].Address == froms[i].Address {
				s := (i + 1) % len(tos)
				tos[i], tos[s] = tos[s], tos[i]
				valid = false
			}
		}
	}

	if c.Bool("dry-run") {
		return sendDryRun(c, froms, tos)
	}

	return sendReal(c, froms, tos)
}

func sendDryRun(c *cli.Context, froms []Recipient, tos []Recipient) error {
	from := mail.Address{Name: "dry_run", Address: "secret-santa@example.com"}
	if c.IsSet("from-name") {
		from.Name = c.String("from-name")
	}
	if c.IsSet("from-address") {
		from.Address = c.String("from-address")
	}

	var body *template.Template
	if c.IsSet("template-file") {
		var err error
		body, err = template.New(filepath.Base(c.String("template-file"))).ParseFiles(c.String("template-file"))
		if err != nil {
			log.WithError(err).Fatal("Decode error")
			return err
		}
	}
	for i := 0; i < len(froms); i++ {
		fields := log.Fields{"from": froms[i].Name, "to": tos[i].Name}
		if body != nil {
			msg, err := formatMessage(c, body, from, froms[i].Address, tos[i].Address, tos[i].Extras)
			if err != nil {
				return fmt.Errorf("failed to format message: %w", err)
			}
			fields["email.from"] = msg.From
			fields["email.subj"] = msg.Subject
			fields["email.body"] = msg.Body
		}
		log.WithFields(fields).Info("match")
	}

	return nil
}

func sendReal(c *cli.Context, froms []Recipient, tos []Recipient) error {
	body, err := template.New(filepath.Base(c.String("template-file"))).ParseFiles(c.String("template-file"))
	if err != nil {
		log.WithError(err).Fatal("Decode error")
		return err
	}

	auth := smtp.PlainAuth(
		"",
		c.String("from-address"),
		c.String("from-password"),
		c.String("smtp-host"),
	)
	from := mail.Address{Name: c.String("from-name"), Address: c.String("from-address")}

	for i := 0; i < len(froms); i++ {
		msg, err := formatMessage(c, body, from, froms[i].Address, tos[i].Address, tos[i].Extras)
		if err != nil {
			log.WithError(err).Fatal("ERROR: attempting to format mail ")
			return err
		}

		err = gophermail.SendMail(
			fmt.Sprintf("%s:%d", c.String("smtp-host"), c.Int("smtp-port")),
			auth,
			msg,
		)
		if err != nil {
			log.WithError(err).Fatal("ERROR: attempting to send a mail ")
			return err
		}
		fields := log.Fields{}
		if c.Bool("show-matches") {
			fields = log.Fields{"from": froms[i].Name, "to": tos[i].Name}
		}
		log.WithFields(fields).Info("sent")
	}
	return nil
}

func formatMessage(c *cli.Context, body *template.Template, from mail.Address, giver mail.Address, receiver mail.Address, extras map[string]string) (*gophermail.Message, error) {
	data := struct {
		From   string
		To     string
		Extras map[string]string
	}{
		From:   giver.Name,
		To:     receiver.Name,
		Extras: extras,
	}

	var tpl bytes.Buffer
	if err := body.Execute(&tpl, data); err != nil {
		return nil, err
	}

	msg := gophermail.Message{
		From:    from,
		To:      []mail.Address{giver},
		Subject: c.String("subject"),
		Body:    tpl.String(),
	}

	return &msg, nil
}

func loadPeople(filename string) ([]Recipient, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var people []Recipient
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&people)
	return people, err
}

type Recipient struct {
	mail.Address
	Extras map[string]string
}
