package common

import "encoding/binary"

// BlockNum2Bytes converts a block number to a byte slice
func BlockNum2Bytes(blockNum uint64) []byte {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, blockNum)

	return key
}

// Bytes2BlockNum converts a byte slice to a block number
func Bytes2BlockNum(key []byte) uint64 {
	return binary.LittleEndian.Uint64(key)
}
