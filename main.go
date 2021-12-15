package main

import (
	"fmt"
	"log"

	"github.com/tensor-programming/golang-blockchain/wallet"
)

func main() {
	ws, err := wallet.CreateWallets()
	if err != nil {
		log.Panic(err)
	}
	w := wallet.CreateWallet()
	ws.AddWallet(w)
	addresses := ws.ListAddress()
	for _, address := range addresses {
		fmt.Println(address)
	}
}
