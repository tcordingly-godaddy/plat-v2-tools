# Backup Data Generator

A tool for generating realistic backup data distributions on WordPress applications running in Nomad. This tool creates files with various sizes and distributions to simulate real-world backup scenarios for testing purposes.

## Overview

The backup data generator creates files with configurable size distributions across different categories (small, medium, large files) to simulate realistic WordPress backup data. It can operate on single jobs, all jobs for an account, or run custom commands.

## Usage

```bash
./backup-data-gen [flags]
```

### Command Line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-jobId` | string | "" | Nomad job ID to execute commands on (optional) |
| `-accountId` | string | "" | Account ID to find all jobs for (optional) |
| `-cmd` | string | "" | Custom command to run on the app (optional) |
| `-size` | string | "medium" | Size distribution for backup generation: medium or large |
| `-rootDir` | string | "./wp-content/backup-gen" | Base root directory for backup generation |
| `-maxFiles` | int | 30 | Maximum files per directory |

### Size Distributions

The tool supports two predefined size distributions:

#### Medium Distribution (Default)
- **Total Size**: 300MB - 2GB
- **Large files** (1MB-5MB): 10% of total
- **Medium files** (400KB-1MB): 60% of total
- **Small files** (150KB-400KB): 30% of total

#### Large Distribution
- **Total Size**: 5GB - 10GB
- **Large files** (1MB-20MB): 45% of total
- **Medium files** (400KB-1MB): 35% of total
- **Small files** (150KB-400KB): 20% of total

## Usage Examples

### Generate backup data on a specific job
```bash
./backup-data-gen -jobId app-12345 -size medium
```

### Generate backup data on all jobs for an account
```bash
./backup-data-gen -accountId acc-67890 -size large
```

### Run a custom command on a specific job
```bash
./backup-data-gen -jobId app-12345 -cmd "ls -la ./wp-content"
```

### Generate large distribution with custom settings
```bash
./backup-data-gen -jobId app-12345 -size large -rootDir "./custom-backup" -maxFiles 50
```

### Generate backup data for all jobs in an account with custom directory
```bash
./backup-data-gen -accountId acc-67890 -rootDir "./test-backups" -maxFiles 25
```

## How It Works

1. **Job Discovery**: If an `accountId` is provided, the tool discovers all Nomad jobs for that account
2. **Command Generation**: Based on the size distribution, the tool generates shell commands to create files with random names and sizes
3. **Remote Execution**: Commands are executed on the target Nomad jobs using the Nomad exec API
4. **Concurrent Processing**: Multiple jobs can be processed concurrently for efficiency

## File Generation Process

The tool generates files using the following process:

1. Creates the base directory structure (`mkdir -p`)
2. Generates files with random names (15-25 characters, alphanumeric)
3. Uses `head -c` to create files with random binary data from `/dev/urandom`
4. Distributes files across size categories (small, medium, large) based on the chosen distribution
5. Continues generating until the total size for each category reaches its limit

## Output

The tool provides detailed logging including:
- Command execution status for each job
- Exit codes from remote commands
- Stdout/stderr output from executed commands
- Error handling for failed operations

## Prerequisites

- Access to a Nomad cluster
- Appropriate permissions to execute commands on Nomad jobs
- Target jobs must have shell access (`/bin/sh` or equivalent)
- Target systems must have `head` and `mkdir` commands available

## Error Handling

The tool includes comprehensive error handling for:
- Invalid size distribution parameters
- Nomad client connection failures
- Job discovery failures
- Command execution failures
- Network connectivity issues

## Performance Considerations

- Commands are executed concurrently across multiple jobs
- File generation uses efficient `head -c` commands with binary data
- Directory structure is optimized to avoid filesystem limitations
- Maximum file limits prevent excessive directory sizes
