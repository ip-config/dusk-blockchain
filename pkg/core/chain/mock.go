package chain

import (
	"math/big"

	"github.com/bwesterb/go-ristretto"
	"github.com/dusk-network/dusk-blockchain/pkg/config"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/block"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/key"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/transactions"
)

// MockVerifier is a mock for the chain.Verifier interface
type MockVerifier struct {
}

// PerformSanityCheck on first N blocks and M last blocks
func (v *MockVerifier) PerformSanityCheck(uint64, uint64, uint64) error {
	return nil
}

// CheckBlock will verify whether a block is valid according to the rules of the consensus
func (v *MockVerifier) CheckBlock(prevBlock block.Block, blk block.Block) error {
	return nil
}

// MockLoader is the mock of the DB loader to help testing the chain
type MockLoader struct {
	blockchain []block.Block
}

// NewMockLoader creates a Mockup of the Loader interface
func NewMockLoader() Loader {
	mockchain := make([]block.Block, 0)
	return &MockLoader{mockchain}
}

// Height returns the height currently known by the Loader
func (m *MockLoader) Height() (uint64, error) {
	return uint64(len(m.blockchain)), nil
}

// LoadTip of the chain
func (m *MockLoader) LoadTip() (*block.Block, error) {
	return &m.blockchain[len(m.blockchain)], nil
}

// PerformSanityCheck on first N blocks and M last blocks
func (m *MockLoader) PerformSanityCheck(uint64, uint64, uint64) error {
	return nil
}

// Clear the mock
func (m *MockLoader) Clear() error {
	return nil
}

// Close the mock
func (m *MockLoader) Close(driver string) error {
	return nil
}

// Append the block to the internal blockchain representation
func (m *MockLoader) Append(blk *block.Block) error {
	m.blockchain = append(m.blockchain, *blk)
	return nil
}

// BlockAt the block to the internal blockchain representation
func (m *MockLoader) BlockAt(index uint64) (block.Block, error) {
	return m.blockchain[index], nil
}

// mocks an intermediate block with a coinbase attributed to a standard
// address. For use only when bootstrapping the network.
func mockFirstIntermediateBlock(prevBlockHeader *block.Header) (*block.Block, error) {
	blk := block.NewBlock()
	blk.Header.Seed = make([]byte, 33)
	blk.Header.Height = 1
	// Something above the genesis timestamp
	blk.Header.Timestamp = 1570000000
	blk.SetPrevBlock(prevBlockHeader)

	tx := mockDeterministicCoinbase()
	blk.AddTx(tx)
	root, err := blk.CalculateRoot()
	if err != nil {
		return nil, err
	}
	blk.Header.TxRoot = root

	hash, err := blk.CalculateHash()
	if err != nil {
		return nil, err
	}
	blk.Header.Hash = hash

	return blk, nil
}

func mockDeterministicCoinbase() transactions.Transaction {
	seed := make([]byte, 32)

	keyPair := key.NewKeyPair(seed)
	tx := transactions.NewCoinbase(make([]byte, 32), make([]byte, 32), 2)
	var r ristretto.Scalar
	r.SetZero()
	tx.SetTxPubKey(r)

	var reward ristretto.Scalar
	reward.SetBigInt(big.NewInt(int64(config.GeneratorReward)))

	_ = tx.AddReward(*keyPair.PublicKey(), reward)
	return tx
}
