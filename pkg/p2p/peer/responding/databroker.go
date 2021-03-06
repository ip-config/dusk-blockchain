package responding

import (
	"bytes"

	"github.com/dusk-network/dusk-blockchain/pkg/config"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/block"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/transactions"
	"github.com/dusk-network/dusk-blockchain/pkg/core/database"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/peer/peermsg"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/message"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/topics"
	"github.com/dusk-network/dusk-blockchain/pkg/util/nativeutils/rpcbus"
)

// DataBroker is a processing unit responsible for handling GetData messages. It
// maintains a connection to the outgoing message queue of the peer it receives this
// message from.
type DataBroker struct {
	db           database.DB
	responseChan chan<- *bytes.Buffer
	rpcBus       *rpcbus.RPCBus
}

// NewDataBroker returns an initialized DataBroker.
func NewDataBroker(db database.DB, rpcBus *rpcbus.RPCBus, responseChan chan<- *bytes.Buffer) *DataBroker {
	return &DataBroker{
		db:           db,
		responseChan: responseChan,
		rpcBus:       rpcBus,
	}
}

// SendItems takes a GetData message from the wire, and iterates through the list,
// sending back each item's complete data to the requesting peer.
func (d *DataBroker) SendItems(m *bytes.Buffer) error {
	msg := &peermsg.Inv{}
	if err := msg.Decode(m); err != nil {
		return err
	}

	for _, obj := range msg.InvList {

		var buf *bytes.Buffer
		switch obj.Type {
		case peermsg.InvTypeBlock:
			// Fetch block from local state. It must be available
			var b *block.Block
			err := d.db.View(func(t database.Transaction) error {
				var err error
				b, err = t.FetchBlock(obj.Hash)
				return err
			})

			if err != nil {
				return err
			}

			// Send the block data back to the initiator node as topics.Block msg
			if buf, err = marshalBlock(b); err != nil {
				return err
			}

		case peermsg.InvTypeMempoolTx:
			// Try to retrieve tx from local mempool state. It might not be
			// available
			txs, err := GetMempoolTxs(d.rpcBus, obj.Hash)
			if err != nil {
				return err
			}

			if len(txs) != 0 {
				// Send topics.Tx with the tx data back to the initiator
				buf, err = marshalTx(txs[0])
				if err != nil {
					return err
				}
			}

			// A txID will not be found in a few situations:
			//
			// - The node has restarted and lost this Tx
			// - The node has recently accepted a block that includes this Tx
			// No action to run in these cases.
		}

		if buf != nil {
			d.responseChan <- buf
		}
	}

	return nil
}

// SendTxsItems will run tx items
func (d *DataBroker) SendTxsItems() error {

	var maxItemsSent = config.Get().Mempool.MaxInvItems
	if maxItemsSent == 0 {
		// responding to wire.Mempool disabled
		return nil
	}

	// TODO: Limit the returned txs slice size based on MaxInvTxs
	txs, err := GetMempoolTxs(d.rpcBus, nil)
	if err != nil {
		return err
	}

	if len(txs) != 0 {
		// Send topics.Inv with the tx data back to the initiator
		msg := &peermsg.Inv{}

		for _, tx := range txs {
			txID, err := tx.CalculateHash()
			if err != nil {
				return err
			}

			msg.AddItem(peermsg.InvTypeMempoolTx, txID)

			maxItemsSent--
			if maxItemsSent == 0 {
				break
			}
		}

		buf, err := marshalInv(msg)
		if err != nil {
			return err
		}

		d.responseChan <- buf
	}

	return nil
}

func marshalBlock(b *block.Block) (*bytes.Buffer, error) {
	//TODO: following is more efficient, saves an allocation and avoids the explicit Prepend
	// buf := topics.Topics[topics.Block].Buffer
	buf := new(bytes.Buffer)
	if err := message.MarshalBlock(buf, b); err != nil {
		return nil, err
	}

	if err := topics.Prepend(buf, topics.Block); err != nil {
		return nil, err
	}

	return buf, nil
}

func marshalTx(tx transactions.Transaction) (*bytes.Buffer, error) {
	//TODO: following is more efficient, saves an allocation and avoids the explicit Prepend
	// buf := topics.Topics[topics.Block].Buffer
	buf := new(bytes.Buffer)
	if err := message.MarshalTx(buf, tx); err != nil {
		return nil, err
	}

	if err := topics.Prepend(buf, topics.Tx); err != nil {
		return nil, err
	}

	return buf, nil
}
