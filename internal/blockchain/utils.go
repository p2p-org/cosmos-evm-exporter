package blockchain

import "fmt"

func (p *BlockProcessor) GetCurrentGap() (int64, error) {
	elHeight, err := p.GetCurrentELHeight()
	if err != nil {
		return 0, fmt.Errorf("failed to get EL height: %v", err)
	}

	clHeight, err := p.GetCurrentHeight()
	if err != nil {
		return 0, fmt.Errorf("failed to get CL height: %v", err)
	}

	return clHeight - elHeight, nil
}

func (p *BlockProcessor) DumpPayload(payload []byte) {
	p.logger.WriteJSONLog("debug", "Payload dump", map[string]interface{}{
		"length": len(payload),
		"hex":    fmt.Sprintf("%x", payload),
	}, nil)

	for i := 0; i < len(payload); i += 32 {
		end := i + 32
		if end > len(payload) {
			end = len(payload)
		}
		p.logger.WriteJSONLog("debug", "Payload chunk", map[string]interface{}{
			"start": i,
			"end":   end - 1,
			"hex":   fmt.Sprintf("%x", payload[i:end]),
		}, nil)
	}
}
