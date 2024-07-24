.PHONY: test-e2e-elderberry-validium
test-e2e-elderberry-validium: stop ## Runs e2e tests checking elderberry/validium
	./run-e2e-seq_sender.sh cdk-validium

.PHONY: test-e2e-elderberry-rollup
test-e2e-elderberry-rollup: stop ## Runs e2e tests checking elderberry/rollup
	./run-e2e-seq_sender.sh rollup

.PHONY: stop
stop:
	kurtosis clean --all


## Help display.
## Pulls comments from beside commands and prints a nicely formatted
## display with the commands and their usage information.
.DEFAULT_GOAL := help

.PHONY: help
help: ## Prints this help
		@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
