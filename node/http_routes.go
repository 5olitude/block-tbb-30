package node

import (
	"blocks/database"
	"net/http"
)

type ErrRes struct {
	Error string `json:"error"`
}

type BalanceRes struct {
	Hash     database.Hash             `json:"block_hash"`
	Balances map[database.Account]uint `json:"balances"`
}

type TxAddReq struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

type TxAddRes struct {
	Hash database.Hash `json:"block_hash"`
}

type StatusRes struct {
	Hash       database.Hash       `json:"block_hash"`
	Number     uint64              `json:"block_number"`
	KnownPeers map[string]PeerNode `json:"peers_known"`
}

func listBalanceHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	writeRes(w, BalanceRes{state.LatestBlockHash(), state.Balances})
}

func txAddHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	req := TxAddReq{}
	err := readReq(r, &req)
	if err != nil {
		writeRes(w, err)
		return
	}
	tx := database.NewTx(database.NewAccount(req.From), database.NewAccount(req.To), req.Value, req.Data)
	err = state.AddTx(tx)
	if err != nil {
		writeRes(w, err)
		return
	}
	hash, err := state.Persist()
	if err != nil {
		writeRes(w, err)
		return
	}
	writeRes(w, TxAddRes{hash})
}

func statusHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	res := StatusRes{
		Hash:       node.state.LatestBlockHash(),
		Number:     node.state.LatestBlock().Header.Number,
		KnownPeers: node.KnownPeers,
	}
	writeRes(w, res)
}
