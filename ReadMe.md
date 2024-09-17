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
│   └── service/
│       ├── service.go
│       ├── filetracker.go
│       └── command/
│           └── commands.go
└── main.go
config.yaml
Makefile
test_api.go
```

## Configuration
The application is configured using a `config.yaml` file:

```yaml
directory: "/path/to/directory"
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

This command runs the application with the osquery socket: You need to install osquery first in your system.

```
go run ./app/cmd --socket /home/directory/.osquery/shell.em
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
- I did not manage to implement Windows-specific features and MSI packaging due to not able to secure a windows machine, everyone around me are linux based.
The UI component using Fyne has been tested on Linux but not on Windows.
- Can always implement more comprehensive error handling and recovery mechanisms and more unit tests to cover edge cases and improve code coverage.

## Contributing
Please submit pull requests or open issues to discuss proposed changes.

