.PHONY: bench_crypto
bench_crypto: ## Benchmark Shannon SDK signing performance with ethereum_secp256k1 optimization
	@echo "🔐 Running Shannon SDK signing benchmark..."
	@CGO_ENABLED=1 go test ./e2e -bench=BenchmarkShannonSigningDirect -run=^$$ -benchtime=10s -tags="bench,ethereum_secp256k1" -benchmem

.PHONY: bench_crypto_raw
bench_crypto_raw: ## Show raw benchmark output for debugging
	@echo "🔬 Raw Benchmark Output..."
	@echo "=== Decred Backend ==="
	@go test ./e2e -bench=BenchmarkShannonSigningDirect -run=^$$ -benchtime=2s -tags="bench" -benchmem || echo "Decred benchmark failed"
	@echo ""
	@echo "=== Ethereum Backend ==="
	@CGO_ENABLED=1 go test ./e2e -bench=BenchmarkShannonSigningDirect -run=^$$ -benchtime=2s -tags="bench,ethereum_secp256k1" -benchmem || echo "Ethereum benchmark failed"

.PHONY: bench_crypto_compare
bench_crypto_compare: ## Compare Shannon SDK signing performance (table only)
	@bash ./e2e/scripts/compare_crypto_performance.sh quiet 10s

.PHONY: bench_crypto_verbose
bench_crypto_verbose: ## Compare Shannon SDK signing performance (verbose output)
	@BENCHMEM=true bash ./e2e/scripts/compare_crypto_performance.sh verbose 10s
