package main

import "net"

type Peer struct {
	conn net.Conn
}

func (p *Peer) readLoop() {

}

func NewPeer(conn net.Conn, msgCh chan Message, delCh chan *Peer) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
		delCh: delCh,
	}
}
