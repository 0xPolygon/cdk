package reorgdetector

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestBlockMap(t *testing.T) {
	t.Parallel()

	// Create a new block map
	bm := newBlockMap(
		block{Num: 1, Hash: common.HexToHash("0x123")},
		block{Num: 2, Hash: common.HexToHash("0x456")},
		block{Num: 3, Hash: common.HexToHash("0x789")},
	)

	t.Run("getSorted", func(t *testing.T) {
		t.Parallel()

		sortedBlocks := bm.getSorted()
		expectedSortedBlocks := []block{
			{Num: 1, Hash: common.HexToHash("0x123")},
			{Num: 2, Hash: common.HexToHash("0x456")},
			{Num: 3, Hash: common.HexToHash("0x789")},
		}
		if !reflect.DeepEqual(sortedBlocks, expectedSortedBlocks) {
			t.Errorf("getSorted() returned incorrect result, expected: %v, got: %v", expectedSortedBlocks, sortedBlocks)
		}
	})

	t.Run("getFromBlockSorted", func(t *testing.T) {
		t.Parallel()

		fromBlockSorted := bm.getFromBlockSorted(2)
		expectedFromBlockSorted := []block{
			{Num: 3, Hash: common.HexToHash("0x789")},
		}
		if !reflect.DeepEqual(fromBlockSorted, expectedFromBlockSorted) {
			t.Errorf("getFromBlockSorted() returned incorrect result, expected: %v, got: %v", expectedFromBlockSorted, fromBlockSorted)
		}

		// Test getFromBlockSorted function when blockNum is greater than the last block
		fromBlockSorted = bm.getFromBlockSorted(4)
		expectedFromBlockSorted = []block{}
		if !reflect.DeepEqual(fromBlockSorted, expectedFromBlockSorted) {
			t.Errorf("getFromBlockSorted() returned incorrect result, expected: %v, got: %v", expectedFromBlockSorted, fromBlockSorted)
		}
	})

	t.Run("getClosestHigherBlock", func(t *testing.T) {
		t.Parallel()

		bm := newBlockMap(
			block{Num: 1, Hash: common.HexToHash("0x123")},
			block{Num: 2, Hash: common.HexToHash("0x456")},
			block{Num: 3, Hash: common.HexToHash("0x789")},
		)

		// Test when the blockNum exists in the block map
		b, exists := bm.getClosestHigherBlock(2)
		require.True(t, exists)
		expectedBlock := block{Num: 2, Hash: common.HexToHash("0x456")}
		if b != expectedBlock {
			t.Errorf("getClosestHigherBlock() returned incorrect result, expected: %v, got: %v", expectedBlock, b)
		}

		// Test when the blockNum does not exist in the block map
		b, exists = bm.getClosestHigherBlock(4)
		require.False(t, exists)
		expectedBlock = block{Num: 0, Hash: common.Hash{}}
		if b != expectedBlock {
			t.Errorf("getClosestHigherBlock() returned incorrect result, expected: %v, got: %v", expectedBlock, b)
		}
	})

	t.Run("removeRange", func(t *testing.T) {
		t.Parallel()

		bm := newBlockMap(
			block{Num: 1, Hash: common.HexToHash("0x123")},
			block{Num: 2, Hash: common.HexToHash("0x456")},
			block{Num: 3, Hash: common.HexToHash("0x789")},
			block{Num: 4, Hash: common.HexToHash("0xabc")},
			block{Num: 5, Hash: common.HexToHash("0xdef")},
		)

		bm.removeRange(3, 5)

		expectedBlocks := []block{
			{Num: 1, Hash: common.HexToHash("0x123")},
			{Num: 2, Hash: common.HexToHash("0x456")},
		}

		sortedBlocks := bm.getSorted()

		if !reflect.DeepEqual(sortedBlocks, expectedBlocks) {
			t.Errorf("removeRange() failed, expected: %v, got: %v", expectedBlocks, sortedBlocks)
		}
	})
}
