package operations

import (
	"errors"
	"fmt"
	"time"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
)

type ConfirmationLevel string

const (
	Trusted  ConfirmationLevel = "trusted"
	Virtual  ConfirmationLevel = "virtual"
	Verified ConfirmationLevel = "verified"
)

func WaitTxToBeMined(hash ethgo.Hash, level ConfirmationLevel, client *jsonrpc.Client, timeout time.Duration) error {
	startingTime := time.Now()
	fmt.Printf("waiting for tx %s to be included in the %s state\n", hash.String(), level)
	var (
		encodedL2BlockNumber string
		isVirtualized        bool
		isVerified           bool
	)
	for time.Since(startingTime) < timeout {
		if encodedL2BlockNumber == "" {
			receipt, err := client.Eth().GetTransactionReceipt(hash)
			if err != nil || receipt == nil {
				time.Sleep(time.Second)
				continue
			}
			fmt.Printf("tx %s included in the trusted state at the L2 block %d\n", hash.String(), receipt.BlockNumber)
			if level == Trusted {
				return nil
			}
			encodedL2BlockNumber = fmt.Sprintf("0x%x", receipt.BlockNumber)
		}

		if !isVirtualized {
			if err := client.Call("zkevm_isBlockVirtualized", &isVirtualized, encodedL2BlockNumber); err != nil {
				time.Sleep(time.Second)
				continue
			}
			if isVirtualized {
				fmt.Printf("tx %s included in the virtual state\n", hash.String())
				if level == Virtual {
					return nil
				}
			} else {
				time.Sleep(time.Second)
				continue
			}
		}
		if err := client.Call("zkevm_isBlockConsolidated", &isVerified, encodedL2BlockNumber); err != nil {
			time.Sleep(time.Second)
			continue
		}
		if isVerified {
			fmt.Printf("tx %s included in the verified state\n", hash.String())
			return nil
		}
	}
	return errors.New("timed out")
}