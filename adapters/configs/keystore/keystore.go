package keystore

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"io/ioutil"
	"sync"
)

var ErrUnknownChain = errors.New("unknown chain")

type Keystore struct {
	EthereumKey EthereumKey
	BitcoinKey  BitcoinKey

	mu   *sync.RWMutex
	path string
}

type Key interface {
	GetKey() *ecdsa.PrivateKey
	GetKeyString() (string, error)
	GetAddress() ([]byte, error)
	PriorityCode() uint32
}

func LoadKeystore(path string) (Keystore, error) {
	var keystore Keystore
	keystore.path = path
	keystore.mu = new(sync.RWMutex)
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return keystore, err
	}
	json.Unmarshal(raw, &keystore)
	return keystore, nil
}

func (keystore *Keystore) Update() error {
	data, err := json.Marshal(keystore)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(keystore.path, data, 700)
}
