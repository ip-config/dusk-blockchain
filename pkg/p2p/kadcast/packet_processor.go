package kadcast

import (
	"encoding/binary"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/dusk-network/dusk-blockchain/pkg/util/container/ring"
)

// ProcessTCPPacket receives a Packet and processes it according to
// it's type. It gets the packets from the circularqueue that
// connects the listeners with the packet processor.
func ProcessTCPPacket(queue *ring.Buffer, router *Router) {
	// Instantiate now the variables to not pollute
	// the stack.
	var err error
	var senderAddr *net.UDPAddr
	var tcpPayload []byte

	for {
		// Get all of the packets that are now on the queue.
		queuePackets, _ := queue.GetAll()
		for _, item := range queuePackets {
			// Get items from the queue packet taken.
			_, senderAddr, tcpPayload, err = decodeRedPacket(item)
			if err != nil {
				log.WithError(err).Warn("Error decoding the TCP packet taken from the ring.")
				continue
			}
			// Build packet struct
			var packet Packet
			unmarshalPacket(tcpPayload[:], &packet)

			// Extract headers info.
			tipus, senderID, nonce, peerRecepPort := packet.getHeadersInfo()

			// Verify IDNonce
			// If we get an error, we just skip the whole process since the
			// Peer was not validated.
			if err := verifyIDNonce(senderID, nonce); err != nil {
				log.WithError(err).Warn("Incorrect TCP packet sender ID. Skipping its processing.")
				continue
			}

			// Build Peer info and put the right port on it subsituting the one
			// used to send the message by the one where the peer wants to receive
			// the messages.
			ip := getPeerNetworkInfo(*senderAddr)
			port := binary.LittleEndian.Uint16(peerRecepPort[:])
			peerInf := MakePeer(ip, port)

			switch tipus {
			case 0:
				{
					if err := handleChunks(packet, router); err != nil {
						log.WithField("peer", router.MyPeerInfo.String()).Warnln(err)
					} else {
						log.WithField("from", peerInf.String()).
							WithField("peer", router.MyPeerInfo.String()).
							Trace("Received CHUNKS message")
					}
				}
			}
		}
	}
}

// ProcessUDPPacket receives a Packet and processes it according to
// it's type. It gets the packets from the circularqueue that
// connects the listeners with the packet processor.
func ProcessUDPPacket(queue *ring.Buffer, router *Router) {
	// Instantiate now the variables to not pollute
	// the stack.
	var err error
	var byteNum int
	var senderAddr *net.UDPAddr
	var udpPayload []byte

	for {
		// Get all of the packets that are now on the queue.
		queuePackets, _ := queue.GetAll()
		for _, item := range queuePackets {
			// Get items from the queue packet taken.
			byteNum, senderAddr, udpPayload, err = decodeRedPacket(item)
			if err != nil {
				log.WithError(err).Warn("Error decoding the UDP packet taken from the ring.")
				continue
			}

			// Build packet struct
			var packet Packet
			unmarshalPacket(udpPayload[:], &packet)

			// Extract headers info.
			tipus, senderID, nonce, peerRecepPort := packet.getHeadersInfo()

			// If we get an error, we just skip the whole process since the
			// Peer was not validated.
			if err := verifyIDNonce(senderID, nonce); err != nil {
				log.WithError(err).Warn("Incorrect UDP packet sender ID. Skipping its processing.")
				continue
			}

			// Build Peer info and put the right port on it subsituting the one
			// used to send the message by the one where the peer wants to receive
			// the messages.
			ip := getPeerNetworkInfo(*senderAddr)
			port := binary.LittleEndian.Uint16(peerRecepPort[:])
			peerInf := MakePeer(ip, port)

			// Check packet type and process it.
			switch tipus {
			case 0:
				log.WithField(
					"src", peerInf.String(),
				).Traceln("Received PING message")
				handlePing(peerInf, router)
			case 1:
				log.WithField(
					"src", peerInf.String(),
				).Traceln("Received PONG message")
				handlePong(peerInf, router)

			case 2:
				log.WithField(
					"src", peerInf.String(),
				).Traceln("Received FIND_NODES message")
				handleFindNodes(peerInf, router)

			case 3:
				log.WithField(
					"src", peerInf.String(),
				).Traceln("Received NODES message")
				handleNodes(peerInf, packet, router, byteNum)
			}
		}
	}
}
