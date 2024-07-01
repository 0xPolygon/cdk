package internal

import (
	"crypto/ecdsa"
	"math/big"
	"os"
	"path/filepath"

	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
)

type Etherman2AuthStore struct {
	chainID uint64
	auth    map[common.Address]bind.TransactOpts // empty in case of read-only client
}

func NewEtherman2L1Auth(chainID uint64) (*Etherman2AuthStore, error) {
	return &Etherman2AuthStore{
		chainID: chainID,
		auth:    map[common.Address]bind.TransactOpts{},
	}, nil
}

// getAuthByAddress tries to get an authorization from the authorizations map
func (etherMan *Etherman2AuthStore) GetAuthByAddress(addr common.Address) (bind.TransactOpts, bool) {
	auth, found := etherMan.auth[addr]
	return auth, found
}

// LoadAuthFromKeyStore loads an authorization from a key store file
// Store the authorization in the authorizations map and returns it
func (etherMan *Etherman2AuthStore) LoadAuthFromKeyStore(path, password string) (*bind.TransactOpts, *ecdsa.PrivateKey, error) {
	auth, pk, err := newAuthFromKeystore(path, password, etherMan.chainID)
	if err != nil {
		return nil, nil, err
	}

	log.Infof("loaded authorization for address: %v", auth.From.String())
	etherMan.auth[auth.From] = auth
	return &auth, pk, nil
}

// AddOrReplaceAuth adds an authorization or replace an existent one to the same account
func (etherMan *Etherman2AuthStore) AddOrReplaceAuth(auth bind.TransactOpts) error {
	log.Infof("added or replaced authorization for address: %v", auth.From.String())
	etherMan.auth[auth.From] = auth
	return nil
}

// newAuthFromKeystore an authorization instance from a keystore file
func newAuthFromKeystore(path, password string, chainID uint64) (bind.TransactOpts, *ecdsa.PrivateKey, error) {
	log.Infof("reading key from: %v", path)
	key, err := newKeyFromKeystore(path, password)
	if err != nil {
		return bind.TransactOpts{}, nil, err
	}
	if key == nil {
		return bind.TransactOpts{}, nil, nil
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key.PrivateKey, new(big.Int).SetUint64(chainID))
	if err != nil {
		return bind.TransactOpts{}, nil, err
	}
	return *auth, key.PrivateKey, nil
}

// newKeyFromKeystore creates an instance of a keystore key from a keystore file
func newKeyFromKeystore(path, password string) (*keystore.Key, error) {
	if path == "" && password == "" {
		return nil, nil
	}
	keystoreEncrypted, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	log.Infof("decrypting key from: %v", path)
	key, err := keystore.DecryptKey(keystoreEncrypted, password)
	if err != nil {
		return nil, err
	}
	return key, nil
}
