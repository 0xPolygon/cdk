package l1infotreesync

import (
	"context"
	"errors"
	"testing"

	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestGetL1InfoTreeMerkleProof(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, _, err := s.GetL1InfoTreeMerkleProof(context.Background(), 0)
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetRollupExitTreeMerkleProof(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetRollupExitTreeMerkleProof(context.Background(), 0, common.Hash{})
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetLatestInfoUntilBlock(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetLatestInfoUntilBlock(context.Background(), 0)
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetInfoByIndex(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetInfoByIndex(context.Background(), 0)
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetL1InfoTreeRootByIndex(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetL1InfoTreeRootByIndex(context.Background(), 0)
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetLastRollupExitRoot(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetLastRollupExitRoot(context.Background())
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetLastL1InfoTreeRoot(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetLastL1InfoTreeRoot(context.Background())
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetLastProcessedBlock(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetLastProcessedBlock(context.Background())
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetLocalExitRoot(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetLocalExitRoot(context.Background(), 0, common.Hash{})
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetLastVerifiedBatches(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetLastVerifiedBatches(0)
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetFirstVerifiedBatches(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetFirstVerifiedBatches(0)
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetFirstVerifiedBatchesAfterBlock(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetFirstVerifiedBatchesAfterBlock(0, 0)
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetFirstL1InfoWithRollupExitRoot(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetFirstL1InfoWithRollupExitRoot(common.Hash{})
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetLastInfo(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetLastInfo()
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetFirstInfo(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetFirstInfo()
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetFirstInfoAfterBlock(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetFirstInfoAfterBlock(0)
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}

func TestGetL1InfoTreeMerkleProofFromIndexToRoot(t *testing.T) {
	s := L1InfoTreeSync{
		processor: &processor{
			halted: true,
		},
	}
	_, err := s.GetL1InfoTreeMerkleProofFromIndexToRoot(context.Background(), 0, common.Hash{})
	require.Error(t, err)
	require.True(t, errors.Is(err, sync.ErrInconsistentState))
}
