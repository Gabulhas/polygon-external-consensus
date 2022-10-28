package external

import (
	"fmt"

	"github.com/Gabulhas/polygon-external-consensus/consensus/external/proto"
	"github.com/Gabulhas/polygon-external-consensus/network"
	"github.com/Gabulhas/polygon-external-consensus/types"
	"github.com/libp2p/go-libp2p/core/peer"
)

type transport interface {
	Multicast(msg *proto.Message) error
}

type gossipTransport struct {
	topic *network.Topic
}

func (g *gossipTransport) Multicast(msg *proto.Message) error {
	return g.topic.Publish(msg)
}

func (d *External) Multicast(msg *proto.Message) {
	if err := d.transport.Multicast(msg); err != nil {
		d.logger.Error("fail to gossip", "err", err)
	}
}

// setupTransport sets up the gossip transport protocol
func (d *External) setupTransport() error {
	// Define a new topic
	topic, err := d.network.NewTopic(externalProto, &proto.Message{})
	if err != nil {
		return err
	}

	// Subscribe to the newly created topic
	if err := topic.Subscribe(
		func(obj interface{}, _ peer.ID) {

			msg, ok := obj.(*proto.Message)
			fmt.Printf("msg: %v\n", msg)
			if !ok {
				d.logger.Error("invalid type assertion for message request")

				return
			}

			d.logger.Debug(
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

	d.transport = &gossipTransport{topic: topic}

	return nil
}
