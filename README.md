# sendhooks-engine

## Roadmap

## 1. Core Functionality

### Data Integration

- [X] Integration with Redis for data input.
- [X] Set up a Redis channel for ingesting the data.

### Sending Data

- [X] HTTP client setup to send data to the endpoint.
- [X] Header configuration for the HTTP client.

### Header Signing

- [X] Algorithm selection for header signing (e.g., HMAC, RSA).
- [X] Implementation of the signing process using a secret key.
  
## 2. Features and Enhancements

### Exponential Backoff

- [X] Implement an exponential backoff mechanism.
- [X] Tests to validate exponential backoff behavior.

### Queuing

- [X] Implement or integrate a queuing mechanism.

### Retry Mechanism

- [X] Implement a retry mechanism upon failure.
- [X] Define max retry count and intervals.

### Logging

- [X] Logger setup for the project.
- [X] Implement file-based persistent logging.

## 3. Additional Considerations

### Security

- [X] Implementation of password security or other authentication methods for Redis.

### Scalability

- [X] Strategy for handling high concurrency using goroutines and worker pools.

## Documentation and Community Building

- [ ] Comprehensive documentation for setup, features, and usage.
- [ ] Contribution guidelines for the community.
- [ ] Discord community

## Testing and Release

- [X] Comprehensive test coverage for all features.
- [X] Continuous Integration (CI) setup.
- [X] First major release with all the initial features.
- [X] Distribution in Dockerhub
