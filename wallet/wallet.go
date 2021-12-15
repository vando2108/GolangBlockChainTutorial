package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const (
	check_sum_length = 4
	version          = byte(0x00)
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func CreateWallet() *Wallet {
	private_key, public_key := NewKeyPair()
	wallet := &Wallet{private_key, public_key}
	return wallet
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	private_key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	public_key := append(private_key.X.Bytes(), private_key.PublicKey.Y.Bytes()...)

	return *private_key, public_key
}

func (w Wallet) Address() []byte {
	public_hash := PublicKeyHash(w.PublicKey)
	version_hash := append([]byte{version}, public_hash...)
	check_sum := CheckSum(version_hash)
	full_hash := append(version_hash, check_sum...)

	return Base58Encode(full_hash)
}

func PublicKeyHash(public_key []byte) []byte {
	public_key_hash := sha256.Sum256(public_key)

	hasher := ripemd160.New()
	_, err := hasher.Write(public_key_hash[:])
	if err != nil {
		log.Panic(err)
	}

	return hasher.Sum(nil)
}

func CheckSum(payload []byte) []byte {
	first_hash := sha256.Sum256(payload)
	second_hash := sha256.Sum256(first_hash[:])

	return second_hash[:check_sum_length]
}
