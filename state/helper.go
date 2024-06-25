package state

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/0xPolygon/cdk/hex"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	double       = 2
	ether155V    = 27
	etherPre155V = 35
	// MaxEffectivePercentage is the maximum value that can be used as effective percentage
	MaxEffectivePercentage = uint8(255)
	// Decoding constants
	headerByteLength uint64 = 1
	sLength          uint64 = 32
	rLength          uint64 = 32
	vLength          uint64 = 1
	c0               uint64 = 192 // 192 is c0. This value is defined by the rlp protocol
	ff               uint64 = 255 // max value of rlp header
	shortRlp         uint64 = 55  // length of the short rlp codification
	f7               uint64 = 247 // 192 + 55 = c0 + shortRlp

	// EfficiencyPercentageByteLength is the length of the effective percentage in bytes
	EfficiencyPercentageByteLength uint64 = 1
)

func prepareRLPTxData(tx *types.Transaction) ([]byte, error) {
	v, r, s := tx.RawSignatureValues()
	sign := 1 - (v.Uint64() & 1)

	nonce, gasPrice, gas, to, value, data, chainID := tx.Nonce(), tx.GasPrice(), tx.Gas(), tx.To(), tx.Value(), tx.Data(), tx.ChainId()

	rlpFieldsToEncode := []interface{}{
		nonce,
		gasPrice,
		gas,
		to,
		value,
		data,
	}

	if !IsPreEIP155Tx(tx) {
		rlpFieldsToEncode = append(rlpFieldsToEncode, chainID)
		rlpFieldsToEncode = append(rlpFieldsToEncode, uint(0))
		rlpFieldsToEncode = append(rlpFieldsToEncode, uint(0))
	}

	txCodedRlp, err := rlp.EncodeToBytes(rlpFieldsToEncode)
	if err != nil {
		return nil, err
	}

	newV := new(big.Int).Add(big.NewInt(ether155V), big.NewInt(int64(sign)))
	newRPadded := fmt.Sprintf("%064s", r.Text(hex.Base))
	newSPadded := fmt.Sprintf("%064s", s.Text(hex.Base))
	newVPadded := fmt.Sprintf("%02s", newV.Text(hex.Base))
	txData, err := hex.DecodeString(hex.EncodeToString(txCodedRlp) + newRPadded + newSPadded + newVPadded)
	if err != nil {
		return nil, err
	}
	return txData, nil
}

// DecodeTxs extracts Transactions for its encoded form
func DecodeTxs(txsData []byte, forkID uint64) ([]*types.Transaction, []byte, []uint8, error) {
	// Process coded txs
	var pos uint64
	var txs []*types.Transaction
	var efficiencyPercentages []uint8
	txDataLength := uint64(len(txsData))
	if txDataLength == 0 {
		return txs, txsData, nil, nil
	}
	for pos < txDataLength {
		num, err := strconv.ParseUint(hex.EncodeToString(txsData[pos:pos+1]), hex.Base, hex.BitSize64)
		if err != nil {
			log.Debug("error parsing header length: ", err)
			return []*types.Transaction{}, txsData, []uint8{}, err
		}
		// First byte is the length and must be ignored
		if num < c0 {
			log.Debugf("error num < c0 : %d, %d", num, c0)
			return []*types.Transaction{}, txsData, []uint8{}, ErrInvalidData
		}
		length := num - c0
		if length > shortRlp { // If rlp is bigger than length 55
			// n is the length of the rlp data without the header (1 byte) for example "0xf7"
			if (pos + 1 + num - f7) > txDataLength {
				log.Debug("error parsing length: ", err)
				return []*types.Transaction{}, txsData, []uint8{}, err
			}
			n, err := strconv.ParseUint(hex.EncodeToString(txsData[pos+1:pos+1+num-f7]), hex.Base, hex.BitSize64) // +1 is the header. For example 0xf7
			if err != nil {
				log.Debug("error parsing length: ", err)
				return []*types.Transaction{}, txsData, []uint8{}, err
			}
			if n+num < f7 {
				log.Debug("error n + num < f7: ", err)
				return []*types.Transaction{}, txsData, []uint8{}, ErrInvalidData
			}
			length = n + num - f7 // num - f7 is the header. For example 0xf7
		}

		endPos := pos + length + rLength + sLength + vLength + headerByteLength

		if forkID >= FORKID_DRAGONFRUIT {
			endPos += EfficiencyPercentageByteLength
		}

		if endPos > txDataLength {
			err := fmt.Errorf("endPos %d is bigger than txDataLength %d", endPos, txDataLength)
			log.Debug("error parsing header: ", err)
			return []*types.Transaction{}, txsData, []uint8{}, ErrInvalidData
		}

		if endPos < pos {
			err := fmt.Errorf("endPos %d is smaller than pos %d", endPos, pos)
			log.Debug("error parsing header: ", err)
			return []*types.Transaction{}, txsData, []uint8{}, ErrInvalidData
		}

		if endPos < pos {
			err := fmt.Errorf("endPos %d is smaller than pos %d", endPos, pos)
			log.Debug("error parsing header: ", err)
			return []*types.Transaction{}, txsData, []uint8{}, ErrInvalidData
		}

		fullDataTx := txsData[pos:endPos]
		dataStart := pos + length + headerByteLength
		txInfo := txsData[pos:dataStart]
		rData := txsData[dataStart : dataStart+rLength]
		sData := txsData[dataStart+rLength : dataStart+rLength+sLength]
		vData := txsData[dataStart+rLength+sLength : dataStart+rLength+sLength+vLength]

		if forkID >= FORKID_DRAGONFRUIT {
			efficiencyPercentage := txsData[dataStart+rLength+sLength+vLength : endPos]
			efficiencyPercentages = append(efficiencyPercentages, efficiencyPercentage[0])
		}

		pos = endPos

		// Decode rlpFields
		var rlpFields [][]byte
		err = rlp.DecodeBytes(txInfo, &rlpFields)
		if err != nil {
			log.Error("error decoding tx Bytes: ", err, ". fullDataTx: ", hex.EncodeToString(fullDataTx), "\n tx: ", hex.EncodeToString(txInfo), "\n Txs received: ", hex.EncodeToString(txsData))
			return []*types.Transaction{}, txsData, []uint8{}, ErrInvalidData
		}

		legacyTx, err := RlpFieldsToLegacyTx(rlpFields, vData, rData, sData)
		if err != nil {
			log.Debug("error creating tx from rlp fields: ", err, ". fullDataTx: ", hex.EncodeToString(fullDataTx), "\n tx: ", hex.EncodeToString(txInfo), "\n Txs received: ", hex.EncodeToString(txsData))
			return []*types.Transaction{}, txsData, []uint8{}, err
		}

		tx := types.NewTx(legacyTx)
		txs = append(txs, tx)
	}
	return txs, txsData, efficiencyPercentages, nil
}

// DecodeTx decodes a string rlp tx representation into a types.Transaction instance
func DecodeTx(encodedTx string) (*types.Transaction, error) {
	b, err := hex.DecodeHex(encodedTx)
	if err != nil {
		return nil, err
	}

	tx := new(types.Transaction)
	if err := tx.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return tx, nil
}

// IsPreEIP155Tx checks if the tx is a tx that has a chainID as zero and
// V field is either 27 or 28
func IsPreEIP155Tx(tx *types.Transaction) bool {
	v, _, _ := tx.RawSignatureValues()
	return tx.ChainId().Uint64() == 0 && (v.Uint64() == 27 || v.Uint64() == 28)
}
