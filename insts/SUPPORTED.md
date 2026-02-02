# ARM64 Instruction Support Status

This document tracks which ARM64 instructions are implemented.

## Legend
- ✅ Implemented and tested
- 🚧 In progress
- ❌ Not implemented

## Data Processing - Immediate
| Instruction | Status | Notes |
|-------------|--------|-------|
| ADD (imm)   | ❌     |       |
| SUB (imm)   | ❌     |       |
| AND (imm)   | ❌     |       |
| ORR (imm)   | ❌     |       |
| EOR (imm)   | ❌     |       |
| MOV (imm)   | ❌     |       |

## Data Processing - Register
| Instruction | Status | Notes |
|-------------|--------|-------|
| ADD (reg)   | ❌     |       |
| SUB (reg)   | ❌     |       |
| AND (reg)   | ❌     |       |
| ORR (reg)   | ❌     |       |
| EOR (reg)   | ❌     |       |

## Load/Store
| Instruction | Status | Notes |
|-------------|--------|-------|
| LDR         | ❌     |       |
| STR         | ❌     |       |
| LDP         | ❌     |       |
| STP         | ❌     |       |

## Branch
| Instruction | Status | Notes |
|-------------|--------|-------|
| B           | ❌     |       |
| BL          | ❌     |       |
| BR          | ❌     |       |
| BLR         | ❌     |       |
| RET         | ❌     |       |
| B.cond      | ❌     |       |

## System
| Instruction | Status | Notes |
|-------------|--------|-------|
| SVC         | ❌     | Syscall |
| NOP         | ❌     |       |
