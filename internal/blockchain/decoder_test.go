package blockchain

import (
	"encoding/base64"
	"encoding/binary"
	"testing"

	"cosmos-evm-exporter/internal/logger"
)

func TestDecodeTx(t *testing.T) {
	testLogger := logger.NewLogger(&logger.Config{
		EnableStdout: true,
		LogFile:      "test.log",
	})

	tests := []struct {
		name        string
		input       []byte
		wantMsgType uint32
		wantLength  uint32
		wantPayload []byte
		wantErr     bool
	}{
		{
			name: "valid transaction",
			input: func() []byte {
				data := make([]byte, 16)
				binary.BigEndian.PutUint32(data[0:4], 1) // MsgType
				binary.BigEndian.PutUint32(data[4:8], 8) // DataLength
				copy(data[8:], []byte("testdata"))       // Payload
				return data
			}(),
			wantMsgType: 1,
			wantLength:  8,
			wantPayload: []byte("testdata"),
			wantErr:     false,
		},
		{
			name:        "empty input",
			input:       []byte{},
			wantMsgType: 0,
			wantLength:  0,
			wantPayload: nil,
			wantErr:     true,
		},
		{
			name:        "too short input",
			input:       []byte{0x00, 0x00, 0x00, 0x01},
			wantMsgType: 0,
			wantLength:  0,
			wantPayload: nil,
			wantErr:     true,
		},
		{
			name: "payload length mismatch",
			input: func() []byte {
				data := make([]byte, 8)
				binary.BigEndian.PutUint32(data[0:4], 1)  // MsgType
				binary.BigEndian.PutUint32(data[4:8], 16) // DataLength (too large)
				return data
			}(),
			wantMsgType: 0,
			wantLength:  0,
			wantPayload: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode test input as base64
			input := base64.StdEncoding.EncodeToString(tt.input)

			got, err := DecodeTx(input, testLogger)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if got.MsgType != tt.wantMsgType {
				t.Errorf("MsgType = %v, want %v", got.MsgType, tt.wantMsgType)
			}
			if got.DataLength != tt.wantLength {
				t.Errorf("DataLength = %v, want %v", got.DataLength, tt.wantLength)
			}
			if string(got.Payload) != string(tt.wantPayload) {
				t.Errorf("Payload = %v, want %v", string(got.Payload), string(tt.wantPayload))
			}
		})
	}
}
