# Contributing to SendHooks Engine

Thank you for considering contributing to SendHooks Engine! We appreciate your help in improving and maintaining this project.

## Table of Contents

- [Introduction](#introduction)
- [Development Environment Setup](#development-environment-setup)
- [Submitting Pull Requests](#submitting-pull-requests)
- [Reporting Bugs or Proposing Features](#reporting-bugs-or-proposing-features)
- [Running Tests](#running-tests)
- [Project-Specific Guidelines and Expectations](#project-specific-guidelines-and-expectations)

## Introduction

This guide outlines the process for contributing to SendHooks Engine. We encourage contributions from the community to make our project better.

## Development Environment Setup

Before you start contributing, make sure you have set up your development environment:

```bash
git clone https://github.com/Transfa/sendhooks-engine.git && cd sendhooks-engine/sendhooks

go mod download
```

At the moment, we are working with Go 1.20.

## Submitting Pull Requests

If you want to contribute code changes or bug fixes, please follow these steps:

1. Fork the SendHooks Engine repository.
2. Create a new branch for your changes: `git checkout -b feature/your-feature-name`.
3. Make your changes and commit them.
4. Push your changes to your forked repository.
5. Submit a pull request to the `main` branch of the SendHooks Engine repository.

## Reporting Bugs or Proposing Features

If you find a bug or want to propose a new feature, please:

1. Check if the issue already exists in our [issue tracker](https://github.com/Transfa/sendhooks-engine/issues).
2. If not, create a new issue and provide detailed information.

## Docker

If you prefer using Docker for development, you can set up the project as follows:

1. **Pull the Docker Image**:

```bash
cd webhooks

docker build . -t sendhooks
```

Then run the container.

```bash
docker run -t sendhooks --env REDIS_ADDRESS=<REDIS_ADDRESS:PORT> sendhooks
```

## Running Tests

To ensure the reliability and correctness of your code changes, it's important to run tests before submitting your pull request.

```bash
go test -v ./...
```

## Project-Specific Guidelines and Expectations

For any project-specific guidelines or expectations, please refer to our [README](README.md).

We appreciate your contributions and look forward to working with you to improve SendHooks Engine!

---
**Note**: This contributing guide is subject to updates and changes. We encourage you to check this document periodically for any updates.
