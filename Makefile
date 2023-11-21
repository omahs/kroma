VERSION := $(shell git describe --tags --abbrev=0 --match v* 2> /dev/null || echo 'v0.0.0')
GIT_COMMIT := $(shell git rev-parse --short=8 HEAD)

LD_FLAGS_ARGS +=-X main.Version=$(VERSION)
LD_FLAGS_ARGS +=-X main.Meta=$(GIT_COMMIT)
LD_FLAGS := -ldflags "$(LD_FLAGS_ARGS)"

build:
	GO111MODULE=on go build -v $(LD_FLAGS) -o bin/kroma-node ./op-node/cmd/main.go
	GO111MODULE=on go build -v $(LD_FLAGS) -o bin/kroma-stateviz ./op-node/cmd/stateviz/main.go
	GO111MODULE=on go build -v $(LD_FLAGS) -o bin/kroma-batcher ./op-batcher/cmd/main.go
	GO111MODULE=on go build -v $(LD_FLAGS) -o bin/kroma-validator ./components/validator/cmd/main.go
.PHONY: build

clean:
	@rm -rf bin/*
.PHONY: clean

test:
	go test ./op-bindings/...
	go test ./op-batcher/...
	go test ./op-node/...
	go test ./op-service/...
	go test ./op-chain-ops/...
	go test ./components/...
	go test ./utils/...
	go test ./op-e2e/... -timeout 30m # requires a minimum of 30min in a CI
	yarn test
.PHONY: test

lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2 \
	&& golangci-lint run
.PHONY: lint

bindings:
	make -C ./op-bindings
.PHONY: op-bindings

contracts-snapshot:
	@(cd ./packages/contracts && yarn gas-snapshot && yarn storage-snapshot)
.PHONY: gas-snapshot

devnet-up:
	@bash ./ops-devnet/devnet-up.sh
.PHONY: devnet-up

devnet-down:
	@(cd ./ops-devnet && GENESIS_TIMESTAMP=$(shell date +%s) docker compose stop)
.PHONY: devnet-down

devnet-clean:
	rm -rf ./packages/contracts/deployments/devnetL1
	rm -rf ./.devnet
	cd ./ops-devnet && docker compose down
	docker image ls 'ops-devnet*' --format='{{.Repository}}' | xargs -r docker rmi
	docker volume ls --filter name=ops-devnet --format='{{.Name}}' | xargs -r docker volume rm
.PHONY: devnet-clean

update-geth:
	@ETH_GETH=$$(go list -m -f '{{.Path}}@{{.Version}}' github.com/ethereum/go-ethereum); \
	KROMA_GETH=$$(go list -m -f '{{.Path}}@{{.Version}}' github.com/kroma-network/go-ethereum@dev); \
	go mod edit -replace $$ETH_GETH=$$KROMA_GETH
	@go mod tidy
.PHONY: update-geth
