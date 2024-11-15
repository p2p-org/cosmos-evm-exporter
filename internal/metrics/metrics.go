package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type BlockMetrics struct {
	Registry             *prometheus.Registry
	TotalProposed        prometheus.Counter
	ExecutionConfirmed   prometheus.Counter
	ExecutionMissed      prometheus.Counter
	EmptyConsensusBlocks prometheus.Counter
	EmptyExecutionBlocks prometheus.Counter
	Errors               prometheus.Counter
	CurrentHeight        prometheus.Gauge
	ElToClGap            prometheus.Gauge
}

func NewBlockMetrics() *BlockMetrics {
	registry := prometheus.NewRegistry()

	return &BlockMetrics{
		Registry: registry,
		TotalProposed: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "validator_total_blocks_proposed",
			Help: "Total number of blocks proposed by our validator",
		}),
		ExecutionConfirmed: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "validator_execution_blocks_confirmed",
			Help: "Number of proposed blocks that made it to the execution layer",
		}),
		ExecutionMissed: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "validator_execution_blocks_missed",
			Help: "Number of proposed blocks that failed to make it to the execution layer",
		}),
		EmptyConsensusBlocks: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "validator_empty_consensus_blocks",
			Help: "Number of blocks proposed with no transactions on consensus layer",
		}),
		EmptyExecutionBlocks: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "validator_empty_execution_blocks",
			Help: "Number of blocks confirmed on execution layer with no transactions",
		}),
		Errors: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "validator_block_processing_errors",
			Help: "Number of errors encountered while processing blocks",
		}),
		CurrentHeight: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "validator_current_block_height",
			Help: "Current block height being processed",
		}),
		ElToClGap: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "validator_el_to_cl_gap",
			Help: "Gap between execution and consensus layer block heights",
		}),
	}
}
