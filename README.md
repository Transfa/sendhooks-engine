# SendHooks Engine
> ⚠️ **Warning**: This project is not stable yet. Feel free to join the Discord so we can build together https://discord.gg/w2fBSz3j

## Introduction
SendHooks Engine is a powerful tool designed to handle webhook-related sending tasks. This document provides instructions on how to set up and run the project.

## Installation

### Using Docker

1. **Pull the Docker Image**:
   ```bash
   docker pull koladev32/sendhooks:latest
   ```

2. **Run the Docker Image**:
   ```bash
   docker run -t sendhooks --env REDIS_ADDRESS=<REDIS_ADDRESS:PORT> koladev32/sendhooks
   ```

### Using the Compiled Binary (macOS and Linux)

1. **Download and Run the Binary**:
   ```bash
   curl -LO https://github.com/koladev32/sendhooks-engine/releases/download/v0.0.1/webhook
   chmod +x webhook
   ./webhook
   ```

## Contributors
We welcome contributions from the community. If you'd like to contribute, please check out our [list of issues](https://github.com/koladev32/sendhooks-engine/issues) to see how you can help.
