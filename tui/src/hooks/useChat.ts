import { useState, useCallback } from 'react';

export interface Message {
  role: 'user' | 'assistant' | 'system';
  content: string;
}

export function useChat() {
  const [messages, setMessages] = useState<Message[]>([]);

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

  return {
    messages,
    addMessage,
    updateLastMessage,
    clearMessages,
  };
}
