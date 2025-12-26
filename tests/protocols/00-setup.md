# Test Environment Setup Protocol

This protocol sets up the test environment before running any tool tests.

## Prerequisites

- Backend must be built: `cd backend && go build -o klaudkod ./cmd/klaudkod`
- `.env` file configured with LLM settings

## Setup Steps

### 1. Clean and Create Test Environment

```bash
cd /Users/jack/workspace/klaudkod
rm -rf tests/.test_env
mkdir -p tests/.test_env
```

### 2. Create Sample Python Project

```bash
cd tests/.test_env
uv init --name sample_app
```

This creates a minimal Python project with:
- `main.py` - hello world script
- `pyproject.toml` - project config
- `.python-version` - Python version

### 3. Start Backend Server

```bash
cd /Users/jack/workspace/klaudkod
./backend/klaudkod &
sleep 2
curl -s http://localhost:8080/health  # Should return "OK"
```

## Teardown

```bash
pkill -f klaudkod
rm -rf tests/.test_env
```

## Verification Checklist

- [ ] `tests/.test_env/sample_app/` exists
- [ ] Backend responds to health check at http://localhost:8080/health
- [ ] `uv run python main.py` outputs "Hello from sample-app!"

## Notes

- The test environment `tests/.test_env/` is gitignored
- Each test run starts fresh with `uv init`
- Backend working directory should be set to `tests/.test_env/sample_app/`
