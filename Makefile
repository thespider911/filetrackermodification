run/app:
	@go run ./app/cmd --socket /home/nate/.osquery/shell.em

test:
	@go test -v -race ./app/cmd/...

vendor:
	@go mod tidy
	@go mod verify
	@go mod vendor