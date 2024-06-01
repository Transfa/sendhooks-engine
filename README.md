# SendHooks Engine
![GitHub Tag](https://img.shields.io/github/v/tag/transfa/sendhooks-engine)
[![Build and Release](https://github.com/Transfa/sendhooks-engine/actions/workflows/release.yml/badge.svg)](https://github.com/Transfa/sendhooks-engine/actions/workflows/release.yml)
[![Go](https://github.com/Transfa/sendhooks-engine/actions/workflows/go.yml/badge.svg)](https://github.com/Transfa/sendhooks-engine/actions/workflows/go.yml)
[![CodeQL](https://github.com/Transfa/sendhooks-engine/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/Transfa/sendhooks-engine/actions/workflows/github-code-scanning/codeql)

> ⚠️ **Warning**: This project is not stable yet. Feel free to join the Discord so we can build together [https://discord.gg/w2fBSz3j](https://discord.gg/2mHxEgxm5r)

## Introduction
SendHooks Engine is a high-performance, scalable tool for managing and delivering webhooks. It uses modern architectural patterns to efficiently handle webhook processing, making it an ideal choice for applications requiring reliable and quick webhook delivery. Written in Go, SendHooks takes full advantage of the language's concurrency and performance features.

## Key Features
- **High Performance**: Utilizes Golang's efficient concurrency model for fast webhook processing.
- **Scalable Architecture**: Designed to handle varying loads with a queue-based approach.
- **Secure**: Supports SSL/TLS for Redis connections and authenticates webhook messages for security.

## Installation

### Using Docker
To run SendHooks using Docker, a `config.json` file is required. This file contains necessary configurations including the Redis server address.

1. **Pull the Docker Image**:
   ```bash
   docker pull transfa/sendhooks:latest
   ```

2. **Run the Docker Image with `config.json`**:
   ```bash
   docker run -v /path/to/config.json:/app/config.json -t transfa/sendhooks
   ```
Replace `/path/to/config.json` with the path to your configuration file.

### Using the Compiled Binary (macOS and Linux)

Ensure that the `config.json` file is located in the same directory as the binary.

1. **Download and Run the Binary**:
   ```bash
   curl -LO https://github.com/Transfa/sendhooks-engine/releases/latest/download/sendhooks
   chmod +x sendhooks
   ./sendhooks
   ```

## Components
- **Redis Client**: Interacts with Redis streams to manage incoming webhook messages.
- **Queue**: Buffers messages ensuring smooth flow and handling of sudden influxes.
- **HTTP Client**: Processes each message, sending it as an HTTP POST request to the intended URL.

## Concurrency Handling
- Utilizes Go routines for simultaneous operations, enhancing efficiency and responsiveness.
- Employs a buffered channel (queue) for concurrent message handling.

## Security
- Configurable for SSL/TLS encryption with Redis, ensuring secure message transmission.
- Includes a secret hash in HTTP headers for message authenticity verification.

## Contributors
We welcome contributions from the community. If you'd like to contribute, please check out our [list of issues](https://github.com/koladev32/sendhooks-engine/issues) to see how you can help.

You can find how to contribute in the [CONTRIBUTING](CONTRIBUTING.md) file.

## Security

You can find our security policies on [SECURITY](SECURITY.md).
