package sequencesender

import (
	"testing"

	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

const (
	txStreamEncoded1 = "f86508843b9aca0082520894617b3a3528f9cdd6630fd3301b9c8911f7bf063d0a808207f5a0579b72a1c1ffdd845fba45317540982109298e2ec8d67ddf2cdaf22e80903677a01831e9a01291c7ea246742a5b5a543ca6938bfc3f6958c22be06fad99274e4ac"
	txStreamEncoded2 = "f86509843b9aca0082520894617b3a3528f9cdd6630fd3301b9c8911f7bf063d0a808207f5a0908a522075e09485166ffa7630cd2b7013897fa1f1238013677d6f0a86efb3d2a0068b12435fcdc8ee254f3b1df8c5b29ed691eeee6065704f061130935976ca99"
	txStreamEncoded3 = "b8b402f8b101268505d21dba0085076c363d8982dc60941929761e87667283f087ea9ab8370c174681b4e980b844095ea7b300000000000000000000000080a64c6d7f12c47b7c66c5b4e20e72bc1fcd5d9effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc001a0dd4db494969139a120e8721842455ec13f82757a4fc49b66d447c7d32d095a1da06ef54068a9aa67ecc4f52d885299a04feb6f3531cdfc771f1412cd3331d1ba4c"
)

func TestStreamTx(t *testing.T) {
	tx1, err := state.DecodeTx(txStreamEncoded1)
	require.NoError(t, err)
	tx2, err := state.DecodeTx(txStreamEncoded2)
	require.NoError(t, err)
	tx3, err := state.DecodeTx(txStreamEncoded3)
	require.NoError(t, err)

	txTest := state.L2TxRaw{
		EfficiencyPercentage: 129,
		TxAlreadyEncoded:     false,
		Tx:                   tx1,
	}
	txTestEncoded := make([]byte, 0)
	txTestEncoded, err = txTest.Encode(txTestEncoded)
	require.NoError(t, err)
	log.Debugf("%s", common.Bytes2Hex(txTestEncoded))

	batch := state.BatchRawV2{
		Blocks: []state.L2BlockRaw{
			{
				ChangeL2BlockHeader: state.ChangeL2BlockHeader{
					DeltaTimestamp:  3633752,
					IndexL1InfoTree: 0,
				},
				Transactions: []state.L2TxRaw{
					{
						EfficiencyPercentage: 129,
						TxAlreadyEncoded:     false,
						Tx:                   tx1,
					},
					{
						EfficiencyPercentage: 97,
						TxAlreadyEncoded:     false,
						Tx:                   tx2,
					},
					{
						EfficiencyPercentage: 97,
						TxAlreadyEncoded:     false,
						Tx:                   tx3,
					},
				},
			},
		},
	}

	printBatch(&batch, true, true)

	encodedBatch, err := state.EncodeBatchV2(&batch)
	require.NoError(t, err)

	decodedBatch, err := state.DecodeBatchV2(encodedBatch)
	require.NoError(t, err)

	printBatch(decodedBatch, true, true)
}
