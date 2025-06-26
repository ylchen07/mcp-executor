# MCP Python Executor

An MCP (Model Context Protocol) server that provides secure Python code execution in isolated Docker containers with built-in Playwright support for web automation and scraping.

## Overview

This project implements an MCP server that exposes a single powerful tool: `execute-python`. The tool allows safe execution of Python code in ephemeral Docker containers using Microsoft's Playwright Python image, making it ideal for data analysis, web scraping, and automation tasks.

## Features

- 🐍 **Secure Python Execution**: Run Python code in isolated Docker containers
- 🎭 **Playwright Support**: Built-in browser automation and web scraping capabilities
- 📦 **Dynamic Module Installation**: Install Python packages on-the-fly per execution
- 🔄 **Dual Protocol Support**: Both stdio and SSE (Server-Sent Events) modes
- 🧹 **Ephemeral Environment**: Each execution starts with a clean container
- 🛡️ **Isolated Execution**: No persistence between runs for enhanced security

## Prerequisites

- **Go 1.23.3+**: Required to build and run the server
- **Docker**: Must be installed and running for Python code execution
- **Internet Connection**: Required for pulling Docker images and installing Python modules

## Installation

1. Clone the repository:

   ```bash
   git clone <repository-url>
   cd mcp-python
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Build the project:

   ```bash
   go build
   ```

## Usage

### Stdio Mode (Default)

Run the MCP server in stdio mode for direct integration with MCP clients:

```bash
go run main.go
```

### SSE Mode

Run the server with HTTP Server-Sent Events support:

```bash
go run main.go --sse
```

The SSE server will start on `http://localhost:8080`.

## Tool: execute-python

The server provides a single MCP tool with the following specification:

### Parameters

| Parameter | Type   | Required | Description                                       |
| --------- | ------ | -------- | ------------------------------------------------- |
| `code`    | string | Yes      | Python code to execute                            |
| `modules` | string | No       | Comma-separated list of Python modules to install |

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

```json
{
  "code": "from playwright.sync_api import sync_playwright\n\nwith sync_playwright() as p:\n    browser = p.chromium.launch()\n    page = browser.new_page()\n    page.goto('https://example.com')\n    title = page.title()\n    print(f'Page title: {title}')\n    browser.close()",
  "modules": "playwright"
}
```

## Architecture

The project follows a clean, modular architecture:

```
├── main.go                 # Entry point
├── cmd/
│   └── execute.go         # CLI command handling
├── internal/
│   ├── config/
│   │   └── config.go      # Configuration constants
│   ├── executor/
│   │   ├── executor.go    # Executor interface
│   │   └── docker.go      # Docker-based executor implementation
│   ├── server/
│   │   └── server.go      # MCP server setup and runners
│   └── tools/
│       └── python.go      # Python execution tool implementation
```

### Key Components

- **MCP Server**: Built using `github.com/mark3labs/mcp-go` library
- **Docker Executor**: Handles Python code execution in containers
- **Python Tool**: MCP tool implementation with parameter validation
- **Configuration**: Centralized constants and settings

## Configuration

Current configuration (in `internal/config/config.go`):

- **Server Name**: `python-executor`
- **Server Version**: `1.0.0`
- **Docker Image**: `mcr.microsoft.com/playwright/python:v1.53.0-noble`
- **SSE Port**: `:8080`
- **SSE Host**: `http://localhost:8080`

## Development

### Building

```bash
go build
```

### Running Tests

```bash
go test ./...
```

### Module Management

```bash
go mod tidy
```

## Security Considerations

- **Isolated Execution**: All Python code runs in separate Docker containers
- **No Persistence**: Containers are removed after each execution (`--rm` flag)
- **Limited Network**: Containers have minimal network access
- **Ephemeral State**: No data persists between executions

## Dependencies

- `github.com/mark3labs/mcp-go v0.8.2` - MCP protocol implementation
- `github.com/google/uuid v1.6.0` - UUID generation (indirect dependency)

## Docker Image

Uses Microsoft's official Playwright Python image:

- **Image**: `mcr.microsoft.com/playwright/python:v1.53.0-noble`
- **Includes**: Python 3.x, Playwright, and common browser binaries
- **OS**: Ubuntu Noble (24.04 LTS)

## Limitations

- **No State Persistence**: Variables and installed modules don't persist between executions
- **Docker Dependency**: Requires Docker to be running
- **Output Only**: Only stdout/stderr is returned; file system changes are not accessible
- **Resource Limits**: Subject to Docker container resource constraints

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

[License information not specified in source code]

## Support

For issues and questions, please refer to the project's issue tracker or documentation.

