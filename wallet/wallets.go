package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
)

const (
	wallet_file = "./temp/wallets.data"
)

type Wallets struct {
	Wallets map[string]*Wallet
}

func CreateWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFile()
	return &wallets, err
}

func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func (ws *Wallets) AddWallet(w *Wallet) {
	address := w.Address()
	ws.Wallets[string(address[:])] = w
}

func (ws Wallets) ListAddress() []string {
	var ret []string
	for _, w := range ws.Wallets {
		ret = append(ret, string(w.Address()[:]))
	}
	return ret
}

func (ws *Wallets) SaveFile() {
	var content bytes.Buffer
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	if err := encoder.Encode(ws); err != nil {
		log.Panic(err)
	}

	if err := ioutil.WriteFile(wallet_file, content.Bytes(), 0644); err != nil {
		log.Panic(err)
	}
}

func (ws *Wallets) LoadFile() error {
	if _, err := os.Stat(wallet_file); os.IsNotExist(err) {
		return err
	}

	var wallets Wallets
	file_content, err := ioutil.ReadFile(wallet_file)
	if err != nil {
		return err
	}
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(file_content))
	if err := decoder.Decode(&wallets); err != nil {
		return err
	}

	ws.Wallets = wallets.Wallets

	return nil
}
