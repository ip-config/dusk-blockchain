package transactor

import (
	ristretto "github.com/bwesterb/go-ristretto"
	"github.com/dusk-network/dusk-blockchain/pkg/core/block"
	"github.com/dusk-network/dusk-blockchain/pkg/core/database"
	"github.com/dusk-network/dusk-blockchain/pkg/core/database/heavy"

	walletdb "github.com/dusk-network/dusk-blockchain/pkg/wallet/database"
	"github.com/dusk-network/dusk-blockchain/pkg/wallet/transactions"
	"github.com/dusk-network/dusk-crypto/mlsag"
	"github.com/dusk-network/dusk-wallet/key"

	"encoding/binary"

	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire"
)

func fetchBlockHeightAndState(db database.DB, height uint64) (*block.Block, []byte, error) {
	var blk *block.Block
	var state *database.State
	err := db.View(func(t database.Transaction) error {
		hash, err := t.FetchBlockHashByHeight(height)
		if err != nil {
			return err
		}
		state, err = t.FetchState()
		if err != nil {
			return err
		}

		blk, err = t.FetchBlock(hash)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return blk, state.TipHash, nil
}

func fetchDecoys(numMixins int) []mlsag.PubKeys {
	_, db := heavy.CreateDBConnection()

	var pubKeys []mlsag.PubKeys
	var decoys []ristretto.Point
	db.View(func(t database.Transaction) error {
		decoys = t.FetchDecoys(numMixins)
		return nil
	})

	// Potential panic if the database does not have enough decoys
	for i := 0; i < numMixins; i++ {

		var keyVector mlsag.PubKeys
		keyVector.AddPubKey(decoys[i])

		var secondaryKey ristretto.Point
		secondaryKey.Rand()
		keyVector.AddPubKey(secondaryKey)

		pubKeys = append(pubKeys, keyVector)
	}
	return pubKeys
}

func fetchInputs(netPrefix byte, db *walletdb.DB, totalAmount int64, key *key.Key) ([]*transactions.Input, int64, error) {
	// Fetch all inputs from database that are >= totalAmount
	// returns error if inputs do not add up to total amount
	privSpend, err := key.PrivateSpend()
	if err != nil {
		return nil, 0, err
	}
	return db.FetchInputs(privSpend.Bytes(), totalAmount)
}

// readUint64Param reads uint64 value from request params buffer
// TODO: Move this to rpcbus package as global util
func readUint64Param(r wire.Req) (uint64, error) {

	valueBytes := make([]byte, 8)
	if _, err := r.Params.Read(valueBytes); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(valueBytes), nil
}
