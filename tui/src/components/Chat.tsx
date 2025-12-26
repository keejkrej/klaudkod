import React from 'react';
import { Box, Text } from 'ink';
import type { Message } from '../hooks/useChat.js';

interface ChatProps {
  messages: Message[];
}

export function Chat({ messages }: ChatProps) {
  if (messages.length === 0) {
    return (
      <Box flexDirection="column" alignItems="center" justifyContent="center" flexGrow={1}>
        <Text color="gray">Welcome to Klaudkod</Text>
        <Text color="gray" dimColor>Type a message to start chatting</Text>
      </Box>
    );
  }

  return (
    <Box flexDirection="column" gap={1}>
      {messages.map((message, index) => (
        <MessageBubble key={index} message={message} />
      ))}
    </Box>
  );
}

function MessageBubble({ message }: { message: Message }) {
  const isUser = message.role === 'user';

  return (
    <Box flexDirection="column">
      <Text color={isUser ? 'blue' : 'green'} bold>
        {isUser ? 'You' : 'Assistant'}
      </Text>
      <Box paddingLeft={2}>
        <Text wrap="wrap">{message.content}</Text>
      </Box>
    </Box>
  );
}
