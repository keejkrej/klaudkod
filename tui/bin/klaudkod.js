#!/usr/bin/env node

import { spawn } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { existsSync, readFileSync } from 'fs';
import { config } from 'dotenv';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Get paths - handle both dev and installed scenarios
const projectRoot = dirname(dirname(__dirname));
const backendPath = join(projectRoot, 'backend', 'klaudkod');
const tuiDistPath = join(__dirname, '..', 'dist', 'index.js');

// Check if files exist
if (!existsSync(backendPath)) {
  console.error(`Error: Backend binary not found at ${backendPath}`);
  process.exit(1);
}

if (!existsSync(tuiDistPath)) {
  console.error(`Error: TUI not built. Run 'npm run build' in ${dirname(tuiDistPath)}`);
  process.exit(1);
}

// Load .env file
const envPath = join(projectRoot, '.env');
if (existsSync(envPath)) {
  config({ path: envPath });
}

// Start backend process
console.log('Starting Klaudkod backend...');
const backendProcess = spawn(backendPath, {
  stdio: 'inherit',
  cwd: projectRoot,
});

backendProcess.on('error', (err) => {
  console.error('Failed to start backend:', err.message);
  process.exit(1);
});

// Give backend time to start
await new Promise(resolve => setTimeout(resolve, 1000));

// Start TUI process
console.log('Starting Klaudkod TUI...');
const tuiProcess = spawn('node', [tuiDistPath], {
  stdio: 'inherit',
  cwd: dirname(tuiDistPath),
});

// Handle cleanup
let isCleaningUp = false;
const cleanup = () => {
  if (isCleaningUp) return;
  isCleaningUp = true;

  if (!tuiProcess.killed) {
    tuiProcess.kill();
  }
  if (!backendProcess.killed) {
    backendProcess.kill();
  }
};

process.on('SIGINT', cleanup);
process.on('SIGTERM', cleanup);

tuiProcess.on('exit', cleanup);
