package main

import (
	"kraicklist/infra/api"
	"kraicklist/infra/cli"
	"kraicklist/infra/seed"
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
