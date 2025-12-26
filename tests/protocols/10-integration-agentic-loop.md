# Test Protocol: Agentic Loop Integration

Tests the full agentic loop with multiple tool calls.

## Prerequisites

- Run `00-setup.md` first

## Test: Multi-Step Task

### Prompt
```
Read main.py, then add a multiply(a, b) function before main(), and verify it works by running a test.
```

### Expected Tool Calls (in order)
1. `glob` or direct path - find main.py
2. `read` - read current content
3. `write` - add the multiply function
4. `bash` - run Python to test the function

### Expected Behavior
- Agent reads the file first (doesn't blindly write)
- Agent adds function in correct location
- Agent verifies the change works

### Verification
```bash
uv run python -c "from main import multiply; print(multiply(3, 4))"
# Should output: 12
```

## Test: Error Recovery

### Prompt
```
Read the file nonexistent.py
```

### Expected Behavior
- `read` tool returns error "file not found"
- Agent acknowledges the error gracefully
- Agent may suggest alternatives or ask for clarification

## Test: Conversation Context

### Multi-turn Conversation
1. First prompt: "What files are in sample_project?"
2. Second prompt: "Read the main Python file"
3. Third prompt: "Add a docstring to the main function"

### Expected Behavior
- Agent remembers context from previous messages
- Uses information from previous tool calls
- Doesn't re-read files unnecessarily

## Checklist

- [ ] Multi-tool task completes successfully
- [ ] Tools are called in logical order
- [ ] Agent reads before writing
- [ ] Agent verifies changes when appropriate
- [ ] Error messages are handled gracefully
- [ ] Conversation context is maintained
