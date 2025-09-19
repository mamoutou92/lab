package main

import (
	"os"

	"github.com/ipfs/go-log/v2"
	"github.com/mamoutou92/lab/args"
)

var logger = log.Logger("minilab")

func main() {
	log.SetAllLoggers(log.LevelInfo)
	log.SetLogLevel("nimp2p-lab", "info")
	if err := args.ExperimentParams.Run(os.Args); err != nil {
		os.Exit(1)
	}
	//logger.Infow("cmdline", "args", args.CurrentConfig)

}
