run-consumer-papa:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=transaction_stream
.PHONY: run-consumer

run-consumer-lsm:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=loan_logs
.PHONY: run-consumer

run-consumer-pas:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=acuan_transaction_notif
.PHONY: run-consumer

run-consumer-pas-pretty:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=acuan_transaction_notif | ./pretitfy-log.sh
.PHONY: run-consumer-pas-pretty

run-consumer-bre:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=billing_repayment_logs
.PHONY: run-consumer

run-consumer-transformer-dlq:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=transformer_stream_dlq
.PHONY: run-consumer

test:
	CGO_ENABLED=1 go test -race -short -count=1 ./... -gcflags=all=-l

test-cover:
	CGO_ENABLED=1 go test -cover -race -short -count=1 -coverprofile=coverage.out ./... -gcflags=all=-l
	go tool cover -html=coverage.out

generate:
	go generate ./...

run-kafka:
	docker stack deploy -c ./deployments/kafka-ui.yml amartha-dev --with-registry-auth

stop-kafka:
	docker stack rm amartha-dev

# Run API Server
api:
	@echo "🚀 Starting API Server..."
	@go run cmd/api/main.go