# Test Environment Setup Protocol

This protocol sets up the test environment and explains how to run tool tests.

## Architecture Overview

```
┌─────────────────────┐     WebSocket      ┌─────────────────┐      HTTP       ┌─────────────────┐
│                     │   ws://host/ws     │                 │    API calls    │                 │
│  Test Agent         │◄──────────────────►│     Backend     │◄───────────────►│   LLM Provider  │
│  (Claude Code or    │                    │   (Go server)   │                 │ (OpenAI, etc.)  │
│   human w/ websocat)│                    │                 │                 │                 │
└─────────────────────┘                    └────────┬────────┘                 └─────────────────┘
                                                    │
                                                    │ Executes
                                                    ▼
                                           ┌─────────────────┐
                                           │     Tools       │
                                           │  read, write,   │
                                           │  glob, grep,    │
                                           │     bash        │
                                           └─────────────────┘
```

## How It Works

1. **Test agent connects** to backend via WebSocket
2. **Agent sends prompt** in natural language (e.g., "Read main.py")
3. **Backend forwards** prompt to LLM with available tool definitions
4. **LLM decides** which tool to call (e.g., `read` with `filePath: "main.py"`)
5. **Backend executes** the tool
6. **Backend streams** results back to agent via WebSocket
7. **Agent verifies** the tool call and result match expectations

## WebSocket Protocol

### Endpoint

```
ws://localhost:8080/ws
```

### Messages: Agent → Backend

| Type | Purpose | Example |
|------|---------|---------|
| `prompt` | Send natural language request | `{"type":"prompt","content":"Read main.py"}` |
| `cancel` | Cancel current operation | `{"type":"cancel"}` |

### Messages: Backend → Agent

| Type | Purpose | Example |
|------|---------|---------|
| `tool_call` | LLM decided to call a tool | See below |
| `tool_result` | Result of tool execution | See below |
| `chunk` | Streamed text from LLM | `{"type":"chunk","content":"Here is..."}` |
| `done` | Response complete | `{"type":"done"}` |
| `error` | Error occurred | `{"type":"error","error":"Something failed"}` |

#### tool_call message format

```json
{
  "type": "tool_call",
  "tool_call": {
    "id": "call_abc123",
    "name": "read",
    "arguments": "{\"filePath\":\"main.py\"}"
  }
}
```

#### tool_result message format

```json
{
  "type": "tool_result",
  "tool_result": {
    "content": "<file>\n00001| def main():\n00002|     print(\"Hello!\")\n</file>",
    "is_error": false
  }
}
```

## Connecting to WebSocket

### Install websocat

```bash
# macOS
brew install websocat

# Linux
cargo install websocat
```

### Interactive Session

```bash
# Connect to backend
websocat ws://localhost:8080/ws

# Type a prompt (press Enter to send)
{"type":"prompt","content":"Read the file main.py"}

# Observe responses (JSON messages will appear)
# Press Ctrl+C to disconnect
```

### One-liner Command

```bash
echo '{"type":"prompt","content":"Read main.py"}' | websocat ws://localhost:8080/ws
```

## Example Session Transcript

```
$ websocat ws://localhost:8080/ws
> {"type":"prompt","content":"Read main.py"}

< {"type":"tool_call","tool_call":{"id":"call_1","name":"read","arguments":"{\"filePath\":\"main.py\"}"}}
< {"type":"tool_result","tool_result":{"content":"<file>\n00001| def main():\n00002|     print(\"Hello from sample-app!\")\n00003| \n00004| \n00005| if __name__ == \"__main__\":\n00006|     main()\n00007| \n\n(End of file - total 7 lines)\n</file>","is_error":false}}
< {"type":"chunk","content":"Here"}
< {"type":"chunk","content":" is the content"}
< {"type":"chunk","content":" of main.py:\n\n```python"}
< {"type":"chunk","content":"\ndef main():"}
...
< {"type":"done"}
```

## Agent Instructions

**When asked to run a test protocol:**

1. **Read this file first** to understand the WebSocket protocol
2. Check if test environment exists at `tests/.test_env/`
3. If not, run the setup steps below
4. Ensure backend is running
5. Read the requested test protocol file (01-xxx.md, etc.)
6. For each test case:
   - Run the websocat command specified in the test
   - Capture all WebSocket messages received
   - Parse each JSON message
   - Verify against the expected messages and pass criteria
   - Record PASS or FAIL
7. Generate a detailed report

## Prerequisites

- Go 1.21+ installed
- `uv` package manager installed
- `websocat` installed
- LLM API key configured in backend `.env`

## Setup Steps

### 1. Build the Backend

```bash
cd /Users/jack/workspace/klaudkod/backend
go build -o klaudkod ./cmd/klaudkod
```

### 2. Create Test Environment

```bash
cd /Users/jack/workspace/klaudkod
rm -rf tests/.test_env
mkdir -p tests/.test_env
cd tests/.test_env
uv init --name sample_app
```

Creates:
```
tests/.test_env/
├── .python-version
├── README.md
├── main.py          # def main(): print("Hello from sample-app!")
└── pyproject.toml
```

### 3. Start Backend Server

**Important:** Start from test environment directory so tools use correct working dir.

```bash
cd /Users/jack/workspace/klaudkod/tests/.test_env
../../backend/klaudkod &
sleep 2
```

### 4. Verify Server is Running

```bash
# Health check
curl -s http://localhost:8080/health
# Expected output: OK

# Test WebSocket (should see connection then timeout)
echo '{"type":"prompt","content":"hello"}' | timeout 10 websocat ws://localhost:8080/ws
```

## Teardown

```bash
# Stop the server
pkill -f klaudkod

# Clean up test environment
rm -rf /Users/jack/workspace/klaudkod/tests/.test_env
```

## Notes

- `tests/.test_env/` is gitignored
- Backend working directory = where server is started
- Tool implementations: `backend/internal/tools/`
- Test protocols numbered: 00, 01, 02...
- All tests are integration tests - LLM decides tool calls
