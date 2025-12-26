import React, { useState, useCallback, useEffect } from 'react';
import { Box, Text, useInput, useApp } from 'ink';
import { Chat } from './components/Chat.js';
import { Input } from './components/Input.js';
import { StatusBar } from './components/StatusBar.js';
import { useWebSocket } from './hooks/useWebSocket.js';
import { useChat, ToolResult } from './hooks/useChat.js';

export function App() {
  const { exit } = useApp();
  const [inputValue, setInputValue] = useState('');
  const [toolResults, setToolResults] = useState<Map<string, ToolResult>>(new Map());

  const { connected, send, onMessage } = useWebSocket('ws://localhost:8080/ws');
  const { 
    messages, 
    activeToolCalls,
    addMessage, 
    updateLastMessage,
    addToolCall,
    updateToolCallStatus,
    addToolResult,
    clearToolCalls
  } = useChat();

  const handleSubmit = useCallback((value: string) => {
    if (!value.trim()) return;

    addMessage({ role: 'user', content: value });
    send({ type: 'prompt', content: value });
    setInputValue('');
    setToolResults(new Map());
  }, [addMessage, send]);

  useEffect(() => {
    if (!onMessage) return;

    onMessage((data: any) => {
      switch (data.type) {
        case 'chunk':
          if (data.isFirst) {
            addMessage({ role: 'assistant', content: '' });
          }
          updateLastMessage(data.content);
          break;

        case 'tool_call':
          const toolCall = data.tool_call;
          addToolCall({
            id: toolCall.id,
            name: toolCall.name,
            arguments: toolCall.arguments,
            status: 'executing'
          });
          break;

        case 'tool_result':
          const toolResultData = data.tool_result;
          const toolResult: ToolResult = {
            toolCallId: toolResultData.toolCallId,
            content: toolResultData.content,
            isError: toolResultData.isError || false
          };
          setToolResults(prev => new Map(prev).set(toolResultData.toolCallId, toolResult));
          addToolResult(toolResult);
          break;

        case 'done':
          clearToolCalls();
          break;

        case 'error':
          addMessage({ role: 'system', content: `Error: ${data.message}` });
          break;
      }
    });
  }, [onMessage, addMessage, updateLastMessage, addToolCall, addToolResult, clearToolCalls]);

  useInput((input, key) => {
    if (key.ctrl && input === 'c') {
      exit();
    }
  });

  return (
    <Box flexDirection="column" height="100%">
      <Box flexGrow={1} flexDirection="column" paddingX={1}>
        <Chat 
          messages={messages} 
          activeToolCalls={activeToolCalls}
          toolResults={toolResults}
        />
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