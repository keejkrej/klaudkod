# Test Protocol: Grep Tool

Tests the `grep` content search tool.

## Prerequisites

- Run `00-setup.md` first

## Test: Search for Function Definition

### Prompt
```
Search for "def main" in this project
```

### Expected Tool Calls
1. `grep` - with `pattern: "def main"`

### Expected Result
- Finds match in `main.py`
- Shows line number and content

### Verification
- Output shows `main.py` with the matching line

## Test: Search with File Pattern

### Prompt
```
Search for "print" in all Python files
```

### Expected Tool Calls
1. `grep` - with `pattern: "print"` and `glob: "*.py"`

### Expected Result
- Finds `print("Hello from sample-app!")` in main.py

## Test: Case Insensitive Search

### Prompt
```
Search for "HELLO" case-insensitively
```

### Expected Tool Calls
1. `grep` - with `pattern: "HELLO"` and `ignoreCase: true`

### Expected Result
- Finds the hello string despite case difference

## Checklist

- [ ] Basic pattern search works
- [ ] File pattern filtering works
- [ ] Case insensitive search works
- [ ] Line numbers are shown
- [ ] Context lines work (if implemented)
