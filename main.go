package main

import (
	"log/slog"
	"net"
)

const DefaultAddr = ":5001"

type Config struct {
	ListenAddr string
}

type Message struct {
	peer *Peer
	cmd  Command
}
type Server struct {
	Config,
	ln net.Listener
	quitCh    chan struct{}
	addPeerCh chan *Peer
	deleteCh  chan struct{}
	msgCh     chan Message
}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = DefaultAddr
	}

	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		deleteCh:  make(chan *Peer),
		msgCh:     make(chan Message),
	}

}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)

	if err != nil {
		return err
	}
	s.ln = ln

	go s.loop()
	slog.Info("goredis server running", "listenAddr", s.ListenAddr)
	return s.acceptLoop()
}

func (s *Server) loop() {
	for {
		select {
		case msg := <-s.msgCh:
			if err := s.handleMessage(msg); err != nil {
				slog.Error("raw message eror", "err", err)
			}
		case <-s.quitCh:
			return
		case peer := <-s.addPeerCh:
			slog.Info("peer connected", "remoteAddr", peer.conn.RemoteAddr())
			s.peers[peer] = true
		case peer := <-s.delPeerCh:
			slog.Info("peer disconnected", "remoteAddr", peer.conn.RemoteAddr())
			delete(s.peers, peer)
		}
	}
}

func (s *Server) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			continue
		}

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn, s.msgCh, s.deleteCh)
	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error", "err", err, "remoteAddr", conn.RemoteAddr())
	}
}
