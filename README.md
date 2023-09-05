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
- [ ] Algorithm selection for header signing (e.g., HMAC, RSA).
- [ ] Implementation of the signing process using a secret key.
  
## 2. Features and Enhancements
### Exponential Backoff
- [X] Implement an exponential backoff mechanism.
- [ ] Tests to validate exponential backoff behavior.

### Queuing
- [X] Implement or integrate a queuing mechanism.
- [ ] Ensure data integrity and no data loss during processing.

### Retry Mechanism
- [X] Implement a retry mechanism upon failure.
- [X] Define max retry count and intervals.

### Logging
- [ ] Logger setup for the project.
- [ ] Implement file-based persistent logging.

## 3. Additional Considerations
### Security
- [ ] Implementation of password security or other authentication methods for Redis.
- [ ] Encryption and secure data handling standards.
- [ ] Implement measures against common security threats like DDoS, data injections, etc.

### Scalability
- [ ] Strategy for handling high concurrency using goroutines and worker pools.

## Documentation and Community Building
- [ ] Comprehensive documentation for setup, features, and usage.
- [ ] Contribution guidelines for the community.
- [ ] Discord community

## Testing and Release
- [ ] Comprehensive test coverage for all features.
- [ ] Continuous Integration (CI) setup.
- [ ] First major release with all the initial features.
 - [ ] Distribution in Dockerhub 
