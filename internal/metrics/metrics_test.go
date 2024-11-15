package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewBlockMetrics(t *testing.T) {
	metrics := NewBlockMetrics()

	tests := []struct {
		name       string
		metric     prometheus.Collector
		metricName string
		help       string
		metricType string
	}{
		{
			name:       "TotalProposed",
			metric:     metrics.TotalProposed,
			metricName: "validator_total_blocks_proposed",
			help:       "Total number of blocks proposed by our validator",
			metricType: "counter",
		},
		{
			name:       "ExecutionConfirmed",
			metric:     metrics.ExecutionConfirmed,
			metricName: "validator_execution_blocks_confirmed",
			help:       "Number of proposed blocks that made it to the execution layer",
			metricType: "counter",
		},
		{
			name:       "ExecutionMissed",
			metric:     metrics.ExecutionMissed,
			metricName: "validator_execution_blocks_missed",
			help:       "Number of proposed blocks that failed to make it to the execution layer",
			metricType: "counter",
		},
		{
			name:       "EmptyConsensusBlocks",
			metric:     metrics.EmptyConsensusBlocks,
			metricName: "validator_empty_consensus_blocks",
			help:       "Number of blocks proposed with no transactions on consensus layer",
			metricType: "counter",
		},
		{
			name:       "EmptyExecutionBlocks",
			metric:     metrics.EmptyExecutionBlocks,
			metricName: "validator_empty_execution_blocks",
			help:       "Number of blocks confirmed on execution layer with no transactions",
			metricType: "counter",
		},
		{
			name:       "Errors",
			metric:     metrics.Errors,
			metricName: "validator_block_processing_errors",
			help:       "Number of errors encountered while processing blocks",
			metricType: "counter",
		},
		{
			name:       "CurrentHeight",
			metric:     metrics.CurrentHeight,
			metricName: "validator_current_block_height",
			help:       "Current block height being processed",
			metricType: "gauge",
		},
		{
			name:       "ElToClGap",
			metric:     metrics.ElToClGap,
			metricName: "validator_el_to_cl_gap",
			help:       "Gap between execution and consensus layer block heights",
			metricType: "gauge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := strings.TrimSpace(`
# HELP `+tt.metricName+` `+tt.help+`
# TYPE `+tt.metricName+` `+tt.metricType+`
`+tt.metricName+` 0
`) + "\n"
			err := testutil.CollectAndCompare(tt.metric, strings.NewReader(expected))
			if err != nil {
				t.Errorf("CollectAndCompare() error = %v", err)
			}
		})
	}
}

func TestMetricOperations(t *testing.T) {
	metrics := NewBlockMetrics()

	tests := []struct {
		name string

		operation func()
		verify    func(t *testing.T)
	}{
		{
			name: "increment counter",
			operation: func() {
				metrics.TotalProposed.Inc()
			},
			verify: func(t *testing.T) {
				if got := testutil.ToFloat64(metrics.TotalProposed); got != 1 {
					t.Errorf("Expected 1, got %f", got)
				}
			},
		},
		{
			name: "set gauge",
			operation: func() {
				metrics.CurrentHeight.Set(100)
			},
			verify: func(t *testing.T) {
				if got := testutil.ToFloat64(metrics.CurrentHeight); got != 100 {
					t.Errorf("Expected 100, got %f", got)
				}
			},
		},
		{
			name: "increment Errors twice",
			operation: func() {
				metrics.Errors.Inc()
				metrics.Errors.Inc()
			},
			verify: func(t *testing.T) {
				if got := testutil.ToFloat64(metrics.Errors); got != 2 {
					t.Errorf("Expected 2, got %f", got)
				}
			},
		},
		{
			name: "set ElToClGap",
			operation: func() {
				metrics.ElToClGap.Set(50)
			},
			verify: func(t *testing.T) {
				if got := testutil.ToFloat64(metrics.ElToClGap); got != 50 {
					t.Errorf("Expected 50, got %f", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.operation()
			tt.verify(t)
		})
	}
}
