package whitelist

import (
	"github.com/Gabulhas/polygon-external-consensus/command/whitelist/deployment"
	"github.com/Gabulhas/polygon-external-consensus/command/whitelist/show"
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	whitelistCmd := &cobra.Command{
		Use:   "whitelist",
		Short: "Top level command for modifying the Polygon Edge whitelists within the config. Only accepts subcommands.",
	}

	registerSubcommands(whitelistCmd)

	return whitelistCmd
}

func registerSubcommands(baseCmd *cobra.Command) {
	baseCmd.AddCommand(
		deployment.GetCommand(),
		show.GetCommand(),
	)
}
