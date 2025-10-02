High-performance URL shortening service in Go supporting 3.5B+ unique URLs with sub-500ns encoding latency. Built with Redis for persistence, base62 encoding for compact URLs, and comprehensive test coverage, including unit tests to integration tests.

The architecture prioritizes:
- Modularity - Clean separation between encoding, storage, and serving layers
- Testability - Interface-based design with both unit and integration tests
- Performance - Benchmarked encoding and concurrent load testing (1,000+ requests)
- Scalability - Stateless design ready for horizontal scaling
