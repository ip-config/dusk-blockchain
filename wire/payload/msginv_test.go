package payload

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toghrulmaharramov/dusk-go/crypto"
	"github.com/toghrulmaharramov/dusk-go/transactions"
)

func TestMsgInvEncodeDecode(t *testing.T) {
	byte32 := []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}

	// Input
	sig, _ := crypto.RandEntropy(2000)
	in := transactions.Input{
		KeyImage:  byte32,
		TxID:      byte32,
		Index:     1,
		Signature: sig,
	}

	// Output
	out := transactions.Output{
		Amount: 200,
		P:      byte32,
	}

	// Type attribute
	ta := transactions.TypeAttributes{
		Inputs:   []transactions.Input{in},
		TxPubKey: byte32,
		Outputs:  []transactions.Output{out},
	}

	R, _ := crypto.RandEntropy(32)
	s := transactions.Stealth{
		Version: 1,
		Type:    1,
		R:       R,
		TA:      ta,
	}

	msg := NewMsgInv()
	msg.AddTx(s)

	// TODO: test AddBlock function when block structure is decided
	buf := new(bytes.Buffer)
	if err := msg.Encode(buf); err != nil {
		t.Fatal(err)
	}

	msg2 := NewMsgInv()
	if err := msg2.Decode(buf); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, msg, msg2)
}