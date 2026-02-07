.PHONY: test mocks clean help

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

mocks: ## Generate all mocks
	@echo "🔧 Generating mocks..."
	@go generate ./...
	@echo "✅ Mocks generated successfully!"

test: mocks ## Generate mocks and run all tests
	@echo ""
	@echo "🧪 Running tests..."
	@start_time=$$(date +%s); \
	go test -v ./...; \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	echo ""; \
	echo "✨ All tests completed successfully!"; \
	echo "⏱️  Total time: $${duration}s"

test-short: mocks ## Generate mocks and run tests without verbose output
	@go test ./...

test-coverage: mocks ## Generate mocks and run tests with coverage report
	@echo "🧪 Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

clean: ## Clean generated files
	@echo "🧹 Cleaning up..."
	@rm -f coverage.out coverage.html
	@echo "✅ Cleanup complete!"

