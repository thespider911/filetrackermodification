run/app:
	@echo 'Running app...'
	@go run ./app/cmd --socket /home/nate/.osquery/shell.em

build/api:
	@echo 'Building app...'
	@go build -ldflags=${linker_flags} -o=./bin/FileModificationTracker ./app/cmd
#	@GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/FileModificationTracker ./app/cmd
	@CC=x86_64-w64-mingw32-gcc \
    CXX=x86_64-w64-mingw32-g++ \
    GOOS=windows \
    GOARCH=amd64 \
    CGO_ENABLED=1 \
    go build -ldflags=${linker_flags} -o=./bin/windows_amd64/FileModificationTracker.exe ./app/cmd

test:
	@echo 'Running tests...'
	@go test -v -race ./app/cmd/...

vendor:
	@echo 'Tidying and verifying module dependencies...'
	@go mod tidy
	@go mod verify
	@echo 'Vendoring dependencies...'
	@go mod vendor