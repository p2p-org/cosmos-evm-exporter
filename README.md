# Cosmos EVM Exporter

A monitoring tool that tracks block production on Cosmos EVM chains consensus and execution layers, specifically designed to monitor validator block proposals and their corresponding execution blocks.

## Features

- Monitors validator block proposals on the consensus layer
- Tracks corresponding blocks on the execution layer
- Provides Prometheus metrics for monitoring
- Configurable via TOML configuration file
- Real-time logging of block production
- Tracks gaps between consensus and execution layers

## Metrics

The exporter provides the following Prometheus metrics:

- `validator_total_blocks_proposed`: Total number of blocks proposed by the validator
- `validator_execution_blocks_confirmed`: Number of blocks confirmed on execution layer
- `validator_execution_blocks_missed`: Number of blocks that failed to make it to execution layer
- `validator_empty_consensus_blocks`: Number of empty blocks on consensus layer
- `validator_empty_execution_blocks`: Number of empty blocks on execution layer
- `validator_block_processing_errors`: Number of errors encountered
- `validator_current_block_height`: Current block height being processed
- `validator_el_to_cl_gap`: Gap between execution and consensus layer heights

## Configuration

Create a `config.toml` file:

```toml
evm_address = "" # Validator's EVM address
target_validator = "" # Validator's consensus address
rpc_endpoint = "" # Consensus layer RPC
eth_endpoint = "" # Execution layer RPC
log_file = "block_monitor.log" # Log file path
metrics_port = ":2113" # Prometheus metrics port
enable_file_log = false # Enable file logging
enable_stdout = true # Enable console logging
```

## Usage

```bash
# Run with config file
go run ./cmd/exporter/main.go --config=./config.toml

# Build binary
go build -o evm-exporter ./cmd/exporter

# Run binary
./evm-exporter --config=./config.toml
```

## Metrics Endpoint

Prometheus metrics are available at `http://localhost:2113/metrics`

## Requirements

- Go 1.22.1 or later
- Access to Cosmos & Ethereum RPC endpoints

## Dependencies

Main dependencies:

- github.com/BurntSushi/toml: Configuration file parsing
- github.com/ethereum/go-ethereum: Ethereum client
- github.com/prometheus/client_golang: Prometheus metrics

## Building from Source

```bash
# Clone repository
git clone [repository-url]

# Install dependencies
go mod download

# Build
go build -o evm-exporter ./cmd/exporter
```

## Example Output

```
{"time":"2024-11-15T10:03:54Z","level":"info","message":"ℹ️ Starting exporter","data":{"metrics_port":":2113"}}
{"time":"2024-11-15T10:53:34Z","level":"debug","message":"Processing block","data":{"height":"7458296","proposer_address":"B2A5C37E25E52A994550C504E4227A9CBB60F61A"}}
{"time":"2024-11-15T10:53:34Z","level":"info","message":"ℹ️ Found validator block","data":{"height":7458296,"proposer_address":"B2A5C37E25E52A994550C504E4227A9CBB60F61A"}}
{"time":"2024-11-15T10:53:34Z","level":"success","message":"✅ Found execution block","data":{"cl_height":7458296,"el_height":6892471,"hash":"0xcf98515011a8245cf680492b57fe22fa042ef963fd0d733b8061d362d1f7ef5b"}}
```

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Commit Messages

This project uses [Conventional Commits](https://www.conventionalcommits.org/) for semantic versioning. Please format your commit messages according to the following rules:

- `feat: description` - for new features (triggers minor version bump)
- `fix: description` - for bug fixes (triggers patch version bump)
- `feat!: description` or `fix!: description` - for breaking changes (triggers major version bump)
- `chore: description` - for maintenance tasks (no version bump)
- `docs: description` - for documentation updates (no version bump)
- `test: description` - for test updates (no version bump)
- `refactor: description` - for code refactoring (no version bump)

Examples:

```bash
git commit -m "feat: add new metric for tracking block confirmations"
git commit -m "fix: correct calculation of EL to CL gap"
git commit -m "docs: update configuration examples"
```

### Pull Request Process

1. Fork the repository and create your branch from `main`
2. Update the README.md with details of changes if applicable
3. Add tests for any new functionality
4. Ensure all tests pass and the build succeeds
5. Update any relevant documentation
6. Submit a pull request

### Release Process

The project uses semantic-release for automated versioning and changelog generation. Upon merging to main:

1. Commit messages are analyzed
2. Version is automatically bumped based on conventional commits
3. CHANGELOG.md is generated/updated
4. Release is created with built binaries
5. Git tags are created
