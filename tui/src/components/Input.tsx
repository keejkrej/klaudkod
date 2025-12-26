import React from 'react';
import { Box, Text } from 'ink';
import TextInput from 'ink-text-input';

interface InputProps {
  value: string;
  onChange: (value: string) => void;
  onSubmit: (value: string) => void;
  placeholder?: string;
}

export function Input({ value, onChange, onSubmit, placeholder }: InputProps) {
  return (
    <Box>
      <Text color="cyan" bold>{'> '}</Text>
      <TextInput
        value={value}
        onChange={onChange}
        onSubmit={onSubmit}
        placeholder={placeholder}
      />
    </Box>
  );
}
