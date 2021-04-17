package cli

import (
	"fmt"
	"os"
)

const (
	defaultCommands = `
kraicklist, a Search Ads Application

Commands:
 api  | run API server, i.e: go run main.go api
 seed | seed master data for first initiation, i.e: go run main.go seed
`
	CmdApi  = "api"
	CmdSeed = "seed"
)

func ParseCommand() (cmd string) {
	args := os.Args
	if len(args) == 1 {
		return
	}

	cmd = args[1]
	return
}

func PrintDefault() {
	fmt.Println(defaultCommands)
}
