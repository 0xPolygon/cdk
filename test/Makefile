DOCKERCOMPOSE := docker compose -f docker-compose.yml

.PHONY: start-l1-network
start-l1-network: ## Starts mock L1 network
	$(DOCKERCOMPOSE) up -d cdk-l1

.PHONY: start-node
start-node: ## Starts db, validium node and validium RPC
	$(DOCKERCOMPOSE) up -d cdk-postgres
	$(DOCKERCOMPOSE) up -d cdk-validium-node-1
	$(DOCKERCOMPOSE) up -d cdk-validium-hermez-rpc-1

.PHONY: start-prover
start-prover: ## Starts prover
	$(DOCKERCOMPOSE) up -d cdk-prover-1

.PHONY: start-aggregator-direct
start-aggregator-direct: ## Starts aggregator with direct settlement
	$(DOCKERCOMPOSE) up -d cdk-validium-hermez-aggregator-direct-1

.PHONY: start-aggregator-agglayer
start-aggregator-agglayer: ## Starts aggregator with agglayer settlement
	$(DOCKERCOMPOSE) up -d cdk-validium-hermez-aggregator-agglayer-1

.PHONY: start-all-components
start-all-components: ## Starts mock l1 network, aggregator with direct settlement, prover and 
	$(MAKE) start-l1-network
	$(MAKE) start-aggregator-direct
	$(MAKE) start-prover
	$(MAKE) start-node

.PHONY: run-tests
run-tests: ## Runs all tests within the workspace
	go test -v -race ./...

.PHONY: help
help: ## Prints the help
		@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'