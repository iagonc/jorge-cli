# CLI Tool for Managing Resources

This CLI tool allows you to interact with an API for managing resources. It provides commands to create, list, update, and delete resources. The tool is built in Go and leverages powerful libraries like Cobra for command-line interactions and Zap for logging.

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Available Commands](#available-commands)
    - [List Resources](#list-resources)
    - [Create a Resource](#create-a-resource)
    - [Update a Resource](#update-a-resource)
    - [Delete a Resource](#delete-a-resource)
- [Examples](#examples)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)
- [Contact](#contact)

## Features

- **Resource Management**: Create, list, update, and delete resources via API interactions.
- **Command-Line Interface**: User-friendly CLI built with Cobra.
- **Context Handling**: Graceful shutdown and cancellation using Go's context package.
- **Logging**: Structured logging with Zap for easy debugging and monitoring.
- **Configuration Management**: Flexible configuration using Viper, supporting environment variables.

## Prerequisites

- **Go**: Version 1.19 or higher is required. You can download it from [golang.org](https://golang.org/dl/).
- **API Server**: The CLI interacts with an API server. Ensure that the API server is running and accessible.

## Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/iagonc/jorge-cli.git
   cd jorge-cli/cmd/cli
   ```
2. **Build the CLI**

```bash
go build -o cli main.go
```

This will generate an executable named cli.

**Configuration**

The CLI uses environment variables for configuration. You can set these variables in your shell or use a .env file with tools like direnv.

**Available Configuration Options**
- API_BASE_URL: The base URL of the API server (default: http://localhost:8080/api/v1).
- TIMEOUT: Timeout for HTTP requests in seconds (default: 10).
- VERSION: The version of the CLI (default: v1.0.0).

**Setting Configuration via Environment Variables**

Example using export:

```bash
export API_BASE_URL="http://your-api-server.com/api/v1"
export TIMEOUT=15
export VERSION="v1.0.1"
```

**Usage**

The CLI provides several commands to manage resources. Use the --help flag to get more information about each command.

```bash
./cli --help
Available Commands
```

**List Resources**

List all resources from the API.

```bash
./cli list
```
**Create a Resource**

Create a new resource by providing a name and DNS.

```bash
./cli create --name "Resource Name" --dns "resource.dns.example.com"

Flags:

--name, -n: Resource name (required).
--dns, -d: Resource DNS (required).
```

**Update a Resource**

Update an existing resource's name and/or DNS.

```bash
./cli update --id 123 --name "New Name" --dns "new.dns.example.com"
Flags:

--id, -i: Resource ID (required).
--name, -n: New resource name (optional).
--dns, -d: New resource DNS (optional).
```

**Delete a Resource**

Delete a resource by its ID.

```bash
./cli delete --id 123
Flags:

--id, -i: Resource ID (required).
```

Note: You will be prompted for confirmation before deletion.

**Examples**

1. **Creating a Resource**

```bash
./cli create -n "Test Resource" -d "test.resource.example.com"
Output:

yaml

Resource Created:
ID: 101
Name: Test Resource
DNS: test.resource.example.com
```

2. **Listing Resources**

```bash
./cli list
Output:

yaml

ID    Name                 DNS                            CreatedAt           UpdatedAt
----------------------------------------------------------------------------------------------
101   Test Resource        test.resource.example.com      2023-08-01 12:00    2023-08-01 12:00
102   Another Resource     another.resource.example.com   2023-08-02 15:30    2023-08-02 15:30
```

3. **Updating a Resource**

```bash
./cli update -i 101 -n "Updated Resource Name"
Output:

yaml

Resource Updated:
ID: 101
Name: Updated Resource Name
DNS: test.resource.example.com
```

4. **Deleting a Resource**


```bash
./cli delete -i 101
Prompt:

vbnet

Resource Details:
ID: 101
Name: Updated Resource Name
DNS: test.resource.example.com
Are you sure you want to delete this resource? (yes/no):
Type yes to confirm deletion.

Output:

yaml

Resource Deleted:
ID: 101
Name: Updated Resource Name
DNS: test.resource.example.com
```