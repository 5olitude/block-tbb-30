package node

import (
	"blocks/database"
	"blocks/fs"
	"context"
	"fmt"
	"math/rand"
	"time"
)

// nonce referr
// git push https://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@github.com/5olitude/block-tbb-30.git
type PendingBlock struct {
	parent database.Hash
	number uint64
	time   uint64
	txs    []database.Tx
}

func NewPendingBlock(parent database.Hash, number uint64, txs []database.Tx) PendingBlock {
	return PendingBlock{parent, number, uint64(time.Now().Unix()), txs}
}

func Mine(ctx context.Context, pb PendingBlock) (database.Block, error) {
	if len(pb.txs) == 0 {
		return database.Block{}, fmt.Errorf("mining empty block is not allowed")
	}
	start := time.Now()
	attempt := 0
	var block database.Block
	var hash database.Hash
	var nonce uint32

	for !database.IsBlockHashValid(hash) {
		select {
		case <-ctx.Done():
			fmt.Println("mining cancelled")
			return database.Block{}, fmt.Errorf("mining cancelled. %s", ctx.Err())
		default:

		}
		attempt++
		nonce = generatedNonce()
		if attempt%1000000 == 0 || attempt == 1 {
			fmt.Printf("Mining %d Pending Txs.Attempt:%d\n", len(pb.txs), attempt)
		}
		block = database.NewBlock(pb.parent, pb.number, nonce, pb.time, pb.txs)
		blockHash, err := block.Hash()
		if err != nil {
			return database.Block{}, fmt.Errorf("couldnt mine .block. %s", err.Error())
		}
		hash = blockHash
	}
	fmt.Printf("\nMined new Block '%x' using PoW🎉🎉🎉%s:\n", hash, fs.Unicode("\\U1F389"))
	fmt.Printf("\tHeight: '%v'\n", block.Header.Number)
	fmt.Printf("\tNonce: '%v'\n", block.Header.Nonce)
	fmt.Printf("\tCreated: '%v'\n", block.Header.Time)
	fmt.Printf("\tMiner: '%v'\n", block.Header.Time)
	fmt.Printf("\tParent: '%v'\n\n", block.Header.Parent.Hex())

	fmt.Printf("\tAttempt: '%v'\n", attempt)
	fmt.Printf("\tTime: %s\n\n", time.Since(start))

	return block, nil

}
func generatedNonce() uint32 {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Uint32()
}
