package blockchain

import (
	"cosmos-evm-exporter/internal/logger"
	"encoding/base64"
	"encoding/binary"
	"fmt"
)

func DecodeTx(txBase64 string, logger *logger.Logger) (*EVMChainTx, error) {
	// Decode base64
	txData, err := base64.StdEncoding.DecodeString(txBase64)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %v", err)
	}

	logger.WriteJSONLog("debug", "Raw tx data", map[string]interface{}{
		"length": len(txData),
	}, nil)

	if len(txData) < 8 {
		return nil, fmt.Errorf("tx data too short: %d bytes", len(txData))
	}

	tx := &EVMChainTx{
		MsgType:    binary.BigEndian.Uint32(txData[0:4]),
		DataLength: binary.BigEndian.Uint32(txData[4:8]),
	}

	if len(txData) < 8+int(tx.DataLength) {
		return nil, fmt.Errorf("payload length mismatch: expected %d, got %d", tx.DataLength, len(txData)-8)
	}

	tx.Payload = make([]byte, tx.DataLength)
	copy(tx.Payload, txData[8:8+tx.DataLength])

	return tx, nil
}
