# plat-v2-tools

Platform v2 tools for managing WordPress applications in Nomad clusters.

## Tools

### backup-data-gen

A tool for generating realistic backup data distributions on WordPress applications running in Nomad.

## Building

### Prerequisites

- Go 1.23.8 or later

### Build Instructions

1. Clone the repository:

   ```bash
   git clone <repository-url>
   cd plat-v2-tools
   ```

2. Download dependencies:

   ```bash
   go mod download
   ```

3. Build the backup-data-gen tool:

   ```bash
   go build -o backup-data-gen ./cmd/backup-data-gen
   ```

4. (Optional) Install the tool to your Go bin directory:

   ```bash
   go install ./cmd/backup-data-gen
   ```

## Testing

Run the test suite:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

## Usage

Most plat-v2-tools will require port forward to connect to the test environment's Nomad servers. Before running plat-v2-tools, export the Nomad Client ENV vars:

```bash
export NOMAD_ADDR=https://127.0.0.1:4646
export NOMAD_SKIP_VERIFY=true
```

After building, see the tool-specific documentation for usage instructions:

- [backup-data-gen](./docs/backu-data-generator.md) - Generate random files and directories for backup agent load testing

## Development

### Project Structure

```text
plat-v2-tools/
├── cmd/                    # Command line applications
│   └── backup-data-gen/    # Backup data generator tool
├── pkg/                    # Reusable packages
│   └── utils/              # Utility packages
│       ├── appexec/        # Nomad app execution utilities
│       └── datagen/        # Data generation utilities
├── go.mod                  # Go module definition
└── go.sum                  # Go module checksums
```

### Adding New Tools

1. Create a new directory under `cmd/` for your tool
2. Add a `main.go` file with your application logic
3. Update this README with build instructions for the new tool
4. Add tool-specific documentation in the tool's directory
