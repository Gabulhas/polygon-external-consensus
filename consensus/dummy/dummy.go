package dummy

import (
	"github.com/Gabulhas/polygon-external-consensus/blockchain"
	"github.com/Gabulhas/polygon-external-consensus/consensus"
	"github.com/Gabulhas/polygon-external-consensus/helper/progress"
	"github.com/Gabulhas/polygon-external-consensus/state"
	"github.com/Gabulhas/polygon-external-consensus/txpool"
	"github.com/Gabulhas/polygon-external-consensus/types"
	"github.com/hashicorp/go-hclog"
)

type Dummy struct {
	logger     hclog.Logger
	notifyCh   chan struct{}
	closeCh    chan struct{}
	txpool     *txpool.TxPool
	blockchain *blockchain.Blockchain
	executor   *state.Executor
}

func Factory(params *consensus.Params) (consensus.Consensus, error) {
	logger := params.Logger.Named("dummy")

	d := &Dummy{
		logger:     logger,
		notifyCh:   make(chan struct{}),
		closeCh:    make(chan struct{}),
		blockchain: params.Blockchain,
		executor:   params.Executor,
		txpool:     params.TxPool,
	}

	return d, nil
}

// Initialize initializes the consensus
func (d *Dummy) Initialize() error {
	d.txpool.SetSealing(true)

	return nil
}

func (d *Dummy) Start() error {
	go d.run()

	return nil
}

func (d *Dummy) VerifyHeader(header *types.Header) error {
	// All blocks are valid
	return nil
}

func (d *Dummy) ProcessHeaders(headers []*types.Header) error {
	return nil
}

func (d *Dummy) GetBlockCreator(header *types.Header) (types.Address, error) {
	return types.BytesToAddress(header.Miner), nil
}

// PreCommitState a hook to be called before finalizing state transition on inserting block
func (d *Dummy) PreCommitState(_header *types.Header, _txn *state.Transition) error {
	return nil
}

func (d *Dummy) GetSyncProgression() *progress.Progression {
	return nil
}

func (d *Dummy) Close() error {
	close(d.closeCh)

	return nil
}

func (d *Dummy) run() {
	d.logger.Info("started")
	// do nothing
	<-d.closeCh
}
