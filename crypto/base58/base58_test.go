package base58

// Note:testing takes around 15seconds.
import (
	"crypto/rand"
	"encoding/hex"
	"testing"
)

type testValues struct {
	dec []byte
	enc string
}

var n = 5000000
var testPairs = make([]testValues, 0, n)

func initTestPairs() {
	if len(testPairs) > 0 {
		return
	}
	// pre-make the test pairs, so it doesn't take up benchmark time...
	data := make([]byte, 32)
	for i := 0; i < n; i++ {
		rand.Read(data)
		testPairs = append(testPairs, testValues{dec: data, enc: Base58Encoding(data)})
	}
}

func randAlphabet() *Alphabet {
	// Permutes [0, 127] and returns the first 58 elements.
	// Like (math/rand).Perm but using crypto/rand.
	var randomness [128]byte
	rand.Read(randomness[:])

	var bts [128]byte
	for i, r := range randomness {
		j := int(r) % (i + 1)
		bts[i] = bts[j]
		bts[j] = byte(i)
	}
	return NewAlphabet(string(bts[:58]))
}

func TestEncodingAndDecoding(t *testing.T) {
	for k := 0; k < 10; k++ {
		testEncDecLoop(t, randAlphabet())
	}
	testEncDecLoop(t, BTCAlphabet)
}

func testEncDecLoop(t *testing.T, alph *Alphabet) {
	for j := 1; j < 256; j++ {
		var b = make([]byte, j)
		for i := 0; i < 100; i++ {
			rand.Read(b)
			fe := Base58EncodingAlphabet(b, alph)

			fd, ferr := Base58DecodingAlphabet(fe, alph)
			if ferr != nil {
				t.Errorf(" error: %v", ferr)
			}

			if hex.EncodeToString(b) != hex.EncodeToString(fd) {
				t.Errorf("decoding err: %s != %s", hex.EncodeToString(b), hex.EncodeToString(fd))
			}
		}
	}
}

func TestBase58WithBitcoinAddresses(t *testing.T) {

	testAddr := []string{
		"1QCaxc8hutpdZ62iKZsn1TCG3nh7uPZojq",
		"1DhRmSGnhPjUaVPAj48zgPV9e2oRhAQFUb",
		"17LN2oPYRYsXS9TdYdXCCDvF2FegshLDU2",
		"14h2bDLZSuvRFhUL45VjPHJcW667mmRAAn",
	}

	for ii, vv := range testAddr {
		// num := Base58Decode([]byte(vv))
		// chk := Base58Encode(num)
		num, err := Base58Decoding(vv)
		if err != nil {
			t.Errorf("Test %d, expected success, got error %s\n", ii, err)
		}
		chk := Base58Encoding(num)
		if vv != string(chk) {
			t.Errorf("Test %d, expected=%s got=%s Address did base58 encode/decode correctly.", ii, vv, chk)
		}
	}
}

func BenchmarkBase58Encoding(b *testing.B) {
	initTestPairs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Base58Encoding(testPairs[i].dec)
	}
}

func BenchmarkBase58Decoding(b *testing.B) {
	initTestPairs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Base58Decoding(testPairs[i].enc)
	}
}
