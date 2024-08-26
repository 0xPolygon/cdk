package reorgdetector

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestBlockMap(t *testing.T) {
	t.Parallel()

	// Create a new block map
	bm := newHeadersList(
		header{Num: 1, Hash: common.HexToHash("0x123")},
		header{Num: 2, Hash: common.HexToHash("0x456")},
		header{Num: 3, Hash: common.HexToHash("0x789")},
	)

	t.Run("len", func(t *testing.T) {
		t.Parallel()

		actualLen := bm.len()
		expectedLen := 3
		if !reflect.DeepEqual(expectedLen, actualLen) {
			t.Errorf("len() returned incorrect result, expected: %v, got: %v", expectedLen, actualLen)
		}
	})

	t.Run("isEmpty", func(t *testing.T) {
		t.Parallel()

		if bm.isEmpty() {
			t.Error("isEmpty() returned incorrect result, expected: false, got: true")
		}
	})

	t.Run("add", func(t *testing.T) {
		t.Parallel()

		tba := header{Num: 4, Hash: common.HexToHash("0xabc")}
		bm.add(tba)
		if !reflect.DeepEqual(tba, bm.headers[4]) {
			t.Errorf("add() returned incorrect result, expected: %v, got: %v", tba, bm.headers[4])
		}
	})

	t.Run("copy", func(t *testing.T) {
		t.Parallel()

		copiedBm := bm.copy()
		if !reflect.DeepEqual(bm, copiedBm) {
			t.Errorf("add() returned incorrect result, expected: %v, got: %v", bm, copiedBm)
		}
	})

	t.Run("get", func(t *testing.T) {
		t.Parallel()

		if !reflect.DeepEqual(*bm.get(3), bm.headers[3]) {
			t.Errorf("get() returned incorrect result, expected: %v, got: %v", bm.get(3), bm.headers[3])
		}
	})

	t.Run("getSorted", func(t *testing.T) {
		t.Parallel()

		sortedBlocks := bm.getSorted()
		expectedSortedBlocks := []header{
			{Num: 1, Hash: common.HexToHash("0x123")},
			{Num: 2, Hash: common.HexToHash("0x456")},
			{Num: 3, Hash: common.HexToHash("0x789")},
		}
		if !reflect.DeepEqual(sortedBlocks, expectedSortedBlocks) {
			t.Errorf("getSorted() returned incorrect result, expected: %v, got: %v", expectedSortedBlocks, sortedBlocks)
		}
	})

	t.Run("removeRange", func(t *testing.T) {
		t.Parallel()

		bm := newHeadersList(
			header{Num: 1, Hash: common.HexToHash("0x123")},
			header{Num: 2, Hash: common.HexToHash("0x456")},
			header{Num: 3, Hash: common.HexToHash("0x789")},
			header{Num: 4, Hash: common.HexToHash("0xabc")},
			header{Num: 5, Hash: common.HexToHash("0xdef")},
		)

		bm.removeRange(3, 5)

		expectedBlocks := []header{
			{Num: 1, Hash: common.HexToHash("0x123")},
			{Num: 2, Hash: common.HexToHash("0x456")},
		}

		sortedBlocks := bm.getSorted()

		if !reflect.DeepEqual(sortedBlocks, expectedBlocks) {
			t.Errorf("removeRange() failed, expected: %v, got: %v", expectedBlocks, sortedBlocks)
		}
	})
}
