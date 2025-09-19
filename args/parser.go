package args

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/mamoutou92/lab/k3sutils"
	"github.com/urfave/cli/v2"
)

var CurrentConfig k3sutils.ExperimentConfig
var ExperimentParams *cli.App = &cli.App{
	EnableBashCompletion: true,
	HideHelp:             false,
	Name:                 "nimp2p-lab",
	Usage:                "A small agent for running multi-tenant large-scale reproducible experiments on kubernetes",
	Commands: []*cli.Command{
		{
			Name:     "create",
			Aliases:  []string{"r", "start", "run"},
			Usage:    "create experiments",
			HideHelp: false,

			Flags: []cli.Flag{
				&cli.StringFlag{Name: "experiment-name", Usage: "A user-friendly name for the experiment (e.g. blue) .", Required: true, Aliases: []string{"name", "n"}},
				&cli.IntFlag{Name: "num-peers", Value: 2, Usage: "Number of peers in the network.", Aliases: []string{"peers", "p"}},
				&cli.IntFlag{Name: "msg-rate", Value: 4, Usage: "Delay between messages in milliseconds."},
				&cli.IntFlag{Name: "msg-size", Value: 1440, Usage: "Size of message in bytes."},
				&cli.IntFlag{Name: "num-conn", Value: 2, Usage: "Number of random connections that a single node will make."},
				&cli.Float64Flag{Name: "cpu", Value: 0.05, Usage: "CPU limit per peer in terms cores (1.0 => 1 full core; 0.5 => half of a core)."},
				&cli.IntFlag{Name: "ram", Value: 16, Usage: "RAM limit per peer in MB."},
				&cli.IntFlag{Name: "downlink-bw", Value: 16, Usage: "Downlink DataRate limit per peer in MB."},
				&cli.IntFlag{Name: "uplink-bw", Value: 16, Usage: "Uplink DataRate limit per peer in MB."},
			},
			Action: func(c *cli.Context) error {

				CurrentConfig.NumberPeers = c.Int("num-peers")
				CurrentConfig.MessageRate = c.Int("msg-rate")
				CurrentConfig.MessageSize = c.Int("msg-size")
				CurrentConfig.ConnectTo = c.Int("num-conn")
				CurrentConfig.CpuPerInstance = float32(c.Float64("cpu"))
				CurrentConfig.RamPerInstance = c.Int("ram")
				CurrentConfig.DlBwMbps = c.Int("downlink-bw")
				CurrentConfig.UlBwMbps = c.Int("uplink-bw")
				CurrentConfig.ExperimentName = strings.ToLower(c.String("experiment-name"))
				return k3sutils.CreateNewExperiment(CurrentConfig)
			},
		},
		{
			Name:     "get",
			Aliases:  []string{"l", "list", "show"},
			Usage:    "list experiments",
			HideHelp: false,

			Flags: []cli.Flag{
				&cli.StringFlag{Name: "owner", Usage: "list all the experiment belonging to a given user.", Required: false, Aliases: []string{"user", "o"}},
			},
			Action: func(c *cli.Context) error {

				owner := c.String("owner")
				var err error
				if strings.EqualFold(owner, "") {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()

					err = k3sutils.ListAllExperiments(ctx)

				} else {
					tw := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
					fmt.Fprintln(tw, "feature not implemented yet (coming soon)")

					err = tw.Flush()
				}

				return err
			},
		},
		{
			Name:     "delete",
			Aliases:  []string{"d", "del", "remove"},
			Usage:    "delete experiments",
			HideHelp: false,

			Flags: []cli.Flag{
				&cli.StringFlag{Name: "experiment-name", Usage: "A user-friendly name for the experiment (e.g. blue) .", Required: true, Aliases: []string{"name", "n"}},
			},
			Action: func(c *cli.Context) error {
				expName := strings.ToLower(c.String("experiment-name"))
				return k3sutils.DeleteExperiment(expName)
			},
		},

		{
			Name:     "scale",
			Aliases:  []string{"update", "s"},
			Usage:    "scale number of peers in experiments",
			HideHelp: false,

			Flags: []cli.Flag{
				&cli.StringFlag{Name: "experiment-name", Usage: "A user-friendly name for the experiment (e.g. blue) .", Required: true, Aliases: []string{"name", "n"}},
				&cli.IntFlag{Name: "num-peers", Usage: "Number of peers in the network.", Aliases: []string{"peers", "p"}, Required: true},
			},
			Action: func(c *cli.Context) error {
				expName := strings.ToLower(c.String("experiment-name"))
				replicas := int32(c.Int("num-peers"))
				return k3sutils.ScaleExperiment(expName, replicas)
			},
		},
	},
}
