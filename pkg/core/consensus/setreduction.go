package consensus

import (
	"bytes"
	"encoding/hex"
	"time"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/crypto/hash"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/payload/consensusmsg"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/crypto"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/payload"
)

// SignatureSetReduction is the signature set reduction phase of the consensus.
func SignatureSetReduction(ctx *Context) error {
	for ; ctx.Step < MaxSteps; ctx.Step++ {
		// Generate our own signature set and propagate
		if err := signatureSetGeneration(ctx); err != nil {
			return err
		}

		// Collect signature set with highest score
		if err := signatureSetCollection(ctx); err != nil {
			return err
		}

		// Vote on collected signature set
		if err := committeeVoteSigSet(ctx); err != nil {
			return err
		}

		// Receive all other votes
		if err := countVotesSigSet(ctx); err != nil {
			return err
		}

		// If we timed out, go back to the beginning of the loop
		if ctx.SignatureSet == nil {
			continue
		}

		ctx.Step++
		if err := committeeVoteSigSet(ctx); err != nil {
			return err
		}

		if err := countVotesSigSet(ctx); err != nil {
			return err
		}

		// If we got a result, terminate
		if ctx.SignatureSet != nil {
			return nil
		}
	}

	return nil
}

func signatureSetCollection(ctx *Context) error {
	var voters [][]byte
	voters = append(voters, []byte(*ctx.Keys.EdPubKey))
	highest := ctx.weight
	timer := time.NewTimer(stepTime)

	for {
	out:
		select {
		case <-timer.C:
			return nil
		case m := <-ctx.SigSetCandidateChan:
			pl := m.Payload.(*consensusmsg.SigSetCandidate)

			// Check if this node's signature set is already recorded
			for _, voter := range voters {
				if bytes.Equal(voter, m.PubKey) {
					break out
				}
			}

			// Verify the message
			valid, stake, err := processMsg(ctx, m)
			if err != nil {
				return err
			}

			if !valid {
				break
			}

			// If the stake is higher than our current one, replace
			if stake > highest {
				highest = stake
				ctx.SignatureSet = pl.SignatureSet
			}
		}
	}
}

func signatureSetGeneration(ctx *Context) error {
	sigs, _ := crypto.RandEntropy(200) // Placeholder
	ctx.SignatureSet = sigs

	pl, err := consensusmsg.NewSigSetCandidate(ctx.BlockHash, ctx.SignatureSet,
		ctx.Keys.BLSPubKey.Marshal(), ctx.Score)
	if err != nil {
		return err
	}

	sigEd, err := createSignature(ctx, pl)
	if err != nil {
		return err
	}

	msg, err := payload.NewMsgConsensus(ctx.Version, ctx.Round, ctx.LastHeader.Hash,
		sigEd, []byte(*ctx.Keys.EdPubKey), pl)
	if err != nil {
		return err
	}

	// Gossip msg
	if err := ctx.SendMessage(ctx.Magic, msg); err != nil {
		return err
	}

	return nil
}

func committeeVoteSigSet(ctx *Context) error {
	// Sign signature set with BLS
	sigBLS, err := ctx.BLSSign(ctx.Keys.BLSSecretKey, ctx.Keys.BLSPubKey, ctx.SignatureSet)
	if err != nil {
		return err
	}

	// Hash signature set
	sigSetHash, err := hash.Sha3256(ctx.SignatureSet)
	if err != nil {
		return err
	}

	// Create signature set vote message to gossip
	pl, err := consensusmsg.NewSigSetVote(ctx.Step, ctx.BlockHash, sigSetHash, sigBLS,
		ctx.Keys.BLSPubKey.Marshal(), ctx.Score)
	if err != nil {
		return err
	}

	sigEd, err := createSignature(ctx, pl)
	msg, err := payload.NewMsgConsensus(ctx.Version, ctx.Round, ctx.LastHeader.Hash,
		sigEd, []byte(*ctx.Keys.EdSecretKey), pl)
	if err != nil {
		return err
	}

	// Gossip message
	if err := ctx.SendMessage(ctx.Magic, msg); err != nil {
		return err
	}

	return nil
}

func countVotesSigSet(ctx *Context) error {
	counts := make(map[string]uint64)
	var voters [][]byte
	voters = append(voters, []byte(*ctx.Keys.EdPubKey))
	counts[hex.EncodeToString(ctx.SignatureSet)] += ctx.weight
	timer := time.NewTimer(stepTime)

	for {
	out:
		select {
		case <-timer.C:
			ctx.SignatureSet = nil
			return nil
		case m := <-ctx.SigSetVoteChan:
			pl := m.Payload.(*consensusmsg.SigSetVote)

			// Check if this node's vote is already recorded
			for _, voter := range voters {
				if bytes.Equal(voter, m.PubKey) {
					break out
				}
			}

			// Verify the message
			valid, stake, err := processMsg(ctx, m)
			if err != nil {
				return err
			}

			if stake == 0 || !valid {
				break
			}

			voters = append(voters, m.PubKey)
			setStr := hex.EncodeToString(pl.SigSetHash)
			counts[setStr] += stake

			// If a set exceeds vote threshold, we will end the loop.
			if counts[setStr] > ctx.VoteLimit {
				timer.Stop()
				ctx.SignatureSet = pl.SigSetHash
				return nil
			}
		}
	}
}
