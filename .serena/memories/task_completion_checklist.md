# Task Completion Checklist

## Code Quality
- [ ] Run linting: `make go_lint`
- [ ] Ensure code follows established patterns and conventions
- [ ] Add appropriate documentation and comments
- [ ] Follow the existing metrics patterns if adding metrics

## Testing
- [ ] Run unit tests: `make test_unit`
- [ ] If changes affect core functionality, run: `make test_all`
- [ ] For protocol changes, run specific E2E tests:
  - Shannon: `make e2e_test SERVICE_IDS`
  - Morse: `make morse_e2e_test SERVICE_IDS`

## Build Verification
- [ ] Ensure code builds successfully: `make path_build`
- [ ] If making significant changes, test in local environment: `make path_up`

## Final Checks
- [ ] Verify changes don't break existing functionality
- [ ] Check that metrics are properly registered if adding new ones
- [ ] Ensure error handling follows Go conventions
- [ ] Confirm logging uses structured logging with polylog