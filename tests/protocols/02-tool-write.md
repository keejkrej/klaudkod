# Test Protocol: Write Tool

Tests the `write` file tool functionality.

## Prerequisites

- Run `00-setup.md` first

## Test: Create New File

### Prompt
```
Create a new file called utils.py with a function called greet(name) that returns "Hello, {name}!"
```

### Expected Tool Calls
1. `write` - with `filePath` and `content`

### Expected Result
- File `utils.py` created
- Contains the greet function

### Verification
```bash
uv run python -c "from utils import greet; print(greet('World'))"
# Should output: Hello, World!
```

## Test: Modify Existing File

### Prompt
```
Add an add(a, b) function to main.py that returns the sum of two numbers. Put it before the main() function.
```

### Expected Tool Calls
1. `read` - to get current content (agent should read first)
2. `write` - with updated content

### Verification
```bash
uv run python -c "from main import add; print(add(2, 3))"
# Should output: 5
```

## Test: Security - Path Traversal Prevention

### Prompt
```
Write a file to /tmp/test.txt with content "hello"
```

### Expected Result
- Should fail with "access denied: path is outside working directory"

## Checklist

- [ ] New file creation works
- [ ] File modification works
- [ ] Python syntax is valid after write
- [ ] Path traversal is blocked
