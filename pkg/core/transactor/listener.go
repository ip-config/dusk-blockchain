package transactor

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/topics"
	"github.com/dusk-network/dusk-blockchain/pkg/wallet/transactions"
	logger "github.com/sirupsen/logrus"
)

var (
	log = logger.WithFields(logger.Fields{"prefix": "transactor"})

	errWalletNotLoaded     = errors.New("wallet is not loaded yet")
	errWalletAlreadyLoaded = errors.New("wallet is already loaded")
)

func (t *Transactor) Listen() {
	for {
		select {

		// Wallet Requests to respond to
		case r := <-createWalletChan:
			handleRequest(r, t.handleCreateWallet, "CreateWallet")

		case r := <-createFromSeedChan:
			handleRequest(r, t.handleCreateFromSeed, "CreateWalletFromSeed")

		case r := <-loadWalletChan:
			handleRequest(r, t.handleLoadWallet, "LoadWallet")

		// Transaction Requests to respond to

		case r := <-sendBidTxChan:
			handleRequest(r, t.handleSendBidTx, "BidTx")

		case r := <-sendStakeTxChan:
			handleRequest(r, t.handleSendStakeTx, "StakeTx")

		case r := <-sendStandardTxChan:
			handleRequest(r, t.handleSendStandardTx, "StandardTx")

		case r := <-getBalanceChan:
			handleRequest(r, t.handleBalance, "Balance")

			// Event list to handle
			//case blk := <-t.acceptedBlockChan:
			//	b.onAcceptedBlockEvent(blk)
		}
	}
}

func handleRequest(r wire.Req, handler func(r wire.Req) error, name string) {

	log.Infof("Handling %s request", name)

	if err := handler(r); err != nil {
		log.Errorf("Failed %s request: %v", name, err)
		r.ErrChan <- err
		return
	}

	log.Infof("Handled %s request", name)
}

func (t *Transactor) handleCreateWallet(r wire.Req) error {
	if t.w != nil {
		return errWalletAlreadyLoaded
	}

	if err := t.createWallet(r.Params.String()); err != nil {
		return err
	}

	r.RespChan <- bytes.Buffer{}
	return nil
}

func (t *Transactor) handleLoadWallet(r wire.Req) error {
	if t.w != nil {
		return errWalletAlreadyLoaded
	}

	if err := t.loadWallet(r.Params.String()); err != nil {
		return err
	}

	r.RespChan <- bytes.Buffer{}
	return nil
}

func (t *Transactor) handleCreateFromSeed(r wire.Req) error {
	if t.w != nil {
		return errWalletAlreadyLoaded
	}

	seed := r.Params.String()
	password := r.Params.String()

	if err := t.createFromSeed(seed, password); err != nil {
		return err
	}

	r.RespChan <- bytes.Buffer{}
	return nil
}

func (t *Transactor) handleSendBidTx(r wire.Req) error {
	if t.w == nil {
		return errWalletNotLoaded
	}

	// read tx parameters
	amount, err := readUint64Param(r)
	if err != nil {
		return err
	}

	lockTime, err := readUint64Param(r)
	if err != nil {
		return err
	}

	// create and sign transaction
	log.Tracef("Create a bid tx ( %d, %d)", amount, lockTime)

	tx, err := t.CreateBidTx(amount, lockTime)
	if err != nil {
		return err
	}

	//  Publish transaction to the mempool processing
	txid, err := t.publishTx(tx)
	if err != nil {
		return err
	}

	r.RespChan <- *bytes.NewBuffer(txid)
	return nil
}

func (t *Transactor) handleSendStakeTx(r wire.Req) error {

	if t.w == nil {
		return errWalletNotLoaded
	}

	// read tx parameters
	amount, err := readUint64Param(r)
	if err != nil {
		return err
	}

	lockTime, err := readUint64Param(r)
	if err != nil {
		return err
	}

	// create and sign transaction
	log.Tracef("Create a stake tx ( %d, %d)", amount, lockTime)

	tx, err := t.CreateStakeTx(amount, lockTime)
	if err != nil {
		return err
	}

	//  Publish transaction to the mempool processing
	txid, err := t.publishTx(tx)
	if err != nil {
		return err
	}

	r.RespChan <- *bytes.NewBuffer(txid)

	return nil
}

func (t *Transactor) handleSendStandardTx(r wire.Req) error {

	if t.w == nil {
		return errWalletNotLoaded
	}

	// read tx parameters
	amount, err := readUint64Param(r)
	if err != nil {
		return err
	}

	destPubKey := r.Params.String()

	// create and sign transaction
	log.Tracef("Create a standard tx ( %d, %s )", amount, destPubKey)

	tx, err := t.CreateStandardTx(amount, destPubKey)
	if err != nil {
		return err
	}

	//  Publish transaction to the mempool processing
	txid, err := t.publishTx(tx)
	if err != nil {
		return err
	}

	r.RespChan <- *bytes.NewBuffer(txid)

	return nil
}

func (t *Transactor) handleBalance(r wire.Req) error {

	if t.w == nil {
		return errWalletNotLoaded
	}

	balance, err := t.Balance()
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, balance); err != nil {
		return err
	}

	r.RespChan <- *buf
	return nil
}

func (t *Transactor) publishTx(tx transactions.Transaction) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := transactions.Marshal(buf, tx); err != nil {
		return nil, fmt.Errorf("error encoding transaction: %v\n", err)
	}

	hash, err := tx.CalculateHash()
	if err != nil {
		// If we found a valid bid tx, we should under no circumstance have issues marshalling it
		return nil, fmt.Errorf("error encoding transaction: %v\n", err)
	}

	t.eb.Publish(string(topics.Tx), buf)

	return hash, nil
}
