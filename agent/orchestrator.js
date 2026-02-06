#!/usr/bin/env node
/**
 * M2Sim Multi-Agent Orchestrator (Standalone)
 * Runs agents in sequence using Claude CLI directly.
 * No OpenClaw dependency.
 */

import { spawn, execSync } from 'child_process';
import { existsSync, mkdirSync, appendFileSync } from 'fs';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const REPO_DIR = join(__dirname, '..');
const SKILL_PATH = join(__dirname, 'skills');
const LOGS_DIR = join(__dirname, 'logs');
const TRACKER_ISSUE = 45;
const INTERVAL_MS = 180_000; // 3 minutes
const AGENT_TIMEOUT_MS = 900_000; // 15 minutes
const MODEL = 'claude-opus-4-5';

// Track currently running agent process
let currentAgentProcess = null;
let currentAgentName = null;

function log(message) {
  const timestamp = new Date().toISOString().replace('T', ' ').slice(0, 19);
  console.log(`[${timestamp}] ${message}`);
}

function exec(cmd, options = {}) {
  try {
    return execSync(cmd, { 
      cwd: REPO_DIR, 
      encoding: 'utf-8',
      stdio: ['pipe', 'pipe', 'pipe'],
      ...options 
    }).trim();
  } catch (e) {
    return e.stdout?.trim() || '';
  }
}

function getActionCount() {
  const body = exec(`gh issue view ${TRACKER_ISSUE} --json body -q '.body'`);
  const match = body.match(/Action Count:\s*(\d+)/);
  return match ? parseInt(match[1], 10) : 0;
}

function getNextAgent() {
  const labels = exec(`gh issue view ${TRACKER_ISSUE} --json labels -q '.labels[].name'`);
  const nextMatch = labels.split('\n').find(l => l.startsWith('next:'));
  return nextMatch ? nextMatch.replace('next:', '') : 'alice';
}

function getActiveAgentLabel() {
  const labels = exec(`gh issue view ${TRACKER_ISSUE} --json labels -q '.labels[].name'`);
  const activeMatch = labels.split('\n').find(l => l.startsWith('active:'));
  return activeMatch ? activeMatch.replace('active:', '') : null;
}

function isLocalAgentRunning() {
  // Check if our tracked process is still running
  if (currentAgentProcess && !currentAgentProcess.killed) {
    try {
      // Check if process is still alive (signal 0 doesn't kill, just checks)
      process.kill(currentAgentProcess.pid, 0);
      return true;
    } catch (e) {
      // Process doesn't exist
      currentAgentProcess = null;
      currentAgentName = null;
      return false;
    }
  }
  return false;
}

function clearStaleLabel(agent) {
  log(`Clearing stale active label for: ${agent}`);
  exec(`gh issue edit ${TRACKER_ISSUE} --remove-label "active:${agent}"`);
}

function shouldRunGrace() {
  const count = getActionCount();
  return count > 0 && count % 10 === 0;
}

async function runAgent(agent) {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
  const logFile = join(LOGS_DIR, `${agent}-${timestamp}.log`);
  
  log(`Running agent: ${agent}`);
  
  const prompt = `You are [${agent}] working on the M2Sim project.

**Config:**
- GitHub Repo: sarchlab/m2sim  
- Local Path: ${REPO_DIR}
- Tracker Issue: #${TRACKER_ISSUE}

**Instructions:**
1. First, read the shared rules from: ${join(SKILL_PATH, 'everyone.md')}
2. Then read your specific role from: ${join(SKILL_PATH, `${agent}.md`)}
3. Execute your full cycle as described in your role file.
4. At START of your work: remove label next:${agent}, add label active:${agent}, add next:{your-next-agent}
5. At END of your work: remove label active:${agent}
6. All GitHub activity (commits, PRs, comments) must start with [${agent}]

Work autonomously. Complete your cycle, then exit.`;

  return new Promise((resolve) => {
    const proc = spawn('claude', [
      '--model', MODEL,
      '--dangerously-skip-permissions',
      prompt
    ], {
      cwd: REPO_DIR,
      stdio: ['ignore', 'pipe', 'pipe']
    });

    // Track this process
    currentAgentProcess = proc;
    currentAgentName = agent;

    const timeout = setTimeout(() => {
      log(`Agent ${agent} timed out, killing...`);
      proc.kill('SIGTERM');
    }, AGENT_TIMEOUT_MS);

    proc.stdout.on('data', (data) => {
      const text = data.toString();
      process.stdout.write(text);
      appendFileSync(logFile, text);
    });

    proc.stderr.on('data', (data) => {
      const text = data.toString();
      process.stderr.write(text);
      appendFileSync(logFile, text);
    });

    proc.on('close', (code) => {
      clearTimeout(timeout);
      currentAgentProcess = null;
      currentAgentName = null;
      log(`Agent ${agent} finished with code ${code}`);
      resolve(code);
    });

    proc.on('error', (err) => {
      clearTimeout(timeout);
      currentAgentProcess = null;
      currentAgentName = null;
      log(`Agent ${agent} error: ${err.message}`);
      resolve(1);
    });
  });
}

async function cycle() {
  log('--- Checking cycle ---');
  
  // Pull latest
  exec('git pull --rebase --quiet');
  
  // Primary check: is our local Claude process running?
  if (isLocalAgentRunning()) {
    log(`Local agent '${currentAgentName}' (pid ${currentAgentProcess.pid}) still running, waiting...`);
    return;
  }
  
  // Secondary check: is there an active label on GitHub?
  const activeLabel = getActiveAgentLabel();
  if (activeLabel) {
    // Label exists but no local process - it's stale
    log(`Stale active label detected for '${activeLabel}', clearing...`);
    clearStaleLabel(activeLabel);
  }
  
  // No active agent, run next one
  let nextAgent = getNextAgent();
  
  // Check if Grace should run (every 10th cycle)
  if (shouldRunGrace() && nextAgent === 'alice') {
    log('Grace cycle (action count divisible by 10)');
    await runAgent('grace');
  } else {
    await runAgent(nextAgent);
  }
}

async function main() {
  log('M2Sim Orchestrator started (Node.js, standalone)');
  log(`Interval: ${INTERVAL_MS / 1000}s, Repo: ${REPO_DIR}, Model: ${MODEL}`);
  
  // Create logs directory
  if (!existsSync(LOGS_DIR)) {
    mkdirSync(LOGS_DIR, { recursive: true });
  }
  
  // Initial cycle
  await cycle();
  
  // Schedule recurring cycles
  setInterval(async () => {
    await cycle();
  }, INTERVAL_MS);
}

// Handle graceful shutdown
process.on('SIGINT', () => {
  log('Shutting down...');
  if (currentAgentProcess) {
    log(`Killing agent ${currentAgentName}...`);
    currentAgentProcess.kill('SIGTERM');
  }
  process.exit(0);
});

process.on('SIGTERM', () => {
  log('Shutting down...');
  if (currentAgentProcess) {
    log(`Killing agent ${currentAgentName}...`);
    currentAgentProcess.kill('SIGTERM');
  }
  process.exit(0);
});

main().catch(console.error);
