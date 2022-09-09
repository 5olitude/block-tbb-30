package database

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type SnapShot [32]byte

type State struct {
	Balances  map[Account]uint
	txMempool []Tx
	dbFile    *os.File
	snapshot  SnapShot
}

func NewStateFromDisk() (*State, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	gen, err := loadGenesis(filepath.Join(cwd, "database", "genesis.json"))
	if err != nil {
		return nil, err
	}

	balances := make(map[Account]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	f, err := os.OpenFile(filepath.Join(cwd, "database", "tx.db"), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)

	state := &State{balances, make([]Tx, 0), f, SnapShot{}}

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var tx Tx
		err = json.Unmarshal(scanner.Bytes(), &tx)
		if err != nil {
			return nil, err
		}

		if err := state.apply(tx); err != nil {
			return nil, err
		}
	}
	err = state.doSnapShot()
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *State) LatestSnapShot() SnapShot {
	return s.snapshot
}

func (s *State) Add(tx Tx) error {
	if err := s.apply(tx); err != nil {
		return err
	}

	s.txMempool = append(s.txMempool, tx)

	return nil
}

func (s *State) Persist() (SnapShot, error) {
	mempool := make([]Tx, len(s.txMempool))
	copy(mempool, s.txMempool)

	for i := 0; i < len(mempool); i++ {
		txJson, err := json.Marshal(s.txMempool[i])
		if err != nil {
			return SnapShot{}, err
		}
		fmt.Printf("Persisting new TX to disk:\n")
		fmt.Printf("\t%s\n", txJson)
		if _, err = s.dbFile.Write(append(txJson, '\n')); err != nil {
			return SnapShot{}, err
		}
		err = s.doSnapShot()
		if err != nil {
			return SnapShot{}, err
		}
		fmt.Printf("NewDb Snapshot %x\n", s.snapshot)
		s.txMempool = append(s.txMempool[:i], s.txMempool[i+1:]...)

	}

	return s.snapshot, nil
}

func (s *State) Close() error {
	return s.dbFile.Close()
}

func (s *State) apply(tx Tx) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if s.Balances[tx.From]-tx.Value < 0 {
		return fmt.Errorf("insufficient balance")
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value

	return nil
}

func (s *State) doSnapShot() error {
	_, err := s.dbFile.Seek(0, 0)
	if err != nil {
		return err
	}
	txsData, err := ioutil.ReadAll(s.dbFile)
	fmt.Println(string(txsData))
	fmt.Println(err)
	if err != nil {
		return err
	}
	s.snapshot = sha256.Sum256(txsData)

	return nil
}
