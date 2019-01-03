package peer

import (
	log "github.com/sirupsen/logrus"
	"net"
	"time"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/payload"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/util"
)

// Handshake either sends or receives a handshake.
// Sending involves writing a 'version' msg to an other peer.
// Receiving processes a received 'version' msg by sending our 'version' with a 'verack' msg.
func (p *Peer) Handshake() error {

	handshakeErr := make(chan error, 1)
	go func() {
		if p.inbound {
			// An other peer wants to handshake with us.
			handshakeErr <- p.inboundHandShake()
		} else {
			// We want to handshake with an other peer.
			handshakeErr <- p.outboundHandShake()
		}
	}()

	select {
	case err := <-handshakeErr:
		if err != nil {
			return err
		}
	case <-time.After(handshakeTimeout):
		return errHandShakeTimeout
	}

	// Log the handshake
	if p.inbound {
		log.WithField("prefix", "peer").Infof("Inbound handshake with %s successful", p.RemoteAddr().String())
	} else {
		log.WithField("prefix", "peer").Infof("Outbound handshake with %s successful", p.RemoteAddr().String())
	}
	return nil
}

// An other Peer wants to handshake with us.
// We will send our Version with a MsgVerAck.
func (p *Peer) inboundHandShake() error {
	var err error
	if err = p.writeLocalMsgVersion(); err != nil {
		return err
	}
	if err = p.readRemoteMsgVersion(); err != nil {
		return err
	}
	verack := payload.NewMsgVerAck()
	err = p.Write(verack)
	return p.readVerack()
}

func (p *Peer) outboundHandShake() error {
	var err error
	err = p.readRemoteMsgVersion()
	if err != nil {
		return err
	}

	err = p.writeLocalMsgVersion()
	if err != nil {
		return err
	}

	err = p.readVerack()
	if err != nil {
		return err
	}
	verack := payload.NewMsgVerAck()
	if err != nil {
		return err
	}
	return p.Write(verack)
}

func (p *Peer) writeLocalMsgVersion() error {
	//nonce := p.config.Nonce
	//relay := p.config.Relay
	fromPort := uint16(p.config.Port)
	//ua := p.config.UserAgent
	//sh := p.config.StartHeight()
	//services := p.config.Services
	version := p.config.ProtocolVer
	localIP, err := util.GetOutboundIP()
	if err != nil {
		return err
	}
	fromAddr := payload.NetAddress{localIP, fromPort}

	toIP := p.conn.RemoteAddr().(*net.TCPAddr).IP
	toPort := p.conn.RemoteAddr().(*net.TCPAddr).Port
	toAddr := payload.NewNetAddress(toIP.String(), uint16(toPort))

	messageVer := payload.NewMsgVersion(version, &fromAddr, toAddr)

	return p.Write(messageVer)
}

func (p *Peer) readRemoteMsgVersion() error {
	readmsg, err := wire.ReadMessage(p.conn, p.config.Net)
	if err != nil {
		return err
	}

	version, ok := readmsg.(*payload.MsgVersion)
	if !ok {
		return err
	}
	return p.OnVersion(version)
}

func (p *Peer) readVerack() error {
	readmsg, err := wire.ReadMessage(p.conn, p.config.Net)

	if err != nil {
		return err
	}

	_, ok := readmsg.(*payload.MsgVerAck)

	if !ok {
		return err
	}
	// should only be accessed on one go-routine
	p.verackReceived = true

	return nil
}
