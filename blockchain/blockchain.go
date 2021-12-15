package blockchain

import (
	"encoding/hex"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

var (
	DB_PATH      = "./temp/badger/"
	DB_FILE      = DB_PATH + "MANIFEST"
	GENESIS_DATA = "FIRST TRANSACTION FROM GENESIS"
)

type BlockChain struct {
	Last_hash []byte
	Db        *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Db          *badger.DB
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	it := &BlockChainIterator{chain.Last_hash, chain.Db}
	return it
}

func (chain *BlockChain) Log() {
	fmt.Printf("\nLog blockchain\n")
	it := chain.Iterator()
	cnt := 0
	for it.CurrentHash != nil {
		block := it.Prev()
		block.Log()
		cnt++
	}
	fmt.Println("Number of block: ", cnt)
}

func (it *BlockChainIterator) Prev() *Block {
	var block *Block
	err := it.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(it.CurrentHash)
		HandleErr(err)
		err = item.Value(func(val []byte) error {
			block = Deserialize(val)
			return nil
		})
		it.CurrentHash = block.Prev_hash
		return err
	})
	HandleErr(err)

	return block
}

func DbExists() bool {
	if _, err := os.Stat(DB_FILE); os.IsNotExist(err) {
		return false
	}
	return true
}

func InitBlockChain(address string) *BlockChain {
	var last_hash []byte

	if DbExists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	db, err := badger.Open(badger.DefaultOptions(DB_PATH))
	HandleErr(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTransaction(address, GENESIS_DATA)
		genesis := GenesisBlock(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		HandleErr(err)
		err = txn.Set([]byte("last_hash"), genesis.Hash)
		last_hash = genesis.Hash
		return err
	})

	blockchain := BlockChain{last_hash, db}
	return &blockchain
}

func ContinueBlockChain(address string) *BlockChain {
	if !DbExists() {
		fmt.Println("No existing blockchain found!")
		runtime.Goexit()
	}

	var last_hash []byte
	db, err := badger.Open(badger.DefaultOptions(DB_PATH))
	HandleErr(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("last_hash"))
		HandleErr(err)
		err = item.Value(func(val []byte) error {
			last_hash = val
			return nil
		})
		return err
	})

	chain := &BlockChain{last_hash, db}
	return chain
}

func (chain *BlockChain) AddBlock(transactions []*Transaction) {
	var last_hash []byte

	err := chain.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("last_hash"))
		HandleErr(err)
		err = item.Value(func(val []byte) error {
			last_hash = val
			return nil
		})
		return err
	})
	HandleErr(err)

	new_block := CreateBlock(transactions, last_hash)

	err = chain.Db.Update(func(txn *badger.Txn) error {
		err = txn.Set(new_block.Hash, new_block.Serialize())
		HandleErr(err)
		txn.Set([]byte("last_hash"), new_block.Hash)

		chain.Last_hash = new_block.Hash
		return err
	})

	new_block.Log()
}

func (chain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspent_txs []Transaction

	spent_txos := make(map[string][]int)

	it := chain.Iterator()
	for {
		block := it.Prev()

		for _, tx := range block.Transactions {
			tx_id := hex.EncodeToString(tx.ID)

		output_loop:
			for out_id, out := range tx.Outputs {
				if spent_txos[tx_id] != nil {
					for _, spent_out := range spent_txos[tx_id] {
						if spent_out == out_id {
							continue output_loop
						}
					}
				}
				if out.CanBeUnlocked(address) {
					unspent_txs = append(unspent_txs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) {
						in_tx_id := hex.EncodeToString(in.ID)
						spent_txos[in_tx_id] = append(spent_txos[in_tx_id], in.Output_id)
					}
				}
			}
		}

		if block.Prev_hash == nil {
			break
		}
	}

	return unspent_txs
}

func (chain *BlockChain) FindUTXO(address string) []TxOutput {
	var ret []TxOutput
	unspent_txs := chain.FindUnspentTransactions(address)

	for _, tx := range unspent_txs {
		for _, out := range tx.Outputs {
			if out.CanBeUnlocked(address) {
				ret = append(ret, out)
			}
		}
	}

	return ret
}

func (chain *BlockChain) FindSpentableOutputs(address string, amount int) (int, map[string][]int) {
	unspent_outs := make(map[string][]int)
	unspent_txs := chain.FindUnspentTransactions(address)
	sum := 0

main_loop:
	for _, tx := range unspent_txs {
		tx_id := hex.EncodeToString(tx.ID)
		for out_idx, out := range tx.Outputs {
			if out.CanBeUnlocked(address) {
				sum += out.Value
				unspent_outs[tx_id] = append(unspent_outs[tx_id], out_idx)

				if sum >= amount {
					break main_loop
				}
			}
		}
	}

	return sum, unspent_outs
}

func (chain *BlockChain) GetBalance(address string) int {
	utxo := chain.FindUTXO(address)
	sum := 0

	fmt.Println(utxo)

	for _, it := range utxo {
		sum += it.Value
	}

	return sum
}

func (chain *BlockChain) Send(from, to string, amount int) {
	transaction := NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*Transaction{transaction})
}
