package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
)

const difficulty = 20

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func (pow *ProofOfWork) Mine() (int, []byte) {
	var int_hash big.Int
	var hash [32]byte

	nonce := 0
	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		int_hash.SetBytes(hash[:])
		if int_hash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}

	fmt.Println()
	return nonce, hash[:]
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty))
	pow := &ProofOfWork{b, target}
	return pow
}

func (b *Block) HashTransaction() []byte {
	var tx_hashes [][]byte
	var tx_hash [32]byte

	for _, tx := range b.Transactions {
		tx_hashes = append(tx_hashes, tx.ID)
	}
	tx_hash = sha256.Sum256(bytes.Join(tx_hashes, []byte{}))

	return tx_hash[:]
}

func (pow *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.Prev_hash,
			pow.Block.HashTransaction(),
			ToHex(int64(nonce)),
			ToHex(difficulty),
		},
		[]byte{},
	)
	return data
}

func ToHex(num int64) []byte {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.BigEndian, num)
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}
