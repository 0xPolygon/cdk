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
)

func TestStreamTx(t *testing.T) {
	tx1, err := state.DecodeTx(txStreamEncoded1)
	require.NoError(t, err)
	tx2, err := state.DecodeTx(txStreamEncoded2)
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

func TestKK(t *testing.T) {
	sut := &SequenceSender{}
	sequences, err := sut.getSequencesToSend()
	require.NoError(t, err)
	require.NotNil(t, sequences)
}
