COMMON_MOCKERY_PARAMS=--with-expecter --exported

.PHONY: generate-mocks
generate-mocks:	
	mockery ${COMMON_MOCKERY_PARAMS}

.PHONY: test-e2e-fork9-validium
test-e2e-fork9-validium: stop
	./run-e2e.sh fork9 cdk-validium
	bats bats/

.PHONY: test-e2e-fork11-rollup
test-e2e-fork11-rollup: stop
	./run-e2e.sh fork11 rollup
	bats bats/

.PHONY: test-e2e-fork12-validium
test-e2e-fork12-validium: stop
	./run-e2e.sh fork12 cdk-validium
	bats bats/

.PHONY: test-e2e-fork12-rollup
test-e2e-fork12-rollup: stop
	./run-e2e.sh fork12 rollup
	bats bats/

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
