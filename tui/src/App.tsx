import React, { useState, useCallback } from 'react';
import { Box, Text, useInput, useApp } from 'ink';
import { Chat } from './components/Chat.js';
import { Input } from './components/Input.js';
import { StatusBar } from './components/StatusBar.js';
import { useWebSocket } from './hooks/useWebSocket.js';
import { useChat } from './hooks/useChat.js';

export function App() {
  const { exit } = useApp();
  const [inputValue, setInputValue] = useState('');

  const { connected, send } = useWebSocket('ws://localhost:8080/ws');
  const { messages, addMessage, updateLastMessage } = useChat();

  const handleSubmit = useCallback((value: string) => {
    if (!value.trim()) return;

    addMessage({ role: 'user', content: value });
    send({ type: 'prompt', content: value });
    setInputValue('');
  }, [addMessage, send]);

  useInput((input, key) => {
    if (key.ctrl && input === 'c') {
      exit();
    }
  });

  return (
    <Box flexDirection="column" height="100%">
      <Box flexGrow={1} flexDirection="column" paddingX={1}>
        <Chat messages={messages} />
      </Box>

      <Box borderStyle="single" borderColor="gray" paddingX={1}>
        <Input
          value={inputValue}
          onChange={setInputValue}
          onSubmit={handleSubmit}
          placeholder="Type a message..."
        />
      </Box>

      <StatusBar connected={connected} />
    </Box>
  );
}
