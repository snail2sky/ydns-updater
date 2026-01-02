package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	ydns "github.com/wyattjoh/ydns-updater/internal"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	app := cli.NewApp()
	app.Name = "ydns-updater"
	app.Version = fmt.Sprintf("%v, commit %v, built at %v", version, commit, date)
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "base",
			Value: "https://ydns.io/api/v1/update/",
			Usage: "base url for api calls on ydns",
		},
		&cli.StringFlag{
			Name:     "host",
			EnvVars:  []string{"YDNS_HOST"},
			Required: true,
			Usage:    "host to update",
		},
		&cli.StringFlag{
			Name:    "ip",
			EnvVars: []string{"YDNS_IP"},
			Usage:   "ip to update",
		},
		&cli.StringFlag{
			Name:    "record_id",
			EnvVars: []string{"YDNS_RECORD_ID"},
			Usage:   "record_id to update",
		},
		&cli.StringFlag{
			Name:     "user",
			EnvVars:  []string{"YDNS_USER"},
			Required: true,
			Usage:    "username for authentication on ydns",
		},
		&cli.StringFlag{
			Name:     "pass",
			EnvVars:  []string{"YDNS_PASS"},
			Required: true,
			Usage:    "password for authentication on ydns",
		},
		&cli.BoolFlag{
			Name:    "daemon",
			EnvVars: []string{"YDNS_DAEMON"},
			Usage:   "enables the updater as a daemon",
		},
		&cli.DurationFlag{
			Name:    "frequency",
			EnvVars: []string{"YDNS_FREQUENCY"},
			Value:   60 * time.Minute,
			Usage:   "sleep time between updates while in daemon mode",
		},
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "enables debug logging",
		},
		&cli.StringFlag{
			Name:    "family",
			EnvVars: []string{"YDNS_FAMILY"},
			Value:   "any",
			Usage:   "force IP family for outgoing requests: ipv4|ipv6|any",
		},
	}
	app.Action = func(c *cli.Context) error {
		base := c.String("base")
		host := c.String("host")
		ip := c.String("ip")
		record_id := c.String("record_id")
		user := c.String("user")
		pass := c.String("pass")
		family := c.String("family")
		daemon := c.Bool("daemon")
		frequency := c.Duration("frequency")

		if c.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if err := ydns.Run(base, host, ip, record_id, user, pass, family); err != nil {
			return cli.Exit(err, 1)
		}

		for daemon {
			logrus.WithField("sleep", frequency).Info("sleeping till next update")
			time.Sleep(frequency)

			if err := ydns.Run(base, host, ip, record_id, user, pass, family); err != nil {
				return cli.Exit(err, 1)
			}
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Fatal()
	}
}
