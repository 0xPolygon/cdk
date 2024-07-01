package operations

import(
	"encoding/hex"
	"github.com/umbracle/ethgo/wallet"
)

const (
	L1RPCURL                = "http://el-1-geth-lighthouse:8545"
	L2ChainID               = 10101

	// Prefunded account present on kurtosis-cdk
	PrivateKeyWithFundsOnL2 = "12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"
)

var L2RPCURL = "http://localhost:8123"

func GetL2Wallet() (*wallet.Key, error) {
	prvKeyBytes, err := hex.DecodeString(PrivateKeyWithFundsOnL2)
	if err != nil {
		return nil, err
	}
	return wallet.NewWalletFromPrivKey(prvKeyBytes)
}