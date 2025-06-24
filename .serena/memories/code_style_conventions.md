# PATH Code Style and Conventions

## Go Style
- Standard Go formatting (gofmt)
- Uses golangci-lint for code quality (`make go_lint`)
- Package-level documentation with purpose and usage
- Structured logging with polylog
- Error handling follows Go conventions

## Naming Conventions
- Package names: lowercase, short, descriptive
- Constants: camelCase for unexported, PascalCase for exported
- Variables: camelCase
- Functions: PascalCase for exported, camelCase for unexported
- Interfaces: Usually end with -er suffix

## Metrics Conventions
- All metrics use `pathProcess = "path"` as subsystem
- Metric names defined as constants (e.g., `requestsTotal = "requests_total"`)
- Package-level metric variables using prometheus constructors
- Registration in `init()` functions using `prometheus.MustRegister()`
- Publishing through dedicated `PublishMetrics()` functions

## Documentation
- Comprehensive comments for exported functions and types
- TODO comments with assignee format: `TODO_MVP(@username): description`
- Usage examples in complex function comments

## Testing
- Unit tests with `-short` flag for quick testing
- E2E tests for integration scenarios
- Table-driven tests for multiple test cases
- Separate test files with `_test.go` suffix