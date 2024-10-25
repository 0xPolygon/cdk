package common

import (
	"encoding/binary"
	"math/big"

	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/keccak256"
)

// Uint64ToBytes converts a uint64 to a byte slice
func Uint64ToBytes(num uint64) []byte {
	const uint64ByteSize = 8

	bytes := make([]byte, uint64ByteSize)
	binary.BigEndian.PutUint64(bytes, num)

	return bytes
}

// BytesToUint64 converts a byte slice to a uint64
func BytesToUint64(bytes []byte) uint64 {
	return binary.BigEndian.Uint64(bytes)
}

// Uint32ToBytes converts a uint32 to a byte slice in big-endian order
func Uint32ToBytes(num uint32) []byte {
	const uint32ByteSize = 4

	key := make([]byte, uint32ByteSize)
	binary.BigEndian.PutUint32(key, num)

	return key
}

// BytesToUint32 converts a byte slice to a uint32
func BytesToUint32(bytes []byte) uint32 {
	return binary.BigEndian.Uint32(bytes)
}

// CalculateAccInputHash computes the hash of accumulated input data for a given batch.
func CalculateAccInputHash(
	logger *log.Logger,
	oldAccInputHash common.Hash,
	batchData []byte,
	l1InfoRoot common.Hash,
	timestampLimit uint64,
	sequencerAddr common.Address,
	forcedBlockhashL1 common.Hash,
) common.Hash {
	v1 := oldAccInputHash.Bytes()
	v2 := batchData
	v3 := l1InfoRoot.Bytes()
	v4 := big.NewInt(0).SetUint64(timestampLimit).Bytes()
	v5 := sequencerAddr.Bytes()
	v6 := forcedBlockhashL1.Bytes()

	// Add 0s to make values 32 bytes long
	for len(v1) < 32 {
		v1 = append([]byte{0}, v1...)
	}

	for len(v3) < 32 {
		v3 = append([]byte{0}, v3...)
	}

	for len(v4) < 8 {
		v4 = append([]byte{0}, v4...)
	}

	for len(v5) < 20 {
		v5 = append([]byte{0}, v5...)
	}

	for len(v6) < 32 {
		v6 = append([]byte{0}, v6...)
	}

	v2 = keccak256.Hash(v2)

	logger.Debugf("OldAccInputHash: %v", oldAccInputHash)
	logger.Debugf("BatchHashData: %v", common.Bytes2Hex(v2))
	logger.Debugf("L1InfoRoot: %v", l1InfoRoot)
	logger.Debugf("TimeStampLimit: %v", timestampLimit)
	logger.Debugf("Sequencer Address: %v", sequencerAddr)
	logger.Debugf("Forced BlockHashL1: %v", forcedBlockhashL1)

	return common.BytesToHash(keccak256.Hash(v1, v2, v3, v4, v5, v6))
}
