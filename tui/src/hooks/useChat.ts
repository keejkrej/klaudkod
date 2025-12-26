import { useState, useCallback } from 'react';

export interface ToolCall {
  id: string;
  name: string;
  arguments: string;
  status: 'pending' | 'executing' | 'completed' | 'error';
}

export interface ToolResult {
  toolCallId: string;
  content: string;
  isError: boolean;
}

export interface Message {
  role: 'user' | 'assistant' | 'system';
  content: string;
  toolCalls?: ToolCall[];
  toolResult?: ToolResult;
}

export function useChat() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [activeToolCalls, setActiveToolCalls] = useState<ToolCall[]>([]);

  const addMessage = useCallback((message: Message) => {
    setMessages((prev) => [...prev, message]);
  }, []);

  const updateLastMessage = useCallback((content: string) => {
    setMessages((prev) => {
      if (prev.length === 0) return prev;
      const updated = [...prev];
      updated[updated.length - 1] = {
        ...updated[updated.length - 1],
        content: updated[updated.length - 1].content + content,
      };
      return updated;
    });
  }, []);

  const clearMessages = useCallback(() => {
    setMessages([]);
  }, []);

  const addToolCall = useCallback((toolCall: ToolCall) => {
    setActiveToolCalls((prev) => [...prev, toolCall]);
  }, []);

  const updateToolCallStatus = useCallback((id: string, status: ToolCall['status']) => {
    setActiveToolCalls((prev) =>
      prev.map((toolCall) =>
        toolCall.id === id ? { ...toolCall, status } : toolCall
      )
    );
  }, []);

  const addToolResult = useCallback((result: ToolResult) => {
    updateToolCallStatus(result.toolCallId, 'completed');
    setMessages((prev) => {
      const lastMessage = prev[prev.length - 1];
      if (lastMessage && lastMessage.toolCalls?.some(tc => tc.id === result.toolCallId)) {
        const updated = [...prev];
        updated.push({
          role: 'assistant',
          content: '',
          toolResult: result,
        });
        return updated;
      }
      return prev;
    });
  }, [updateToolCallStatus]);

  const clearToolCalls = useCallback(() => {
    setActiveToolCalls([]);
  }, []);

  return {
    messages,
    activeToolCalls,
    addMessage,
    updateLastMessage,
    clearMessages,
    addToolCall,
    updateToolCallStatus,
    addToolResult,
    clearToolCalls,
  };
}