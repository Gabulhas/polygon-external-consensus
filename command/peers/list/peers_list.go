package list

import (
	"context"

	"github.com/Gabulhas/polygon-external-consensus/command"
	"github.com/Gabulhas/polygon-external-consensus/command/helper"
	"github.com/Gabulhas/polygon-external-consensus/server/proto"
	"github.com/spf13/cobra"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

func GetCommand() *cobra.Command {
	peersListCmd := &cobra.Command{
		Use:   "list",
		Short: "Returns the list of connected peers, including the current node",
		Run:   runCommand,
	}

	return peersListCmd
}

func runCommand(cmd *cobra.Command, _ []string) {
	outputter := command.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	peersList, err := getPeersList(helper.GetGRPCAddress(cmd))
	if err != nil {
		outputter.SetError(err)

		return
	}

	outputter.SetCommandResult(
		newPeersListResult(peersList.Peers),
	)
}

func getPeersList(grpcAddress string) (*proto.PeersListResponse, error) {
	client, err := helper.GetSystemClientConnection(grpcAddress)
	if err != nil {
		return nil, err
	}

	return client.PeersList(context.Background(), &empty.Empty{})
}
