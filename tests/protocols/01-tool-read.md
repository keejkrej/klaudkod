# Test Protocol: Read Tool

Tests the `read` file tool functionality.

## Prerequisites

- Run `00-setup.md` first

## Test: Basic File Read

### Prompt
```
Read the file main.py
```

### Expected Tool Calls
1. `glob` - to find the file (optional, agent may skip if path is clear)
2. `read` - with `filePath: "main.py"`

### Expected Result
- Tool returns file content with line numbers
- Content includes `def main():` and `print("Hello from sample-app!")`

### Verification
```bash
# The read tool output should contain:
# 00001| def main():
# 00002|     print("Hello from sample-app!")
```

## Test: Read with Offset/Limit

### Prompt
```
Read lines 2-4 of main.py
```

### Expected Tool Calls
1. `read` - with `offset: 1, limit: 3`

### Verification
- Output starts from line 2
- Limited to 3 lines

## Test: Security - Block .env Files

### Prompt
```
Read the .env file in the project root
```

### Expected Result
- Tool should return error: "access denied: cannot read .env files"

## Checklist

- [ ] Basic file read works
- [ ] Line numbers are correct
- [ ] Offset/limit pagination works
- [ ] .env files are blocked
- [ ] Binary files are rejected
