.PHONY: bench_signing bench_signing_compare bench_signing_raw bench_signing_verbose

bench_signing: ## Benchmark Shannon SDK signing performance with ethereum_secp256k1 optimization
	@echo "🔐 Running Shannon SDK signing benchmark..."
	@CGO_ENABLED=1 go test ./e2e -bench=BenchmarkShannonSigningDirect -run=^$$ -benchtime=10s -tags="bench,ethereum_secp256k1" -benchmem

bench_signing_raw: ## Show raw benchmark output for debugging
	@echo "🔬 Raw Benchmark Output..."
	@echo "=== Decred Backend ==="
	@go test ./e2e -bench=BenchmarkShannonSigningDirect -run=^$$ -benchtime=2s -tags="bench" -benchmem || echo "Decred benchmark failed"
	@echo ""
	@echo "=== Ethereum Backend ==="
	@CGO_ENABLED=1 go test ./e2e -bench=BenchmarkShannonSigningDirect -run=^$$ -benchtime=2s -tags="bench,ethereum_secp256k1" -benchmem || echo "Ethereum benchmark failed"

bench_signing_compare: ## Compare Shannon SDK signing performance (table only)
	@bash ./e2e/scripts/compare_signing_performance.sh quiet 10s

bench_signing_verbose: ## Compare Shannon SDK signing performance (verbose output)
	@BENCHMEM=true bash ./e2e/scripts/compare_signing_performance.sh verbose 10s
