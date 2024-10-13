# GOOK

This project is a flexible webhook processing server written in Go. It receives webhook payloads, processes them based on configured conditions, and forwards modified payloads to specified endpoints.

## Features

- Configurable webhook endpoints
- Conditional processing of incoming payloads
- Template-based payload modification
- Concurrent processing of multiple outputs
- Dockerized for easy deployment
- Graceful shutdown support

## Prerequisites

- Go 1.23 or later
- Docker (optional, for containerized deployment)

## Configuration

The server is configured using a `config.json` file. Here's an example structure:

```json
{
  "webhooks": [
    {
      "name": "Example Webhook",
      "path": "/webhook1",
      "outputs": [
        {
          "url": "https://example.com/endpoint",
          "condition": "{{ .someField }} == 'someValue'",
          "template": {
            "newField": "Value: {{ .someField }}"
          }
        }
      ]
    }
  ]
}
```

## Setup and Running

### Local Development

1. Clone the repository:
   ```
   git clone https://github.com/frxncodominguez/gook.git
   cd gook
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Create your `config.json` file in the project root.

4. Run the server:
   ```
   go run main.go
   ```

The server will start on port 8080.

### Docker Deployment

1. Build the Docker image:
   ```
   docker build -t gook .
   ```

2. Run the container:
   ```
   docker run -p 8080:8080 -v $(pwd)/config.json:/root/config.json gook
   ```

## Usage

Send HTTP POST requests to the configured webhook paths with JSON payloads. The server will process the payload according to the configuration and forward it to the specified outputs.

Example:
```
curl -X POST http://localhost:8080/webhook1 -H "Content-Type: application/json" -d '{"someField": "someValue"}'
```

## Graceful Shutdown

The server supports graceful shutdown. When receiving a SIGINT or SIGTERM signal, it will:
1. Stop accepting new connections
2. Complete processing of ongoing requests
3. Shut down after a timeout or when all requests are completed

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.