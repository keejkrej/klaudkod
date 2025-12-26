# Test Protocol: Glob Tool

Tests the `glob` file pattern matching tool.

## Prerequisites

- Run `00-setup.md` first

## Test: Find Python Files

### Prompt
```
Find all Python files in this project
```

### Expected Tool Calls
1. `glob` - with `pattern: "**/*.py"`

### Expected Result
- Returns `main.py` (and any other .py files)

### Verification
- Output contains `main.py`

## Test: Recursive Pattern Matching

### Prompt
```
Find all .toml files in this project
```

### Expected Tool Calls
1. `glob` - with `pattern: "**/*.toml"`

### Expected Result
- Returns `pyproject.toml`

## Test: List All Files

### Prompt
```
List all files in this directory
```

### Expected Tool Calls
1. `glob` - with `pattern: "*"`

### Verification
- Lists main.py, pyproject.toml, .python-version

## Checklist

- [ ] Basic glob pattern works
- [ ] Recursive `**` matching works
- [ ] Directory-specific search works
- [ ] Results are sorted alphabetically
