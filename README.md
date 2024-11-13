# Berachain Block Exporter

A monitoring tool that tracks block production on Berachain's consensus and execution layers, specifically designed to monitor validator block proposals and their corresponding execution blocks.

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
go run main.go --config=./config.toml

# Build binary
go build -o berachain-exporter

# Run binary
./berachain-exporter --config=./config.toml
```

## Metrics Endpoint

Prometheus metrics are available at `http://localhost:2113/metrics`

## Requirements

- Go 1.22.1 or later
- Access to Berachain RPC endpoints

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
go build -o berachain-exporter
```

## Example Output

```
=== Processing Validator Block ===
CL Height: 7368349
Expected EL Height: 6808651 (gap: 559698)
✅ Found our block!
EL Height: 6808651
Hash: 0xb91ff1929453776e373832be64f9083a3de9aba962187b1919a778297971ae73
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
