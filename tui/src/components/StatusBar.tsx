import React from 'react';
import { Box, Text } from 'ink';

interface StatusBarProps {
  connected: boolean;
}

export function StatusBar({ connected }: StatusBarProps) {
  return (
    <Box paddingX={1} justifyContent="space-between">
      <Text color="gray">Klaudkod v0.1.0</Text>
      <Text color={connected ? 'green' : 'red'}>
        {connected ? '● Connected' : '○ Disconnected'}
      </Text>
    </Box>
  );
}
