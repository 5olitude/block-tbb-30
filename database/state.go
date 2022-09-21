package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
)

type SnapShot [32]byte

type State struct {
	Balances        map[Account]uint
	dbFile          *os.File
	latestBlock     Block
	latestBlockHash Hash
	hasGenesisBlock bool
}

func NewStateFromDisk(dataDir string) (*State, error) {
	err := initDataDirIfNotExists(dataDir)
	if err != nil {
		return nil, err
	}

	gen, err := loadGenesis(getGenesisJsonFilePath(dataDir))
	if err != nil {
		return nil, err
	}

	balances := make(map[Account]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	dbFilepath := getBlocksDbFilePath(dataDir)
	f, err := os.OpenFile(dbFilepath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)

	state := &State{balances, f, Block{}, Hash{}, false}

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		blockFsJson := scanner.Bytes()
		if len(blockFsJson) == 0 {
			break
		}
		var blockFs BlockFS
		err = json.Unmarshal(blockFsJson, &blockFs)
		if err != nil {
			return nil, err
		}

		err := applyBlock(blockFs.Value, state)
		if err != nil {
			return nil, err
		}
		state.latestBlock = blockFs.Value
		state.latestBlockHash = blockFs.Key
	}
	return state, nil
}

func (s *State) AddBlocks(blocks []Block) error {
	for _, b := range blocks {
		_, err := s.AddBlock(b)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *State) AddBlock(b Block) (Hash, error) {
	pendingState := s.copy()
	err := applyBlock(b, &pendingState)
	if err != nil {
		return Hash{}, err
	}
	blockHash, err := b.Hash()
	if err != nil {
		return Hash{}, err
	}
	BlockFs := BlockFS{blockHash, b}
	BlockFsJson, err := json.Marshal(BlockFs)
	if err != nil {
		return Hash{}, err
	}
	fmt.Printf("\npersisting new Block to disk:\n")
	fmt.Printf("\t%s\n", BlockFsJson)
	_, err = s.dbFile.Write(append(BlockFsJson, '\n'))
	if err != nil {
		return Hash{}, err
	}
	s.Balances = pendingState.Balances
	s.latestBlockHash = blockHash
	s.latestBlock = b
	return blockHash, nil
}
func (s *State) LatestBlock() Block {
	return s.latestBlock
}
func (s *State) LatestBlockHash() Hash {
	return s.latestBlockHash
}

func (s *State) Close() error {
	return s.dbFile.Close()
}
func (s *State) copy() State {
	c := State{}
	c.latestBlock = s.latestBlock
	c.latestBlockHash = s.LatestBlockHash()
	c.Balances = make(map[Account]uint)
	for acc, balance := range s.Balances {
		c.Balances[acc] = balance
	}
	return c
}
func applyBlock(b Block, s *State) error {
	nextExpectedBlockNumber := s.latestBlock.Header.Number + 1

	if s.hasGenesisBlock && b.Header.Number != nextExpectedBlockNumber {
		return fmt.Errorf("next expected block must '%d' not '%d'", nextExpectedBlockNumber, b.Header.Number)
	}

	if s.hasGenesisBlock && s.latestBlock.Header.Number > 0 && !reflect.DeepEqual(b.Header.Parent, s.latestBlockHash) {
		return fmt.Errorf("next block parent hash must be '%x' not '%x'", s.latestBlockHash, b.Header.Parent)
	}
	hash, err := b.Hash()
	if err != nil {
		return err
	}
	if !IsBlockHashValid(hash) {
		return fmt.Errorf("Invalid block hash  %x", hash)
	}

	err = applyTXs(b.Txs, s)
	if err != nil {
		return err
	}
	s.Balances[b.Header.Miner] += BlockReward
	return nil
}
func applyTXs(txs []Tx, s *State) error {
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Time < txs[j].Time
	})
	for _, tx := range txs {
		err := applyTx(tx, s)
		if err != nil {
			return err
		}
	}

	return nil
}
func applyTx(tx Tx, s *State) error {
	if tx.Value > s.Balances[tx.From] {
		return fmt.Errorf("wrong TX. Sender '%s' balance is %d TBB. Tx cost is %d TBB", tx.From, s.Balances[tx.From], tx.Value)
	}
	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value
	return nil
}
