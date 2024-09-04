package db

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"

	tree "github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/russross/meddler"
	"modernc.org/sqlite"
)

// initMeddler registers tags to be used to read/write from SQL DBs using meddler
func initMeddler() {
	meddler.Default = meddler.SQLite
	meddler.Register("bigint", BigIntMeddler{})
	meddler.Register("merkleproof", MerkleProofMeddler{})
	meddler.Register("hash", HashMeddler{})
}

func SQLiteErr(err error) (*sqlite.Error, bool) {
	if sqliteErr, ok := err.(*sqlite.Error); ok {
		return sqliteErr, true
	}
	if driverErr, ok := meddler.DriverErr(err); ok {
		if sqliteErr, ok := driverErr.(*sqlite.Error); ok {
			return sqliteErr, true
		}
	}
	return nil, false
}

// SliceToSlicePtrs converts any []Foo to []*Foo
func SliceToSlicePtrs(slice interface{}) interface{} {
	v := reflect.ValueOf(slice)
	vLen := v.Len()
	typ := v.Type().Elem()
	res := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(typ)), vLen, vLen)
	for i := 0; i < vLen; i++ {
		res.Index(i).Set(v.Index(i).Addr())
	}
	return res.Interface()
}

// SlicePtrsToSlice converts any []*Foo to []Foo
func SlicePtrsToSlice(slice interface{}) interface{} {
	v := reflect.ValueOf(slice)
	vLen := v.Len()
	typ := v.Type().Elem().Elem()
	res := reflect.MakeSlice(reflect.SliceOf(typ), vLen, vLen)
	for i := 0; i < vLen; i++ {
		res.Index(i).Set(v.Index(i).Elem())
	}
	return res.Interface()
}

// BigIntMeddler encodes or decodes the field value to or from JSON
type BigIntMeddler struct{}

// PreRead is called before a Scan operation for fields that have the BigIntMeddler
func (b BigIntMeddler) PreRead(fieldAddr interface{}) (scanTarget interface{}, err error) {
	// give a pointer to a byte buffer to grab the raw data
	return new(string), nil
}

// PostRead is called after a Scan operation for fields that have the BigIntMeddler
func (b BigIntMeddler) PostRead(fieldPtr, scanTarget interface{}) error {
	ptr := scanTarget.(*string)
	if ptr == nil {
		return fmt.Errorf("BigIntMeddler.PostRead: nil pointer")
	}
	field := fieldPtr.(**big.Int)
	var ok bool
	*field, ok = new(big.Int).SetString(*ptr, 10)
	if !ok {
		return fmt.Errorf("big.Int.SetString failed on \"%v\"", *ptr)
	}
	return nil
}

// PreWrite is called before an Insert or Update operation for fields that have the BigIntMeddler
func (b BigIntMeddler) PreWrite(fieldPtr interface{}) (saveValue interface{}, err error) {
	field := fieldPtr.(*big.Int)

	return field.String(), nil
}

// MerkleProofMeddler encodes or decodes the field value to or from JSON
type MerkleProofMeddler struct{}

// PreRead is called before a Scan operation for fields that have the ProofMeddler
func (b MerkleProofMeddler) PreRead(fieldAddr interface{}) (scanTarget interface{}, err error) {
	// give a pointer to a byte buffer to grab the raw data
	return new(string), nil
}

// PostRead is called after a Scan operation for fields that have the ProofMeddler
func (b MerkleProofMeddler) PostRead(fieldPtr, scanTarget interface{}) error {
	ptr := scanTarget.(*string)
	if ptr == nil {
		return fmt.Errorf("ProofMeddler.PostRead: nil pointer")
	}
	field := fieldPtr.(*tree.Proof)
	strHashes := strings.Split(*ptr, ",")
	if len(strHashes) != int(tree.DefaultHeight) {
		return fmt.Errorf("unexpected len of hashes: expected %d actual %d", tree.DefaultHeight, len(strHashes))
	}
	for i, strHash := range strHashes {
		field[i] = common.HexToHash(strHash)
	}
	return nil
}

// PreWrite is called before an Insert or Update operation for fields that have the ProofMeddler
func (b MerkleProofMeddler) PreWrite(fieldPtr interface{}) (saveValue interface{}, err error) {
	field := fieldPtr.(tree.Proof)
	var s string
	for _, f := range field {
		s += f.Hex() + ","
	}
	s = strings.TrimSuffix(s, ",")
	return s, nil
}

// HashMeddler encodes or decodes the field value to or from JSON
type HashMeddler struct{}

// PreRead is called before a Scan operation for fields that have the ProofMeddler
func (b HashMeddler) PreRead(fieldAddr interface{}) (scanTarget interface{}, err error) {
	// give a pointer to a byte buffer to grab the raw data
	return new(string), nil
}

// PostRead is called after a Scan operation for fields that have the ProofMeddler
func (b HashMeddler) PostRead(fieldPtr, scanTarget interface{}) error {
	ptr := scanTarget.(*string)
	if ptr == nil {
		return fmt.Errorf("HashMeddler.PostRead: nil pointer")
	}
	field := fieldPtr.(*common.Hash)
	*field = common.HexToHash(*ptr)
	return nil
}

// PreWrite is called before an Insert or Update operation for fields that have the ProofMeddler
func (b HashMeddler) PreWrite(fieldPtr interface{}) (saveValue interface{}, err error) {
	field := fieldPtr.(common.Hash)
	return field.Hex(), nil
}
