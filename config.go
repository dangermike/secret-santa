package main

import (
	"gopkg.in/urfave/cli.v1"
)

const appVersion = "0.0.1"

var appFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "from-name",
		Usage: "Display name for sender",
		Value: "Secret Santa Robotic Elf",
	},
	cli.StringFlag{
		Name:  "from-address",
		Usage: "Email address to send from (required unless dry run)",
	},
	cli.StringFlag{
		Name:  "from-password",
		Usage: "Password for the from-address (e.g. gmail password)",
	},
	cli.StringFlag{
		Name:  "smtp-host",
		Usage: "Host to send to",
		Value: "smtp.gmail.com",
	},
	cli.IntFlag{
		Name:  "smtp-port",
		Usage: "Port to send to",
		Value: 587,
	},
	cli.StringFlag{
		Name:  "source-file",
		Usage: "JSON file containing an array of name/address objects",
		Value: "data/people.json",
	},
	cli.BoolFlag{
		Name:  "show-matches",
		Usage: "Print the pairings out when sending",
	},
	cli.BoolFlag{
		Name:  "dry-run",
		Usage: "Do not send. Implies --show-matches",
	},
}
