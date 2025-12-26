import { useState, useEffect, useCallback, useRef } from 'react';
import WebSocket from 'ws';

interface WSMessage {
  type: string;
  [key: string]: unknown;
}

export function useWebSocket(url: string) {
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const onMessageRef = useRef<((data: WSMessage) => void) | null>(null);

  useEffect(() => {
    const connect = () => {
      try {
        const ws = new WebSocket(url);

        ws.on('open', () => {
          setConnected(true);
        });

        ws.on('message', (data) => {
          try {
            const parsed = JSON.parse(data.toString()) as WSMessage;
            onMessageRef.current?.(parsed);
          } catch {
            // Ignore parse errors
          }
        });

        ws.on('close', () => {
          setConnected(false);
          // Reconnect after 2 seconds
          setTimeout(connect, 2000);
        });

        ws.on('error', () => {
          setConnected(false);
        });

        wsRef.current = ws;
      } catch {
        setConnected(false);
        setTimeout(connect, 2000);
      }
    };

    connect();

    return () => {
      wsRef.current?.close();
    };
  }, [url]);

  const send = useCallback((message: WSMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    }
  }, []);

  const onMessage = useCallback((handler: (data: WSMessage) => void) => {
    onMessageRef.current = handler;
  }, []);

  return {
    connected,
    send,
    onMessage,
  };
}
