# Test Protocol: Read Tool

Tests the `read` file tool functionality via WebSocket.

**Implementation:** `backend/internal/tools/read_file.go`

## Prerequisites

1. Read `00-setup.md` first to understand the WebSocket protocol
2. Test environment exists at `tests/.test_env/`
3. Backend server running at `ws://localhost:8080/ws`

---

## Test 1: Basic File Read

### Description
Verify the tool can read a file and return contents with line numbers.

### Command
```bash
echo '{"type":"prompt","content":"Read the file main.py"}' | websocat ws://localhost:8080/ws
```

### Expected WebSocket Messages

1. **tool_call** - LLM calls read tool:
```json
{
  "type": "tool_call",
  "tool_call": {
    "id": "<any>",
    "name": "read",
    "arguments": "{\"filePath\":\"main.py\"}"
  }
}
```

2. **tool_result** - File content returned:
```json
{
  "type": "tool_result",
  "tool_result": {
    "content": "<file>\n00001| def main():\n00002|     print(\"Hello from sample-app!\")\n...",
    "is_error": false
  }
}
```

3. **chunk** (one or more) - LLM response text
4. **done** - Complete

### Pass Criteria
- [ ] Received `tool_call` with `name: "read"`
- [ ] `tool_call.arguments` contains `filePath` set to `main.py`
- [ ] Received `tool_result` with `is_error: false`
- [ ] `tool_result.content` contains `def main():`
- [ ] `tool_result.content` contains `print("Hello from sample-app!")`
- [ ] Received `done` message

### How to Verify
1. Run the websocat command
2. Parse each JSON line received
3. Find message with `type: "tool_call"` - check `name` and `arguments`
4. Find message with `type: "tool_result"` - check `is_error` and `content`
5. Confirm `done` message received

---

## Test 2: Read with Offset/Limit

### Description
Verify the LLM can request specific lines from a file.

### Command
```bash
echo '{"type":"prompt","content":"Read lines 2 to 4 of main.py"}' | websocat ws://localhost:8080/ws
```

### Expected WebSocket Messages

1. **tool_call** with offset/limit parameters:
```json
{
  "type": "tool_call",
  "tool_call": {
    "name": "read",
    "arguments": "{\"filePath\":\"main.py\",\"offset\":2,\"limit\":3}"
  }
}
```

2. **tool_result** starting from line 2:
```json
{
  "type": "tool_result",
  "tool_result": {
    "content": "<file>\n00002|     print(\"Hello from sample-app!\")\n00003| ...",
    "is_error": false
  }
}
```

### Pass Criteria
- [ ] `tool_call.arguments` contains `offset` parameter
- [ ] `tool_result.content` starts with line 2 (shows `00002|`)
- [ ] Line numbers match actual file positions

---

## Test 3: Block .env Files

### Description
Verify that `.env` files cannot be read (security check).

### Setup
```bash
echo "SECRET_KEY=test123" > /Users/jack/workspace/klaudkod/tests/.test_env/.env
```

### Command
```bash
echo '{"type":"prompt","content":"Read the .env file"}' | websocat ws://localhost:8080/ws
```

### Expected WebSocket Messages

1. **tool_call** - LLM attempts to read .env:
```json
{
  "type": "tool_call",
  "tool_call": {
    "name": "read",
    "arguments": "{\"filePath\":\".env\"}"
  }
}
```

2. **tool_result** - Access denied:
```json
{
  "type": "tool_result",
  "tool_result": {
    "content": "...",
    "is_error": true
  }
}
```
The content or error should mention "access denied" or similar.

### Cleanup
```bash
rm /Users/jack/workspace/klaudkod/tests/.test_env/.env
```

### Pass Criteria
- [ ] `tool_result.is_error` is `true`
- [ ] Response mentions "access denied" or blocks the read
- [ ] Secret content `SECRET_KEY=test123` is NOT in any response

---

## Test 4: Block .env Variants

### Description
Verify all .env file patterns are blocked.

### Setup
```bash
echo "SECRET=1" > /Users/jack/workspace/klaudkod/tests/.test_env/.env.local
echo "SECRET=2" > /Users/jack/workspace/klaudkod/tests/.test_env/.env.production
echo "SECRET=3" > /Users/jack/workspace/klaudkod/tests/.test_env/production.env
```

### Commands (test each separately)
```bash
echo '{"type":"prompt","content":"Read .env.local"}' | websocat ws://localhost:8080/ws
echo '{"type":"prompt","content":"Read .env.production"}' | websocat ws://localhost:8080/ws
echo '{"type":"prompt","content":"Read production.env"}' | websocat ws://localhost:8080/ws
```

### Expected Result
All three should return `tool_result` with `is_error: true`

### Cleanup
```bash
rm /Users/jack/workspace/klaudkod/tests/.test_env/.env.local
rm /Users/jack/workspace/klaudkod/tests/.test_env/.env.production
rm /Users/jack/workspace/klaudkod/tests/.test_env/production.env
```

### Pass Criteria
- [ ] `.env.local` read blocked
- [ ] `.env.production` read blocked
- [ ] `production.env` read blocked

---

## Test 5: Allow .env.example Files

### Description
Verify example/template env files ARE readable (whitelisted).

### Setup
```bash
echo "# Example config" > /Users/jack/workspace/klaudkod/tests/.test_env/.env.example
echo "API_KEY=your-key-here" >> /Users/jack/workspace/klaudkod/tests/.test_env/.env.example
```

### Command
```bash
echo '{"type":"prompt","content":"Read .env.example"}' | websocat ws://localhost:8080/ws
```

### Expected WebSocket Messages

1. **tool_result** - File readable:
```json
{
  "type": "tool_result",
  "tool_result": {
    "content": "...# Example config...API_KEY=your-key-here...",
    "is_error": false
  }
}
```

### Cleanup
```bash
rm /Users/jack/workspace/klaudkod/tests/.test_env/.env.example
```

### Pass Criteria
- [ ] `tool_result.is_error` is `false`
- [ ] `tool_result.content` contains the example config

---

## Test 6: Block Binary Files

### Description
Verify binary files cannot be read.

### Setup
```bash
printf '\x00\x01\x02\x03' > /Users/jack/workspace/klaudkod/tests/.test_env/test.bin
```

### Command
```bash
echo '{"type":"prompt","content":"Read test.bin"}' | websocat ws://localhost:8080/ws
```

### Expected WebSocket Messages

**tool_result** with error:
```json
{
  "type": "tool_result",
  "tool_result": {
    "is_error": true
  }
}
```

### Cleanup
```bash
rm /Users/jack/workspace/klaudkod/tests/.test_env/test.bin
```

### Pass Criteria
- [ ] `tool_result.is_error` is `true`
- [ ] Response mentions "binary file"

---

## Test 7: File Not Found

### Description
Verify appropriate error for non-existent files.

### Command
```bash
echo '{"type":"prompt","content":"Read nonexistent_file.txt"}' | websocat ws://localhost:8080/ws
```

### Expected WebSocket Messages

**tool_result** with error:
```json
{
  "type": "tool_result",
  "tool_result": {
    "is_error": true
  }
}
```

### Pass Criteria
- [ ] `tool_result.is_error` is `true`
- [ ] Response mentions "not found" or "does not exist"

---

## Test 8: Path Traversal Blocked

### Description
Verify paths outside working directory are blocked.

### Command
```bash
echo '{"type":"prompt","content":"Read the file ../../../etc/passwd"}' | websocat ws://localhost:8080/ws
```

### Expected WebSocket Messages

**tool_result** with error:
```json
{
  "type": "tool_result",
  "tool_result": {
    "is_error": true
  }
}
```

### Pass Criteria
- [ ] `tool_result.is_error` is `true`
- [ ] Response mentions "outside working directory" or similar
- [ ] System file content NOT exposed

---

## Summary Checklist

| # | Test | Status |
|---|------|--------|
| 1 | Basic file read | [ ] |
| 2 | Offset/limit pagination | [ ] |
| 3 | Block .env files | [ ] |
| 4 | Block .env variants | [ ] |
| 5 | Allow .env.example | [ ] |
| 6 | Block binary files | [ ] |
| 7 | File not found error | [ ] |
| 8 | Path traversal blocked | [ ] |

---

## Report Template

```markdown
# Read Tool Test Report

**Date:** YYYY-MM-DD
**Backend Version:** (git commit hash)
**Environment:** tests/.test_env/

## Results

| Test | Status | Notes |
|------|--------|-------|
| Basic File Read | PASS/FAIL | |
| Offset/Limit | PASS/FAIL | |
| Block .env | PASS/FAIL | |
| Block .env Variants | PASS/FAIL | |
| Allow .env.example | PASS/FAIL | |
| Block Binary Files | PASS/FAIL | |
| File Not Found | PASS/FAIL | |
| Path Traversal | PASS/FAIL | |

**Overall:** X/8 tests passed

## Raw WebSocket Transcripts

### Test 1: Basic File Read
```
> {"type":"prompt","content":"Read the file main.py"}
< {"type":"tool_call",...}
< {"type":"tool_result",...}
< {"type":"chunk",...}
< {"type":"done"}
```

(Include transcripts for failed tests)

## Issues Found
- (List any failures with actual vs expected)

## Recommendations
- (List any suggested fixes)
```
