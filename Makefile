run/app:
	@go run ./app/cmd --socket /home/nate/.osquery/shell.em

test:
	go test -v -race ./app/cmd/...
