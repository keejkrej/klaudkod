import React from 'react';
import { Box, Text } from 'ink';
import type { ToolCall, ToolResult } from '../hooks/useChat.js';

interface ToolCallDisplayProps {
  toolCall: ToolCall;
  result?: ToolResult;
}

export function ToolCallDisplay({ toolCall, result }: ToolCallDisplayProps) {
  const getStatusIndicator = () => {
    switch (toolCall.status) {
      case 'pending':
        return <Text color="yellow">[ ]</Text>;
      case 'executing':
        return <Text color="blue">[*]</Text>;
      case 'completed':
        return <Text color="green">[+]</Text>;
      case 'error':
        return <Text color="red">[x]</Text>;
      default:
        return null;
    }
  };

  const formatArguments = () => {
    try {
      const parsed = JSON.parse(toolCall.arguments);
      if (typeof parsed === 'object' && parsed !== null) {
        return Object.entries(parsed)
          .map(([key, value]) => `${key}=${value}`)
          .join(' ');
      }
      return toolCall.arguments;
    } catch {
      return toolCall.arguments;
    }
  };

  const displayResult = result ? (
    <Box paddingLeft={2} paddingTop={1}>
      <Box flexDirection="column" paddingX={1}>
        <Text color={result.isError ? 'red' : 'white'}>
          {result.content.length > 500
            ? result.content.substring(0, 500) + '...'
            : result.content}
        </Text>
      </Box>
    </Box>
  ) : null;

  return (
    <Box flexDirection="column">
      <Box flexDirection="row" gap={1}>
        {getStatusIndicator()}
        <Text color="cyan" bold>
          {toolCall.name}
        </Text>
      </Box>
      <Box paddingLeft={4}>
        <Text color="gray">{formatArguments()}</Text>
      </Box>
      {displayResult}
    </Box>
  );
}