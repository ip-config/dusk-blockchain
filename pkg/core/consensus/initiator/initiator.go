package initiator

import (
	"bytes"

	cfg "github.com/dusk-network/dusk-blockchain/pkg/config"
	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus/factory"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/block"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/transactions"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/wallet"
	"github.com/dusk-network/dusk-blockchain/pkg/core/database"
	"github.com/dusk-network/dusk-blockchain/pkg/core/database/heavy"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/peer/processing/chainsync"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/message"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/topics"
	"github.com/dusk-network/dusk-blockchain/pkg/util/nativeutils/eventbus"
	"github.com/dusk-network/dusk-blockchain/pkg/util/nativeutils/rpcbus"
	zkproof "github.com/dusk-network/dusk-zkproof"
	log "github.com/sirupsen/logrus"
)

// LaunchConsensus start the whole consensus algorithm
func LaunchConsensus(eventBroker *eventbus.EventBus, rpcBus *rpcbus.RPCBus, w *wallet.Wallet, _ *chainsync.Counter) {
	storeBidValues(eventBroker, rpcBus, w)
	if err := startProvisioner(eventBroker, rpcBus, w); err != nil {
		log.Panic(err)
	}
}

func startProvisioner(eventBroker *eventbus.EventBus, rpcBus *rpcbus.RPCBus, w *wallet.Wallet) error {
	// Setting up the consensus factory
	pubKey := w.PublicKey()
	f := factory.New(eventBroker, rpcBus, cfg.ConsensusTimeOut, &pubKey, w.Keys())
	f.StartConsensus()

	// If we are on genesis, we should kickstart the consensus
	resp, err := rpcBus.Call(topics.GetLastBlock, rpcbus.NewRequest(bytes.Buffer{}), 0)
	if err != nil {
		return err
	}
	blk := resp.(block.Block)

	if blk.Header.Height == 0 {
		msg := message.New(topics.Initialization, bytes.Buffer{})
		eventBroker.Publish(topics.Initialization, msg)
	}
	return nil
}

// storeBidValues finds the most recent bid belonging to the given
// wallet, and stores the relevant values needed by the consensus.
// This allows the components for block generation to properly function.
//nolint:unparam
func storeBidValues(eventBroker eventbus.Broker, rpcBus *rpcbus.RPCBus, w *wallet.Wallet) {
	k, err := w.ReconstructK()
	if err != nil {
		log.Panic(err)
	}

	m := zkproof.CalculateM(k)
	_, db := heavy.CreateDBConnection()
	for i := uint64(0); ; i++ {
		hash, err := getBlockHashForHeight(db, i)
		if err == database.ErrBlockNotFound {
			// We hit the end of the chain, so just exit here
			return
		} else if err != nil {
			log.Panic(err)
		}

		txs, err := getTxsForBlock(db, hash)
		if err != nil {
			log.Panic(err)
		}

		// Check if we should store any of these transactions
		for _, tx := range txs {
			bid, ok := tx.(*transactions.Bid)
			if !ok {
				continue
			}

			if bytes.Equal(bid.M, m.Bytes()) {
				err := db.Update(func(t database.Transaction) error {
					return t.StoreBidValues(bid.Outputs[0].Commitment.Bytes(), k.Bytes(), bid.Lock)
				})
				if err != nil {
					log.Panic(err)
				}
			}
		}
	}
}

func getBlockHashForHeight(db database.DB, height uint64) ([]byte, error) {
	var hash []byte
	err := db.View(func(t database.Transaction) error {
		var err error
		hash, err = t.FetchBlockHashByHeight(height)
		return err
	})
	return hash, err
}

func getTxsForBlock(db database.DB, hash []byte) ([]transactions.Transaction, error) {
	var txs []transactions.Transaction
	err := db.View(func(t database.Transaction) error {
		var err error
		txs, err = t.FetchBlockTxs(hash)
		return err
	})
	return txs, err
}
