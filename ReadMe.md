# File Modification Tracker

## Overview
This project implements a File Modification Tracker in Go, designed to monitor and record modifications to files in a specified directory. The application runs as a background service, integrates system monitoring via osquery, and provides configuration management.

## Features
- Monitors specified directory for file modifications
- Runs as a background service with two independent threads:
    - Worker Thread: Maintains a queue of shell commands and executes them
    - Timer Thread: Periodically retrieves file modification stats using osquery
- Exposes HTTP endpoints for health check and log retrieval
- Configurable via YAML file
- Logging mechanism for debugging and monitoring
- API integration for sending collected data to a remote endpoint

## Requirements
- Go (Golang)
- osquery
- viper (for configuration management)
- validator (for configuration validation)

## Project Structure
```
app/
├── cmd/
│   ├── errors.go
│   ├── handler.go
│   ├── routes.go
│   ├── server.go
│   ├── response.go
│   └── task.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── helpers/
│   │   └── helpers.go
│   ├── testutil/
│   │   └── setup.go
│   └── service/
│       ├── service.go
│       ├── filetracker.go
│       └── command/
│           └── commands.go
└── main.go
config.yaml
install.wxs
Makefile
test_api.go
```

## Configuration
The application is configured using a `config.yaml` file:

```yaml
directory: "{{.HomeDir}}/Desktop/test_tracker"
check_interval: 5
queue_size: 100
http_port: 4000
api_port: 4041
api_endpoint: "http://localhost:4041/file-endpoint"
```

## Building and Running
To setup the go project, run
```
make vendor
```
this will install all the module dependencies required to run the application in gomod

### Running the Application
To run the application:

```
make run/app
```

This command runs the application with the osquery socket: 
You need to install `osquery` first in your system.

```
go run ./app/cmd --socket \\.\pipe\shell.em
```
update the .osquery path in your machine. The result is logged in a logger file and can be accessed by running the api


`
NOTE::When running the program, make sure you have a test_data folder on your route project and you can add as many files here before running the program. 
This folder with files, using `app/cmd/testutil/setup.go` will create a folder on your os desktop named `test_tracker` which will be used to read the files from, for this test program.`

### Building the Application
To run the application:

```
make build/api
```

This command will create two executables, on for go and an .exe for windows

```
go build -ldflags=${linker_flags} -o=./bin/FileModificationTracker ./app/cmd

CC=x86_64-w64-mingw32-gcc \
CXX=x86_64-w64-mingw32-g++ \
GOOS=windows \
GOARCH=amd64 \
CGO_ENABLED=1 \
go build -ldflags=${linker_flags} -o=./bin/windows_amd64/FileModificationTracker.exe ./app/cmd
```

### Running Tests
To run tests:
```
make test
```

This command runs the tests with race detection:

```
go test -v -race ./app/cmd/...
```

## Window Installer
To install the program:

After building the application, a folder `bin` will be created with the `.exe` for windows in `bin/windows_amd64/FileModificationTracker.exe`

To create installer for this, make sure you have `wix` version 5 installed then run this command on the project root where we have this `installer.wxs` file.
```
wix build FileModificationTracker.wxs
```

This will create the installation files and you can see `FileModificationTracker.msi` on the project root.
```
FileModificationTracker.msi
```
(You can delete the one I have generated and create another one)

Double click on this  `FileModificationTracker.msi` and a `FileModificationTracker` will be installed on your machine.  
To confirm this you can check on the system Apps installed where you can also click to `unistall` the program.

Once the program is created, go to: ` Program Files(x86) `

and look for `FileModificationTracker`. Open the folder and before you double click to run the program, copy the `test_data` folder in project root (for testing purpose).

This will create the `test_tracker` folder in your desktop where the program will look for files modified.
also copy the `config.yaml` to the same FileModificationTracker program folder.

Your Program Files(x86)/FileModificationTracker should have.

```
NOTE:: MAKE SURE THE USER HAS ALL PERMISSION TO THIS INSTALLATION FOLDER.
```

````
test_data

config.yaml

FileModificationTracker
````

Now double click the `FileModificationTracker` to run the program.




## API Endpoints
- Health Check: `/health` this is to check the application if is running ok
- Logs Retrieval: `/logs` this will log the data in logs
- Command Query: `/help` this will show all the commands you need to run available for this app
- Command Execution: `/execute` execute requires command and path as described in help
- Start Service: `/start` start will start the service
- Stop Service: `/stop` stop will stop the service

## Test API
A separate test API is provided to simulate the remote endpoint for receiving file modification data. To run the test API:

```
go run test_api.go
```

The test API provides two endpoints:
- `/file-endpoint`: Receives file update data (POST)
- `/view-data`: Displays all received data (GET)

## UI Component
The application uses the Fyne library to create a simple native UI dialog box for starting/stopping the service and viewing logs.

## Limitations on Features
- Windows depending on your set up might require you to download some programs for this to learn especially fynne setup.
- Can always implement more comprehensive error handling and recovery mechanisms and more unit tests to cover edge cases and improve code coverage.

## Contributing
Please submit pull requests or open issues to discuss proposed changes.

