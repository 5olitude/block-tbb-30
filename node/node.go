package node

import (
	"blocks/database"
	"context"
	"fmt"
	"net/http"
)

const DefaultIP = "127.0.0.1"
const DefaultHTTPort = 8080
const endpointStatus = "/node/status"
const endpointSync = "node/sync"
const endpointSyncQueryKeyFromBlock = "fromBlock"
const endpointAddPeer = "/node/peer"
const endpointAddPeerQueryKeyIP = "ip"
const endpointAddPeerQueryKeyPort = "port"

type PeerNode struct {
	IP          string `json:"ip"`
	Port        uint64 `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`
	connected   bool
}

func (pn PeerNode) TcpAddress() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
}

type Node struct {
	dataDir    string
	ip         string
	port       uint64
	state      *database.State
	KnownPeers map[string]PeerNode
}

func New(dataDir string, ip string, port uint64, bootstrap PeerNode) *Node {
	knownPeers := make(map[string]PeerNode)
	knownPeers[bootstrap.TcpAddress()] = bootstrap
	return &Node{
		dataDir:    dataDir,
		port:       port,
		ip:         ip,
		KnownPeers: knownPeers,
	}
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, connected bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, connected}
}
func (n *Node) Run() error {
	ctx := context.Background()
	fmt.Println(fmt.Sprintf("Listening on: %s:%d", n.ip, n.port))
	state, err := database.NewStateFromDisk(n.dataDir)
	if err != nil {
		return err
	}
	defer state.Close()
	n.state = state
	go n.sync(ctx)
	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		listBalanceHandler(w, r, state)
	})
	http.HandleFunc("/tx/add", func(w http.ResponseWriter, r *http.Request) {
		txAddHandler(w, r, state)
	})

	http.HandleFunc("endpointStatus", func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, n)
	})
	http.HandleFunc(endpointSync, func(w http.ResponseWriter, r *http.Request) {
		syncHandler(w, r, n)
	})
	http.HandleFunc(endpointAddPeer, func(w http.ResponseWriter, r *http.Request) {
		addPeerHandler(w, r, n)
	})
	return http.ListenAndServe(fmt.Sprintf("%s:%d", n.ip, n.port), nil)
}
func (n *Node) AddPeer(peer PeerNode) {
	n.KnownPeers[peer.TcpAddress()] = peer
}

func (n *Node) RemovePeer(peer PeerNode) {
	delete(n.KnownPeers, peer.TcpAddress())
}

func (n *Node) IsKnownPeer(peer PeerNode) bool {
	if peer.IP == n.ip && peer.Port == n.port {
		return true
	}

	_, isKnownPeer := n.KnownPeers[peer.TcpAddress()]

	return isKnownPeer
}
