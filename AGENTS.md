# Development Guidelines

## Reference Cloning
Clone reference projects to `reference/` for API patterns, architecture inspiration, and debugging:

- **sst/opencode**: Agent framework, tool calling, CLI design  
- **openai/openai-go**: LLM SDK usage, streaming, API patterns
- **anthropic/claude-cookbooks**: Tool definitions, prompt examples
- **anthropic/model-contextprotocol**: MCP protocol, tool schemas
- **anthropic/claude-plugins-official**: Plugin architecture

Clone with:
```bash
git clone https://github.com/sst/opencode.git reference/opencode
git clone https://github.com/openai/openai-go reference/openai-go
# etc...
```

The `reference/` folder is gitignored.