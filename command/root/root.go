package root

import (
	"fmt"
	"os"

	"github.com/Gabulhas/polygon-external-consensus/command/backup"
	"github.com/Gabulhas/polygon-external-consensus/command/genesis"
	"github.com/Gabulhas/polygon-external-consensus/command/helper"
	"github.com/Gabulhas/polygon-external-consensus/command/ibft"
	"github.com/Gabulhas/polygon-external-consensus/command/license"
	"github.com/Gabulhas/polygon-external-consensus/command/loadbot"
	"github.com/Gabulhas/polygon-external-consensus/command/monitor"
	"github.com/Gabulhas/polygon-external-consensus/command/peers"
	"github.com/Gabulhas/polygon-external-consensus/command/secrets"
	"github.com/Gabulhas/polygon-external-consensus/command/server"
	"github.com/Gabulhas/polygon-external-consensus/command/status"
	"github.com/Gabulhas/polygon-external-consensus/command/txpool"
	"github.com/Gabulhas/polygon-external-consensus/command/version"
	"github.com/Gabulhas/polygon-external-consensus/command/whitelist"
	"github.com/spf13/cobra"
)

type RootCommand struct {
	baseCmd *cobra.Command
}

func NewRootCommand() *RootCommand {
	rootCommand := &RootCommand{
		baseCmd: &cobra.Command{
			Short: "Polygon Edge is a framework for building Ethereum-compatible Blockchain networks",
		},
	}

	helper.RegisterJSONOutputFlag(rootCommand.baseCmd)

	rootCommand.registerSubCommands()

	return rootCommand
}

func (rc *RootCommand) registerSubCommands() {
	rc.baseCmd.AddCommand(
		version.GetCommand(),
		txpool.GetCommand(),
		status.GetCommand(),
		secrets.GetCommand(),
		peers.GetCommand(),
		monitor.GetCommand(),
		loadbot.GetCommand(),
		ibft.GetCommand(),
		backup.GetCommand(),
		genesis.GetCommand(),
		server.GetCommand(),
		whitelist.GetCommand(),
		license.GetCommand(),
	)
}

func (rc *RootCommand) Execute() {
	if err := rc.baseCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}
