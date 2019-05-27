package reduction

import (
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus/committee"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus/reputation"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus/user"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire"
)

const committeeSize = 50

// Reducers defines a committee of reducers, and provides the ability to detect those
// who are not properly participating in this phase of the consensus.
type Reducers interface {
	committee.Committee
	reputation.Filter
}

type reductionCommittee struct {
	*committee.Extractor
}

func newReductionCommittee(eventBroker wire.EventBroker) *reductionCommittee {
	return &reductionCommittee{
		Extractor: committee.NewExtractor(eventBroker),
	}
}

// IsMember checks if the BLS key belongs to one of the Provisioners in the committee
func (r *reductionCommittee) IsMember(pubKeyBLS []byte, round uint64, step uint8) bool {
	votingCommittee := r.UpsertCommitteeCache(round, step, r.size())
	return votingCommittee.IsMember(pubKeyBLS)
}

// Quorum returns the amount of votes to reach a quorum
func (r *reductionCommittee) Quorum() int {
	return int(float64(r.size()) * 0.75)
}

func (r *reductionCommittee) size() int {
	provisioners := r.Provisioners()
	if provisioners.Size() > committeeSize {
		return committeeSize
	}
	return provisioners.Size()
}

func (r *reductionCommittee) FilterAbsentees(evs []wire.Event, round uint64, step uint8) user.VotingCommittee {
	votingCommittee := r.UpsertCommitteeCache(round, step, r.size())
	for _, ev := range evs {
		votingCommittee.Remove(ev.Sender())
	}
	return votingCommittee
}