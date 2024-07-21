# tellm

tellm is a local LLM (Large Language Model) logging tool that allows you to record prompts and responses from your LLM interactions. It provides a simple SDK for logging and retrieving entries, along with a local server for viewing logs.

## Setup

### Prerequisites

- Go 1.16 or later

### Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/tellm.git
cd tellm
```

2. Build the server:

```bash
cd cmd/server
go build
```

## Usage

### Running the Server

To start the tellm server:

```bash
./server
```

The server will start on `http://localhost:8080`.

### Using the SDK

To use the tellm SDK in your Go project:

1. Add the SDK to your project:

```bash
go get github.com/yourusername/tellm/sdk
```

2. Import and use the SDK in your code:

```go
package main

import (
	"fmt"
	"log"

	"github.com/yourusername/tellm/sdk"
)

func main() {
	client := sdk.NewClient("http://localhost:8080")

	// Log a prompt and response
	err := client.Log("What is the capital of France?", "The capital of France is Paris.")
	if err != nil {
		log.Fatalf("Failed to log: %v", err)
	}

	// Retrieve all logs
	logs, err := client.GetLogs()
	if err != nil {
		log.Fatalf("Failed to get logs: %v", err)
	}

	// Print the logs
	for _, entry := range logs {
		fmt.Printf("Timestamp: %s\nPrompt: %s\nResponse: %s\n\n", entry.Timestamp, entry.Prompt, entry.Response)
	}
}
```

## Viewing Logs

Open `http://localhost:8080` in your web browser to view the logged entries in a simple HTML interface.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.