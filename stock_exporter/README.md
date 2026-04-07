# Stock Exporter

## Overview
The Stock Exporter is a Prometheus exporter designed to collect and expose stock market metrics. It gathers data from various stock APIs and provides a metrics endpoint for Prometheus to scrape.

```
Kite WebSocket → OnTick → TickStore (sync.RWMutex) → StockCollector → /metrics
```

## Features
- Collects stock-related metrics from multiple sources.
- Implements the Prometheus Collector interface for easy integration.
- Configurable through environment variables and configuration files.

## Project Structure
```
stock_exporter
├── cmd
│   └── main.go          # Entry point of the application
├── collector
│   ├── collector.go     # Defines the Collector struct
│   └── stock.go         # Logic for collecting stock metrics
├── config
│   └── config.go        # Configuration structure and loading methods
├── internal
│   └── client
│       └── stock_client.go # Responsible for API calls to fetch stock data
├── go.mod                # Module dependencies
├── go.sum                # Checksums for module dependencies
├── Makefile              # Build instructions and commands
├── Dockerfile            # Instructions for building a Docker image
└── README.md             # Project documentation
```

## Installation
To install the Stock Exporter, clone the repository and run the following commands:

```bash
go mod tidy
```

Kite WebSocket → OnTick → TickStore (sync.RWMutex) → StockCollector → /metrics## Usage
To run the Stock Exporter, execute the following command:

```bash
go run cmd/main.go
```

The exporter will start an HTTP server and expose metrics at the `/metrics` endpoint.

## Configuration
The Stock Exporter can be configured using environment variables or a configuration file. Refer to `config/config.go` for available configuration options.

## Contributing
Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License
This project is licensed under the MIT License. See the LICENSE file for more details.