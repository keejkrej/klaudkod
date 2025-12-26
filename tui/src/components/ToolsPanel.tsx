import React from 'react';
import { Box, Text } from 'ink';
import type { ToolCall, ToolResult } from '../hooks/useChat.js';
import { ToolCallDisplay } from './ToolCall.js';

interface ToolsPanelProps {
  toolCalls: ToolCall[];
  toolResults: Map<string, ToolResult>;
}

export function ToolsPanel({ toolCalls, toolResults }: ToolsPanelProps) {
  if (toolCalls.length === 0) {
    return null;
  }

  return (
    <Box
      borderStyle="round"
      borderColor="blue"
      paddingX={1}
      marginY={1}
      flexDirection="column"
    >
      <Text color="blue" bold>
        Tool Execution
      </Text>
      {toolCalls.map((toolCall) => (
        <ToolCallDisplay
          key={toolCall.id}
          toolCall={toolCall}
          result={toolResults.get(toolCall.id)}
        />
      ))}
    </Box>
  );
}