## Summary

< ONE_LINE_SUMMARY>

### Primary Changes:

- < Change 1 >
- < Change 2 >

### Secondary Changes:

- < Change 1 >
- < Change 2 >

You can use the following as a helper for an LLM of your choice:

```bash
git --no-pager diff main  -- ':!*.pb.go' ':!*.pulsar.go' ':!*.json' ':!*.yaml' ':!*.yml' ':!*.gif' ':!*.md' | diff2html -s side --format json -i stdin -o stdout | pbcopy
```

## Issue

- Issue or PR: #{ISSUE_OR_PR_NUMBER}

## Type of change

Select one or more from the following:

- [ ] New feature, functionality or library
- [ ] Bug fix
- [ ] Code health or cleanup
- [ ] Documentation
- [ ] Other (specify)

## QoS Checklist

### E2E Validation & Tests

- [ ] `make path_up`
- [ ] `make test_e2e_evm_shannon`
- [ ] `make test_e2e_evm_morse`

### Observability

- [ ] 1. `make path_up`
- [ ] 2. Run one of the following:
  - For `Shannon` with `anvil`: `make test_request__shannon_relay_util_100`
  - For `Morse` with `F00C`: `make test_request__morse_relay_util_100`
- [ ] 3. Visit [PATH Relay Grafana Dashboard](http://localhost:3003/d/relays/path-service-requests) to view results

## Sanity Checklist

- [ ] I have updated the GitHub Issue `assignees`, `reviewers`, `labels`, `project`, `iteration` and `milestone`
- [ ] For docs, I have run `make docusaurus_start`
- [ ] For code, I have run `make test_all`
- [ ] For configurations, I have updated the documentation
- [ ] I added `TODO`s where applicable
