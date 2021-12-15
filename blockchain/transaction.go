package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxOutput struct {
	Value   int
	Pub_key string
}

type TxInput struct {
	ID        []byte
	Output_id int
	Sign      string
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	HandleErr(err)

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput
	acc, valid_outputs := chain.FindSpentableOutputs(from, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for tx_id, outs := range valid_outputs {
		tx_id, err := hex.DecodeString(tx_id)
		HandleErr(err)

		for _, out := range outs {
			input := TxInput{tx_id, out, from}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TxOutput{amount, to})

	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

func CoinbaseTransaction(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txinput := TxInput{[]byte{}, -1, data}
	txoutput := TxOutput{100, to}

	tx := Transaction{nil, []TxInput{txinput}, []TxOutput{txoutput}}
	tx.SetID()

	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Output_id == -1
}

func (inp *TxInput) CanUnlock(address string) bool {
	return inp.Sign == address
}

func (out *TxOutput) CanBeUnlocked(address string) bool {
	// fmt.Println("Log: CanBeUnlocked function ", address)
	return out.Pub_key == address
}
