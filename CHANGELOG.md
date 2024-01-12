# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased] - yyyy-mm-dd

### Added

- Persisting data about the webhooks (#42)
- Integration of Logrus for Improved Logging (#55)
- Send response details in case of webhook failed (#39)

### Fixed

## [0.2.0-beta] - 2023-11-23

### Added

- Using JSON for configuration instead of environment variables (#44)
- Migrate to Redis streams (#45)

### Fixed

## [0.1.0] - 2023-10-31

### Added

- Make Header Name Dynamic through Environment Variable (#40)
- Implement Redis Channel for Webhook Delivery Status Updates (#11)
- Add a .env.example file (#36)
- Add Conditional SSL Support for Redis Connection (#21)
- Adding a CONTRIBUTING.md file (#27)
- Adding a SECURITY.md file( #29)


## [0.0.1] - 2023-09-10

### Added

- Dockerize Project for Distribution on DockerHub (#8)
- Add tests suite for the Webhook sender package (#19)
- Refactor ProcessWebhooks for clarity, safety, and testability (#3 )
- Implement Payload Signing with webhook-signature Header in SendWebhook (#6)
- Implement File Logging Module for Enhanced Error Handling (#7)
- Refactor Subscribe function in the redis package for better error handling and modularity( #4 )
- Improve Payload Flexibility to Accept Various JSON Formats( #1 )
- Refactor SendWebhook function in the sender package for enhanced modularity, error handling, and logging ( #5 )

### Fixed

- Fix calculateBackoff redaclarring function (#16 )
