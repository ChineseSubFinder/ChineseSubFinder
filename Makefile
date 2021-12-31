# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
# see https://github.com/mongodb/mongocli/blob/master/Makefile

TEST_CMD?=go test
UNIT_TAGS?=unit
INTEGRATION_TAGS?=integration

.PHONY: dltest
dltest: ## Clone test sample files
	@echo "==> Cloning test sample files..."
	git clone https://github.com/wyq977/CSFTestFiles.git TestData

.PHONY: test
test: temp-test

.PHONY: temp-test
temp-test: ## [TEMP] Run simple test scripts
	@scripts/temp-test.sh

# .PHONY: unit-test
# unit-test: ## Run unit-tests
# 	@echo "==> Running unit tests..."
# 	$(TEST_CMD) --tags="$(UNIT_TAGS)" -race -cover -count=1 -coverprofile $(COVERAGE) ./internal...

.PHONY: list
list: ## List all make targets
	@${MAKE} -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | sort

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
