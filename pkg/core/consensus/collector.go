package consensus

import (
	"bytes"
	"encoding/binary"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus/msg"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus/user"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire"
)

// roundCollector is a simple wrapper over a channel to get round notifications.
// It is not supposed to be used directly. Components interestesd in Round updates
// should use InitRoundUpdate instead
type (
	roundCollector struct {
		roundChan chan uint64
	}

	regenerationCollector struct {
		regenerationChan chan AsyncState
	}

	bidListCollector struct {
		bidListChan chan<- user.BidList
	}
)

// UpdateRound is a shortcut for propagating a round
func UpdateRound(bus wire.EventPublisher, round uint64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, round)
	bus.Publish(msg.RoundUpdateTopic, bytes.NewBuffer(b))
}

// InitRoundUpdate initializes a Round update channel and fires up the TopicListener
// as well. Its purpose is to lighten up a bit the amount of arguments in creating
// the handler for the collectors. Also it removes the need to store subscribers on
// the consensus process
func InitRoundUpdate(subscriber wire.EventSubscriber) <-chan uint64 {
	roundChan := make(chan uint64, 1)
	roundCollector := &roundCollector{roundChan}
	go wire.NewTopicListener(subscriber, roundCollector, string(msg.RoundUpdateTopic)).Accept()
	return roundChan
}

// Collect as specified in the EventCollector interface. In this case Collect simply
// performs unmarshalling of the round event
func (r *roundCollector) Collect(roundBuffer *bytes.Buffer) error {
	round := binary.LittleEndian.Uint64(roundBuffer.Bytes())
	r.roundChan <- round
	return nil
}

// InitBlockRegenerationCollector initializes a regeneration channel, creates a
// regenerationCollector, and subscribes this collector to the BlockRegenerationTopic.
// The channel is then returned.
func InitBlockRegenerationCollector(subscriber wire.EventSubscriber) chan AsyncState {
	regenerationChan := make(chan AsyncState, 1)
	collector := &regenerationCollector{regenerationChan}
	go wire.NewTopicListener(subscriber, collector, msg.BlockRegenerationTopic).Accept()
	return regenerationChan
}

func (rg *regenerationCollector) Collect(r *bytes.Buffer) error {
	round := binary.LittleEndian.Uint64(r.Bytes()[:8])
	step := uint8(r.Bytes()[8])
	state := AsyncState{
		Round: round,
		Step:  step,
	}
	rg.regenerationChan <- state
	return nil
}

// InitBidListUpdate creates and initiates a channel for the updates in the BidList
func InitBidListUpdate(subscriber wire.EventSubscriber) chan user.BidList {
	bidListChan := make(chan user.BidList)
	collector := &bidListCollector{bidListChan}
	go wire.NewTopicListener(subscriber, collector, string(msg.BidListTopic)).Accept()
	return bidListChan
}

// Collect implements EventCollector.
// It reconstructs the bidList and sends it on its BidListChan
func (b *bidListCollector) Collect(r *bytes.Buffer) error {
	rCopy := *r
	bidList, err := user.ReconstructBidListSubset(rCopy.Bytes())
	if err != nil {
		return nil
	}
	b.bidListChan <- bidList
	return nil
}
