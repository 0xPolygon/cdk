package localbridgesync

import (
	"context"
	"math/big"
	"testing"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridge"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	contractAddr = common.HexToAddress("1234567890")
)

const (
	syncBlockChunck = 10
)

func TestGetEventsByBlockRange(t *testing.T) {
	clientMock := NewL2Mock(t)
	ctx := context.Background()
	d, err := newDownloader(contractAddr, syncBlockChunck, clientMock)
	require.NoError(t, err)
	log, bridge := generateRandomBridge(t)
	logs := []types.Log{
		*log,
	}
	clientMock.
		On("FilterLogs", mock.Anything, mock.Anything).
		Return(logs, nil)

	blocks := d.getEventsByBlockRange(ctx, 0, 1)
	b := blocks[0]
	assert.Equal(t, log.BlockHash, b.Hash)
	assert.Equal(t, log.BlockNumber, b.Num)
	assert.Equal(t, bridge, b.Events.Bridges[0])
}

func generateRandomBridge(t *testing.T) (*types.Log, Bridge) {
	b := Bridge{
		LeafType:           1,
		OriginNetwork:      1,
		OriginAddress:      contractAddr,
		DestinationNetwork: 1,
		DestinationAddress: contractAddr,
		Amount:             big.NewInt(1),
		Metadata:           common.Hex2Bytes("01"),
		DepositCount:       1,
	}
	abi, err := polygonzkevmbridge.PolygonzkevmbridgeMetaData.GetAbi()
	require.NoError(t, err)
	event, err := abi.EventByID(bridgeEventSignature)
	require.NoError(t, err)
	data, err := event.Inputs.Pack(
		b.LeafType,
		b.OriginNetwork,
		b.OriginAddress,
		b.DestinationNetwork,
		b.DestinationAddress,
		b.Amount,
		b.Metadata,
		b.DepositCount,
	)
	require.NoError(t, err)
	log := &types.Log{
		Address:     contractAddr,
		BlockNumber: 1,
		BlockHash:   common.HexToHash("01"),
		Topics:      []common.Hash{bridgeEventSignature},
		Data:        data,
	}
	return log, b
}
