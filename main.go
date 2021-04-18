package main

import (
	"github.com/isdzulqor/kraicklist/infra/api"
	"github.com/isdzulqor/kraicklist/infra/cli"
	"github.com/isdzulqor/kraicklist/infra/seed"
)

func main() {
	cmd := cli.ParseCommand()

	switch cmd {
	case cli.CmdApi:
		api.Exec()
	case cli.CmdSeed:
		seed.Exec()
	default:
		cli.PrintDefault()
	}
}
