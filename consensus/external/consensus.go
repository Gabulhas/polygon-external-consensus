package external

import (
	"time"

	"github.com/Gabulhas/polygon-external-consensus/blockchain"
	"github.com/Gabulhas/polygon-external-consensus/consensus"
	"github.com/Gabulhas/polygon-external-consensus/helper/progress"
	"github.com/Gabulhas/polygon-external-consensus/network"
	"github.com/Gabulhas/polygon-external-consensus/secrets"
	"github.com/Gabulhas/polygon-external-consensus/state"
	"github.com/Gabulhas/polygon-external-consensus/syncer"
	"github.com/Gabulhas/polygon-external-consensus/txpool"
	"github.com/Gabulhas/polygon-external-consensus/types"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
)

const (
	externalConsensus = "external-consensus"
	externalProto     = "/external/0.1"
)

type External struct {
	logger hclog.Logger // Reference to the logging

	notifyCh chan struct{}
	closeCh  chan struct{}

	Grpc           *grpc.Server
	blockchain     *blockchain.Blockchain
	config         *consensus.Config
	executor       *state.Executor // Executes the newer state (from a transaction) in the current state
	metrics        *consensus.Metrics
	network        *network.Server
	secretsManager secrets.SecretsManager
	syncer         syncer.Syncer
	txpool         *txpool.TxPool
	transport      transport
	blockTime      time.Duration
}

// Factory implements the base factory method
func Factory(
	params *consensus.Params,
) (consensus.Consensus, error) {
	logger := params.Logger.Named("external")

	d := &External{
		logger: logger,

		notifyCh: make(chan struct{}),
		closeCh:  make(chan struct{}),

		Grpc:           params.Grpc,
		blockchain:     params.Blockchain,
		config:         params.Config,
		executor:       params.Executor,
		metrics:        params.Metrics,
		secretsManager: params.SecretsManager,
		network:        params.Network,
		txpool:         params.TxPool,
		syncer: syncer.NewSyncer(
			params.Logger,
			params.Network,
			params.Blockchain,
			time.Duration(params.BlockTime)*3*time.Second,
		),
		blockTime: time.Duration(params.BlockTime) * time.Second,
	}

	return d, nil
}

// Initialize initializes the consensus
func (d *External) Initialize() error {
	// register the grpc operator
	//	if d.Grpc != nil {
	//		d.Grpc.RegisterService()
	//	}

	// start the transport protocol
	if err := d.setupTransport(); err != nil {
		return err
	}

	return nil
}

// Start starts the consensus mechanism
func (d *External) Start() error {
	go d.run()

	return nil
}

func (d *External) run() {
	d.logger.Info("consensus started")

	for {
		// wait until there is a new txn
		select {
		case <-d.closeCh:
			return
		}

		// There are new transactions in the pool, try to seal them
		header := d.blockchain.Header()
		if err := d.writeNewBlock(header); err != nil {
			d.logger.Error("failed to mine block", "err", err)
		}
	}
}

type transitionInterface interface {
	Write(txn *types.Transaction) error
}

func (d *External) writeTransactions(gasLimit uint64, transition transitionInterface) []*types.Transaction {
	var successful []*types.Transaction

	d.txpool.Prepare()

	for {
		tx := d.txpool.Peek()
		if tx == nil {
			break
		}

		if tx.ExceedsBlockGasLimit(gasLimit) {
			d.txpool.Drop(tx)

			continue
		}

		if err := transition.Write(tx); err != nil {
			if _, ok := err.(*state.GasLimitReachedTransitionApplicationError); ok { //nolint:errorlint
				break
			} else if appErr, ok := err.(*state.TransitionApplicationError); ok && appErr.IsRecoverable { //nolint:errorlint
				d.txpool.Demote(tx)
			} else {
				d.txpool.Drop(tx)
			}

			continue
		}

		// no errors, pop the tx from the pool
		d.txpool.Pop(tx)

		successful = append(successful, tx)
	}

	d.logger.Info("picked out txns from pool", "num", len(successful), "remaining", d.txpool.Length())

	return successful
}

// writeNewBLock generates a new block based on transactions from the pool,
// and writes them to the blockchain
func (d *External) writeNewBlock(parent *types.Header) error {
	// Generate the base block
	num := parent.Number
	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     num + 1,
		GasLimit:   parent.GasLimit, // Inherit from parent for now, will need to adjust dynamically later.
		Timestamp:  uint64(time.Now().Unix()),
	}

	// calculate gas limit based on parent header
	gasLimit, err := d.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return err
	}

	header.GasLimit = gasLimit

	miner, err := d.GetBlockCreator(header)
	if err != nil {
		return err
	}

	transition, err := d.executor.BeginTxn(parent.StateRoot, header, miner)

	if err != nil {
		return err
	}

	txns := d.writeTransactions(gasLimit, transition)

	// Commit the changes
	_, root := transition.Commit()

	// Update the header
	header.StateRoot = root
	header.GasUsed = transition.TotalGas()

	// Build the actual block
	// The header hash is computed inside buildBlock
	block := consensus.BuildBlock(consensus.BuildBlockParams{
		Header:   header,
		Txns:     txns,
		Receipts: transition.Receipts(),
	})

	if err := d.blockchain.VerifyFinalizedBlock(block); err != nil {
		return err
	}

	// Write the block to the blockchain
	if err := d.blockchain.WriteBlock(block, externalConsensus); err != nil {
		return err
	}

	// after the block has been written we reset the txpool so that
	// the old transactions are removed
	d.txpool.ResetWithHeaders(block.Header)

	return nil
}

// REQUIRED BASE INTERFACE METHODS //

func (d *External) VerifyHeader(header *types.Header) error {
	// All blocks are valid
	return nil
}

func (d *External) ProcessHeaders(headers []*types.Header) error {
	return nil
}

func (d *External) GetBlockCreator(header *types.Header) (types.Address, error) {
	return types.BytesToAddress(header.Miner), nil
}

// PreCommitState a hook to be called before finalizing state transition on inserting block
func (d *External) PreCommitState(_header *types.Header, _txn *state.Transition) error {
	return nil
}

func (d *External) GetSyncProgression() *progress.Progression {
	return nil
}

func (d *External) Close() error {
	close(d.closeCh)

	return nil
}
