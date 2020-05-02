package transactions

import (
	"bytes"

	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/encoding"
)

// Transaction according to the Phoenix model
type Transaction struct {
	Inputs  []*TransactionInput  `protobuf:"bytes,1,rep,name=inputs,proto3" json:"inputs,omitempty"`
	Outputs []*TransactionOutput `protobuf:"bytes,2,rep,name=outputs,proto3" json:"outputs,omitempty"`
	Fee     *TransactionOutput   `protobuf:"bytes,3,opt,name=fee,proto3" json:"fee,omitempty"`
	Proof   []byte               `protobuf:"bytes,4,opt,name=proof,proto3" json:"proof,omitempty"`
	hash    []byte
}

func (t *Transaction) setHash(h []byte) {
	t.hash = h
}

// CalculateHash complies with merkletree.Payload interface
func (t *Transaction) CalculateHash() ([]byte, error) {
	return t.hash, nil
}

// Type complies with ContractCall interface. Returns "Tx"
func (t *Transaction) Type() TxType {
	return Tx
}

// StandardTx returns the transaction itself. It complies with the ContractCall
// interface
func (t *Transaction) StandardTx() *Transaction {
	return t
}

//MarshalTransaction into a buffer
func MarshalTransaction(r *bytes.Buffer, t Transaction) error {
	if err := encoding.WriteVarInt(r, uint64(len(t.Inputs))); err != nil {
		return err
	}
	for _, input := range t.Inputs {
		if err := MarshalTransactionInput(r, *input); err != nil {
			return err
		}
	}

	if err := encoding.WriteVarInt(r, uint64(len(t.Outputs))); err != nil {
		return err
	}
	for _, output := range t.Outputs {
		if err := MarshalTransactionOutput(r, *output); err != nil {
			return err
		}
	}

	if err := MarshalTransactionOutput(r, *t.Fee); err != nil {
		return err
	}

	if err := encoding.WriteVarBytes(r, t.Proof); err != nil {
		return err
	}

	if err := encoding.WriteVarBytes(r, t.hash); err != nil {
		return err
	}
	return nil
}

// UnmarshalTransaction from a buffer
func UnmarshalTransaction(r *bytes.Buffer, t *Transaction) error {

	nIn, eerr := encoding.ReadVarInt(r)
	if eerr != nil {
		return eerr
	}

	t.Inputs = make([]*TransactionInput, nIn)
	for i := range t.Inputs {
		tIn := new(TransactionInput)
		if err := UnmarshalTransactionInput(r, tIn); err != nil {
			return err
		}
		t.Inputs[i] = tIn
	}

	nOut, err := encoding.ReadVarInt(r)
	if err != nil {
		return err
	}

	t.Outputs = make([]*TransactionOutput, nOut)
	for i := range t.Outputs {
		tOut := new(TransactionOutput)
		if err := UnmarshalTransactionOutput(r, tOut); err != nil {
			return err
		}
		t.Outputs[i] = tOut
	}

	t.Fee = new(TransactionOutput)
	if err := UnmarshalTransactionOutput(r, t.Fee); err != nil {
		return err
	}

	if err := encoding.ReadVarBytes(r, &t.Proof); err != nil {
		return err
	}

	if err := encoding.ReadVarBytes(r, &t.hash); err != nil {
		return err
	}
	return nil
}

// TransactionInput includes the notes, the nullifier and the transaction merkleroot
type TransactionInput struct {
	Note       *Note      `protobuf:"bytes,1,opt,name=note,proto3" json:"note"`
	Pos        uint64     `protobuf:"fixed64,2,opt,name=pos,proto3" json:"pos"`
	Sk         *SecretKey `protobuf:"bytes,3,opt,name=sk,proto3" json:"sk"`
	Nullifier  *Nullifier `protobuf:"bytes,4,opt,name=nullifier,proto3" json:"nullifier"`
	MerkleRoot *Scalar    `protobuf:"bytes,5,opt,name=merkle_root,json=merkleRoot,proto3" json:"merkle_root"`
}

// TransactionOutput is the spendable output of the transaction
type TransactionOutput struct {
	Note           *Note      `protobuf:"bytes,1,opt,name=note,proto3" json:"note"`
	Pk             *PublicKey `protobuf:"bytes,2,opt,name=pk,proto3" json:"pk"`
	Value          uint64     `protobuf:"fixed64,3,opt,name=value,proto3" json:"value"`
	BlindingFactor *Scalar    `protobuf:"bytes,4,opt,name=blinding_factor,json=blindingFactor,proto3" json:"blinding_factor"`
}

// MarshalTransactionInput to a buffer
func MarshalTransactionInput(r *bytes.Buffer, t TransactionInput) error {
	if err := MarshalNote(r, *t.Note); err != nil {
		return err
	}
	if err := encoding.WriteUint64LE(r, t.Pos); err != nil {
		return err
	}
	if err := MarshalSecretKey(r, *t.Sk); err != nil {
		return err
	}
	if err := MarshalNullifier(r, *t.Nullifier); err != nil {
		return err
	}
	if err := MarshalScalar(r, *t.MerkleRoot); err != nil {
		return err
	}
	return nil
}

// UnmarshalTransactionInput from a buffer
func UnmarshalTransactionInput(r *bytes.Buffer, t *TransactionInput) error {
	t.Note = new(Note)
	if err := UnmarshalNote(r, t.Note); err != nil {
		return err
	}
	if err := encoding.ReadUint64LE(r, &t.Pos); err != nil {
		return err
	}
	t.Sk = new(SecretKey)
	if err := UnmarshalSecretKey(r, t.Sk); err != nil {
		return err
	}
	t.Nullifier = new(Nullifier)
	if err := UnmarshalNullifier(r, t.Nullifier); err != nil {
		return err
	}
	t.MerkleRoot = new(Scalar)
	if err := UnmarshalScalar(r, t.MerkleRoot); err != nil {
		return err
	}
	return nil
}

// MarshalTransactionOutput into a buffer
func MarshalTransactionOutput(r *bytes.Buffer, t TransactionOutput) error {
	if err := MarshalNote(r, *t.Note); err != nil {
		return err
	}
	if err := MarshalPublicKey(r, *t.Pk); err != nil {
		return err
	}
	if err := encoding.WriteUint64LE(r, t.Value); err != nil {
		return err
	}
	if err := MarshalScalar(r, *t.BlindingFactor); err != nil {
		return err
	}
	return nil
}

// UnmarshalTransactionOutput from a buffer
func UnmarshalTransactionOutput(r *bytes.Buffer, t *TransactionOutput) error {
	t.Note = &Note{}
	if err := UnmarshalNote(r, t.Note); err != nil {
		return err
	}
	t.Pk = &PublicKey{}
	if err := UnmarshalPublicKey(r, t.Pk); err != nil {
		return err
	}
	if err := encoding.ReadUint64LE(r, &t.Value); err != nil {
		return err
	}
	t.BlindingFactor = &Scalar{}
	if err := UnmarshalScalar(r, t.BlindingFactor); err != nil {
		return err
	}
	return nil
}
