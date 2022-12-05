package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"text/template"
	"time"

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
	app.Run(os.Args)
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

	source := rand.NewSource(time.Now().UnixNano())

	// copy the people into a "from" slice and a "to" slice
	froms := make([]mail.Address, len(people))
	tos := make([]mail.Address, len(people))
	copy(froms, people)
	copy(tos, people)

	// This will always happen at least once
	for !validateShuffle(froms, tos) {
		log.Info("Shuffling...")
		shuffle(&froms, source)
		shuffle(&tos, source)
	}

	if c.Bool("dry-run") {
		return sendDryRun(c, froms, tos)
	}

	return sendReal(c, froms, tos)
}

func sendDryRun(c *cli.Context, froms []mail.Address, tos []mail.Address) error {
	for i := 0; i < len(froms); i++ {
		fields := log.Fields{"from": froms[i].Name, "to": tos[i].Name}
		log.WithFields(fields).Info("match")
	}
	return nil
}

func sendReal(c *cli.Context, froms []mail.Address, tos []mail.Address) error {
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
		msg, err := formatMessage(c, body, from, froms[i], tos[i])
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

func formatMessage(c *cli.Context, body *template.Template, from mail.Address, giver mail.Address, receiver mail.Address) (*gophermail.Message, error) {
	data := struct {
		From string
		To   string
	}{
		From: giver.Name,
		To:   receiver.Name,
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

func validateShuffle(froms []mail.Address, tos []mail.Address) bool {
	if len(froms) != len(tos) {
		return false
	}
	for i := 0; i < len(froms); i++ {
		if froms[i].Name == tos[i].Name {
			return false
		}
	}
	return true
}

func loadPeople(filename string) ([]mail.Address, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var people []mail.Address
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&people)
	return people, err
}

func shuffle(array *[]mail.Address, source rand.Source) {
	random := rand.New(source)
	for i := len(*array) - 1; i > 0; i-- {
		j := random.Intn(i + 1)
		(*array)[i], (*array)[j] = (*array)[j], (*array)[i]
	}
}
