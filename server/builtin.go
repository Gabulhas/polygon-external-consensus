package server

import (
	"github.com/Gabulhas/polygon-external-consensus/consensus"
	consensusDev "github.com/Gabulhas/polygon-external-consensus/consensus/dev"
	consensusDummy "github.com/Gabulhas/polygon-external-consensus/consensus/dummy"
	consensusExternal "github.com/Gabulhas/polygon-external-consensus/consensus/external"
	consensusIBFT "github.com/Gabulhas/polygon-external-consensus/consensus/ibft"
	"github.com/Gabulhas/polygon-external-consensus/secrets"
	"github.com/Gabulhas/polygon-external-consensus/secrets/awsssm"
	"github.com/Gabulhas/polygon-external-consensus/secrets/gcpssm"
	"github.com/Gabulhas/polygon-external-consensus/secrets/hashicorpvault"
	"github.com/Gabulhas/polygon-external-consensus/secrets/local"
)

type ConsensusType string

const (
	DevConsensus      ConsensusType = "dev"
	IBFTConsensus     ConsensusType = "ibft"
	DummyConsensus    ConsensusType = "dummy"
	ExternalConsensus ConsensusType = "external"
)

var consensusBackends = map[ConsensusType]consensus.Factory{
	DevConsensus:      consensusDev.Factory,
	IBFTConsensus:     consensusIBFT.Factory,
	DummyConsensus:    consensusDummy.Factory,
	ExternalConsensus: consensusExternal.Factory,
}

// secretsManagerBackends defines the SecretManager factories for different
// secret management solutions
var secretsManagerBackends = map[secrets.SecretsManagerType]secrets.SecretsManagerFactory{
	secrets.Local:          local.SecretsManagerFactory,
	secrets.HashicorpVault: hashicorpvault.SecretsManagerFactory,
	secrets.AWSSSM:         awsssm.SecretsManagerFactory,
	secrets.GCPSSM:         gcpssm.SecretsManagerFactory,
}

func ConsensusSupported(value string) bool {
	_, ok := consensusBackends[ConsensusType(value)]

	return ok
}
