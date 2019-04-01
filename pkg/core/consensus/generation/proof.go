package generation

import (
	ristretto "github.com/bwesterb/go-ristretto"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus/user"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus/zkproof"
)

// GenerateProof will generate the proof of blind bid, needed to successfully
// propose a block to the voting committee.
func (g *generator) generateProof(seed []byte) (zkproof.ZkProof, error) {
	// Turn values into scalars
	seedScalar := ristretto.Scalar{}
	seedScalar.Derive(seed)

	// Create a slice of scalars with a number of random bids (up to 10)
	bidListSubset := g.getBidListSubset()
	bidListScalars := convertBidListToScalars(bidListSubset)

	proof := zkproof.Prove(g.d, g.k, seedScalar, bidListScalars)

	return proof, nil
}

// bidsToScalars will take a global public list, take a subset from it, and then
// return it as a slice of scalars.
func (g *generator) getBidListSubset() user.BidList {
	numBids := g.getNumBids()
	return g.bidList.Subset(numBids)
}

// getNumBids will return how many bids to include in the bid list subset
// for the proof.
func (g *generator) getNumBids() int {
	numBids := len(g.bidList)
	if numBids > 10 {
		numBids = 10
	}

	return numBids
}

// convertBidListToScalars will take a BidList, and create a slice of scalars from it.
func convertBidListToScalars(bidList user.BidList) []ristretto.Scalar {
	scalarList := make([]ristretto.Scalar, len(bidList))
	for i, bid := range bidList {
		bidScalar := zkproof.BytesToScalar(bid[:])
		scalarList[i] = bidScalar
	}

	return scalarList
}
