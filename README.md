# MCP Executor

An MCP (Model Context Protocol) server that provides Python and Bash execution in either subprocess or isolated Docker environments. Built with Go and the Cobra CLI framework, featuring multiple transport modes, flexible execution modes, and built-in Playwright support for web automation.

## Overview

This project implements a robust MCP server that exposes two powerful tools: `execute-python` and `execute-bash`. These tools enable execution of Python code and bash scripts in either:

- **Subprocess mode** (default): Fast execution directly on the host machine
- **Docker mode**: Isolated execution in ephemeral Docker containers

Perfect for data analysis, web scraping, system administration, and automation tasks with the flexibility to choose between speed (subprocess) and isolation (Docker).

## Features

- ⚡ **Dual Execution Modes**: Choose between subprocess (fast) or Docker (isolated) execution
- 🐍 **Python Execution**: Run Python code with pip package installation support
- 🔧 **Bash Execution**: Execute shell commands and scripts
- 🎭 **Playwright Support**: Built-in browser automation in Docker mode
- 📦 **Dynamic Package Installation**: Install Python modules (subprocess/Docker) and Ubuntu packages (Docker only)
- 🔄 **Triple Protocol Support**: stdio, SSE (Server-Sent Events), and HTTP transport modes
- 🧹 **Clean Execution**: Subprocess mode or ephemeral Docker containers
- 🛡️ **Flexible Security**: Balance between speed (subprocess) and isolation (Docker)
- 📊 **Verbose Logging**: Optional detailed logging for debugging and monitoring
- 🚀 **CLI Framework**: Built with Cobra for robust command-line interface
- 🧪 **Comprehensive Testing**: Full test coverage with make targets

## Prerequisites

- **Go 1.23.3+**: Required to build and run the server
- **Python 3**: Required for subprocess mode (default)
- **Bash**: Required for bash subprocess execution (usually pre-installed)
- **Docker** (optional): Only required for Docker execution mode (`--execution-mode docker`)
- **Internet Connection**: Required for installing dependencies and pulling Docker images (Docker mode)

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/ylchen07/mcp-executor.git
   cd mcp-executor
   ```

2. Install dependencies:

   ```bash
   make deps
   ```

3. Build the project:

   ```bash
   make build
   ```

4. (Optional) Run tests:

   ```bash
   make test
   ```

## Usage

The server supports three transport modes and two execution modes. Use `make help` to see all available commands.

### Default Behavior (Stdio + Subprocess Mode)

Run the MCP server in stdio mode with subprocess execution (fastest, runs on host):

```bash
./bin/mcp-executor
# or
make run
```

### Execution Modes

#### Subprocess Mode (Default - Fast)

Code runs directly on host machine:

```bash
./bin/mcp-executor serve
# or explicitly specify
./bin/mcp-executor serve --execution-mode subprocess
```

#### Docker Mode (Isolated)

Code runs in isolated Docker containers:

```bash
./bin/mcp-executor serve --execution-mode docker
# Short flag
./bin/mcp-executor serve -e docker
```

### Transport Modes

#### SSE Mode

Run the server with HTTP Server-Sent Events support:

```bash
./bin/mcp-executor serve --mode sse
# With Docker execution
./bin/mcp-executor serve --mode sse --execution-mode docker
```

The SSE server will start on `http://localhost:8080`.

#### HTTP Mode

Run the server with streamable HTTP transport:

```bash
./bin/mcp-executor serve --mode http
# With subprocess execution (default)
./bin/mcp-executor serve --mode http --execution-mode subprocess
```

The HTTP server will start on `http://localhost:8081`.

### Combined Options

Combine transport and execution modes with verbose logging:

```bash
# SSE transport + Docker execution + verbose logging
./bin/mcp-executor serve -m sse -e docker -v

# HTTP transport + subprocess execution + verbose
./bin/mcp-executor serve --mode http --execution-mode subprocess --verbose
```

## Tools

The server provides two MCP tools that work in both subprocess and Docker execution modes:

### Tool: execute-python

Executes Python code in either subprocess (default) or Docker container based on server's `--execution-mode` setting.

**Execution Mode Differences:**

- **Subprocess**: Uses host's `python3`, installs packages with `pip install --user`
- **Docker**: Uses Playwright Python image, includes browser automation support

### Parameters

| Parameter | Type   | Required | Description                                                         |
| --------- | ------ | -------- | ------------------------------------------------------------------- |
| `code`    | string | Yes      | Python code to execute                                              |
| `modules` | string | No       | Comma-separated list of Python modules to install                   |
| `env`     | string | No       | Comma-separated KEY=VALUE pairs injected into execution environment |

### Example Usage

#### Basic Python Execution

```json
{
  "code": "print('Hello, World!')\nprint(2 + 2)"
}
```

#### With Module Installation

```json
{
  "code": "import requests\nresponse = requests.get('https://api.github.com')\nprint(response.status_code)",
  "modules": "requests"
}
```

#### Web Scraping with Playwright

> **Note**: Playwright browser automation works best in Docker mode (`--execution-mode docker`) as the Playwright image includes pre-installed browser binaries. In subprocess mode, you'll need browsers installed on your host system.

```json
{
  "code": "from playwright.sync_api import sync_playwright\n\nwith sync_playwright() as p:\n    browser = p.chromium.launch()\n    page = browser.new_page()\n    page.goto('https://example.com')\n    title = page.title()\n    print(f'Page title: {title}')\n    browser.close()",
  "modules": "playwright"
}
```

### Tool: execute-bash

Executes bash scripts in either subprocess (default) or Docker container based on server's `--execution-mode` setting.

**Execution Mode Differences:**

- **Subprocess**: Uses host's `bash`, package installation disabled for security
- **Docker**: Uses Ubuntu 22.04 container, supports `apt-get` package installation

#### Parameters

| Parameter  | Type   | Required | Description                                                           |
| ---------- | ------ | -------- | --------------------------------------------------------------------- |
| `script`   | string | Yes      | Bash script or commands to execute                                    |
| `packages` | string | No       | Comma-separated list of Ubuntu packages to install (Docker mode only) |
| `env`      | string | No       | Comma-separated KEY=VALUE pairs injected into execution environment   |

#### Example Usage

##### Basic Bash Execution

```json
{
  "script": "echo 'Hello, World!'\nls -la"
}
```

##### With Package Installation

> **Note**: Package installation via `packages` parameter only works in Docker mode (`--execution-mode docker`). In subprocess mode, required packages must be pre-installed on the host system.

```json
{
  "script": "curl -s https://api.github.com/repos/octocat/Hello-World | jq '.name'",
  "packages": "curl,jq"
}
```

##### System Information

```json
{
  "script": "echo 'System Info:'\nuname -a\ndf -h\nfree -h",
  "packages": ""
}
```

## Architecture

The project follows a clean, modular architecture built with the Cobra CLI framework:

```mermaid
graph TB
    subgraph "CLI Layer"
        A[main.go] --> B[cmd/root.go]
        B --> C[cmd/serve.go]
        C --> D[cmd/version.go]
    end

    subgraph "Server Layer"
        C --> E[internal/server/server.go]
        E --> F[MCP Server Instance]
    end

    subgraph "Tools Layer"
        F --> G[internal/tools/python.go]
        F --> H[internal/tools/bash.go]
    end

    subgraph "Execution Layer"
        G --> I[internal/executor/subprocess.go<br/>Default]
        G --> I2[internal/executor/docker.go<br/>Optional]
        H --> I
        H --> I2
        I --> J[Python/Bash Subprocess<br/>Host Machine]
        I2 --> K[Python Container<br/>Playwright Image]
        I2 --> L[Bash Container<br/>Ubuntu 22.04]
    end

    subgraph "Support Layer"
        E --> M[internal/config/config.go]
        E --> N[internal/logger/logger.go]
        C --> N
    end

    subgraph "Transport Modes"
        F --> O[Stdio Transport]
        F --> P[SSE Transport<br/>:8080]
        F --> Q[HTTP Transport<br/>:8081]
    end
```

### Directory Structure

```
├── main.go                    # Application entry point
├── Makefile                   # Build, test, and development commands
├── .gitignore                 # Git ignore rules
├── cmd/
│   ├── root.go               # Root command and CLI setup
│   ├── serve.go              # Serve command with execution-mode flag
│   └── version.go            # Version command
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration constants
│   ├── executor/
│   │   ├── executor.go       # Executor interface definition
│   │   ├── subprocess.go     # Subprocess executor (default)
│   │   ├── subprocess_test.go # Subprocess executor tests
│   │   └── docker.go         # Docker-based executor (optional)
│   ├── logger/
│   │   └── logger.go         # Logging utilities and verbose output
│   ├── server/
│   │   └── server.go         # MCP server setup with executor injection
│   └── tools/
│       ├── python.go         # Python execution tool implementation
│       └── bash.go           # Bash execution tool implementation
```

### Key Components

- **CLI Framework**: Built using `github.com/spf13/cobra` for robust command-line interface
- **MCP Server**: Built using `github.com/mark3labs/mcp-go` library with multiple transport support
- **Executor Interface**: Abstraction for different execution strategies (subprocess, Docker)
- **Subprocess Executor**: Default executor running code directly on host machine
- **Docker Executor**: Optional executor for isolated container execution
- **Dependency Injection**: Server selects executors based on `--execution-mode` flag
- **Python Tool**: MCP tool implementation working with any executor
- **Bash Tool**: MCP tool implementation working with any executor
- **Logger**: Centralized logging with verbose mode support
- **Configuration**: Centralized constants and settings
- **Makefile**: Comprehensive build, test, and development targets

## Configuration

### Server Configuration (`internal/config/config.go`)

- **Server Name**: `mcp-executor`
- **Server Version**: `1.0.0`
- **Transport Ports**:
  - SSE Port: `:8080` (`http://localhost:8080`)
  - HTTP Port: `:8081` (`http://localhost:8081`)
  - Stdio: Standard input/output (default)

### Execution Modes

#### Subprocess Mode (Default)

- **Python Binary**: `python3`
- **Bash Binary**: `bash`
- **Python Packages**: Installed via `pip install --user`
- **Bash Packages**: Not supported (must be pre-installed on host)
- **Environment**: Inherits from host + custom variables

#### Docker Mode (Optional)

- **Python Image**: `mcr.microsoft.com/playwright/python:v1.53.0-noble`
- **Bash Image**: `ubuntu:22.04`
- **Python Packages**: Installed via `pip install` in container
- **Bash Packages**: Installed via `apt-get install` in container
- **Environment**: Isolated container environment + custom variables

## Development

All development tasks are managed through the Makefile for consistency and ease of use.

### Available Make Targets

```bash
make help              # Show all available commands
make build             # Build the binary to bin/mcp-executor
make test              # Run tests with verbose output (no cache)
make test-coverage     # Run tests with coverage report
make fmt               # Format Go code
make lint              # Run golangci-lint
make deps              # Tidy Go dependencies
make run               # Run the application
make clean             # Remove build artifacts and cache
```

### Building

```bash
make build
```

### Running Tests

```bash
# Run all tests with verbose output
make test

# Generate coverage report
make test-coverage
# Opens coverage/coverage.html
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Manage dependencies
make deps
```

## Security Considerations

### Subprocess Mode (Default)

⚠️ **Security Warning**: Code runs directly on the host machine with user permissions

- **No Isolation**: Code has access to the host filesystem and environment
- **User Permissions**: Runs with the same permissions as the server process
- **Package Persistence**: Python packages installed via pip persist in user's cache
- **Bash Limitations**: System package installation disabled for safety
- **Recommendation**: Only use for trusted code or development environments

### Docker Mode

✅ **Enhanced Security**: Full container isolation

- **Isolated Execution**: Code runs in ephemeral Docker containers
- **No Persistence**: Containers are removed after each execution (`--rm` flag)
- **Default Networking**: Containers run with Docker's default bridge network
- **Ephemeral State**: No data persists between executions
- **Recommendation**: Use for untrusted code or production deployments

### Choosing the Right Mode

| Use Case            | Recommended Mode | Reason                             |
| ------------------- | ---------------- | ---------------------------------- |
| Development/Testing | Subprocess       | Faster execution, easier debugging |
| Trusted Scripts     | Subprocess       | No Docker overhead                 |
| Untrusted Code      | Docker           | Full isolation required            |
| Production          | Docker           | Enhanced security and consistency  |
| CI/CD Pipelines     | Either           | Depends on trust level             |

## Dependencies

- `github.com/mark3labs/mcp-go v0.32.0` - MCP protocol implementation
- `github.com/spf13/cobra v1.9.1` - CLI framework
- `github.com/google/uuid v1.6.0` - UUID generation (indirect dependency)
- `github.com/spf13/cast v1.9.2` - Type conversion utilities (indirect)
- `github.com/spf13/pflag v1.0.6` - Command-line flag parsing (indirect)

## Docker Images (Docker Mode Only)

When using `--execution-mode docker`, the following Docker images are used:

**Python Execution:**

- **Image**: `mcr.microsoft.com/playwright/python:v1.53.0-noble`
- **Includes**: Python 3.x, Playwright, and common browser binaries
- **OS**: Ubuntu Noble (24.04 LTS)
- **Use Case**: Web scraping, browser automation, complex Python tasks

**Bash Execution:**

- **Image**: `ubuntu:22.04`
- **Includes**: Standard Ubuntu utilities and package manager
- **OS**: Ubuntu 22.04 LTS
- **Use Case**: System administration, package installation, isolated bash scripts

## Limitations

### Subprocess Mode

- **No Isolation**: Code runs on host with full filesystem access
- **Package Persistence**: Python packages persist in user's pip cache (not truly ephemeral)
- **No Bash Packages**: Cannot install system packages for security reasons
- **Security Risk**: Untrusted code can potentially harm the host system

### Docker Mode

- **Docker Dependency**: Requires Docker to be running
- **Slower Performance**: Container startup overhead compared to subprocess
- **Image Size**: Large Docker images (Playwright image is ~1.5GB)
- **Resource Limits**: Subject to Docker container resource constraints

### Both Modes

- **Output Only**: Only stdout/stderr is returned; file system changes are not accessible
- **No State Persistence**: Variables don't persist between executions (except pip packages in subprocess mode)
- **Single Execution**: Each tool call is independent and isolated

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues and questions, please refer to the project's issue tracker or documentation.
