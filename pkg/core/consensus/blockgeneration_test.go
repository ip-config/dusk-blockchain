package consensus

import (
	"encoding/hex"
	"testing"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/protocol"

	"github.com/stretchr/testify/assert"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/crypto"
)

func TestGenerateM(t *testing.T) {
	k := []byte{0, 1, 1}
	expect := "c66263a231f51157ce892c54031eed87822c534e9e6dd5471404ef2cabe4f2f7"

	hash, err := generateM(k)
	assert.Equal(t, nil, err)

	assert.Equal(t, expect, hex.EncodeToString(hash))
}
func TestGenerateY(t *testing.T) {

	d := uint64(20)
	S := []byte{0, 1, 1, 1, 1, 1}
	k := []byte{0, 1, 1}
	expect := "f0bcef720bf66bcaf0a8427212fc5bf9554bb8f5a6eb43ffc4cbb105b7958309"

	hash, err := generateY(d, S, k)
	assert.Equal(t, nil, err)

	assert.Equal(t, expect, hex.EncodeToString(hash))
}

func TestGenerateX(t *testing.T) {
	d := uint64(20)
	k := []byte{0, 1, 1}
	expected := "3eb8a8217cdd45475f91001355180e94ab8b182cb084ed839cd536673e99e41d"

	hash, err := generateX(d, k)

	assert.Equal(t, nil, err)
	assert.Equal(t, expected, hex.EncodeToString(hash))
}

//XXX: Add fixed test input vectors to outputs
func TestBlockGeneration(t *testing.T) {
	for i := 0; i < 1000; i++ {
		ctx, err := NewContext(200, 0, 150000, nil, protocol.TestNet, randtestKeys(t))
		assert.Equal(t, nil, err)

		k, err := crypto.RandEntropy(32)
		assert.Equal(t, err, nil)

		ctx.k = k
		err = GenerateBlock(ctx)
		assert.Equal(t, nil, err)
	}
}

// helper to generate random consensus keys
func randtestKeys(t *testing.T) *Keys {
	keys, err := NewRandKeys()
	if err != nil {
		t.FailNow()
	}
	return keys
}
