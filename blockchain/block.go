package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"time"
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction
	Prev_hash    []byte
	Nonce        int
}

func CreateBlock(transactions []*Transaction, prev_hash []byte) *Block {
	block := &Block{[]byte{}, transactions, prev_hash, 0}
	pow := NewProof(block)
	start := time.Now()
	nonce, hash := pow.Mine()
	elapsed := time.Since(start)
	fmt.Printf("Time Counter %s\n", elapsed)

	block.Hash = hash
	block.Nonce = nonce

	return block
}

func GenesisBlock(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, nil)
}

func (b *Block) Serialize() []byte {
	var ret bytes.Buffer
	encoder := gob.NewEncoder(&ret)

	err := encoder.Encode(b)
	if err != nil {
		panic(err)
	}

	return ret.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	if err := decoder.Decode(&block); err != nil {
		panic(err)
	}

	return &block
}
func (b *Block) Log() {
	fmt.Printf("Previous Hash: %x\n", b.Prev_hash)
	// fmt.Printf("Data in block: %s\n", b.Data)
	fmt.Printf("Hash: %x\n", b.Hash)
	fmt.Printf("Nonce: %x\n", b.Nonce)
	fmt.Println("Transactions: ")
	for _, tx := range b.Transactions {
		fmt.Println("Transaction id: ", hex.EncodeToString(tx.ID))
		for idx, input := range tx.Inputs {
			fmt.Println("Input ", idx)
			fmt.Println("- Input id: ", hex.EncodeToString(input.ID))
			fmt.Println("- Output index: ", input.Output_id)
			fmt.Println("- Sign: ", input.Sign)
		}
		fmt.Println()
		for idx, output := range tx.Outputs {
			fmt.Println("Output ", idx)
			fmt.Println("- Output value: ", output.Value)
			fmt.Println("- Pubkey: ", output.Pub_key)
		}
	}
	fmt.Println()
}
