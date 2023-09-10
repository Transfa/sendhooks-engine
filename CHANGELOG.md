# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased] - yyyy-mm-dd

## [0.0.1] - 2023-09-10

### Added

- Dockerize Project for Distribution on DockerHub (#8)
- Add tests suite for the Webhook sender package (#19)
- Refactor ProcessWebhooks for clarity, safety, and testability (#3 )
- Implement Payload Signing with webhook-signature Header in SendWebhook (#6)
- Implement File Logging Module for Enhanced Error Handling (#7)
- Refactor Subscribe function in the redis package for better error handling and modularity( #4 )
- Improve Payload Flexibility to Accept Various JSON Formats( #1 )
- Fix calculateBackoff redaclarring function (#16 )
- Refactor SendWebhook function in the sender package for enhanced modularity, error handling, and logging ( #5 )
