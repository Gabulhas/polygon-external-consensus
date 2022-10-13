package external

import (
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/proto"
)

const (
	DefaultEpochSize = 100000
	IbftKeyName      = "node.key"
	KeyEpochSize     = "epochSize"

	ibftProto = "/external/0.2"
)

// setupTransport sets up the gossip transport protocol
func (d *External) setupTransport() error {
	// Define a new topic
	topic, err := i.network.NewTopic(ibftProto, &proto.Message{})
	if err != nil {
		return err
	}

	// Subscribe to the newly created topic
	if err := topic.Subscribe(
		func(obj interface{}, _ peer.ID) {
			if !i.isActiveValidator() {
				return
			}

			msg, ok := obj.(*proto.Message)
			if !ok {
				i.logger.Error("invalid type assertion for message request")

				return
			}

			i.consensus.AddMessage(msg)

			i.logger.Debug(
				"validator message received",
				"type", msg.Type.String(),
				"height", msg.GetView().Height,
				"round", msg.GetView().Round,
				"addr", types.BytesToAddress(msg.From).String(),
			)
		},
	); err != nil {
		return err
	}

	i.transport = &gossipTransport{topic: topic}

	return nil
}
