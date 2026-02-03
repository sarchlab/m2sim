# M2Sim

A cycle-accurate Apple M2 CPU simulator built on the [Akita](https://github.com/sarchlab/akita) simulation framework.

## Overview

M2Sim simulates ARM64 user-space programs with high timing accuracy, targeting <2% average error compared to real Apple M2 hardware. The simulator separates functional emulation from timing simulation, enabling accurate performance predictions for benchmarks in the μs to ms range.

## Features

### Current Status (M3 - Timing Model in Progress)

- ✅ ARM64 instruction decoder
- ✅ Register file (X0-X30, SP, PC)
- ✅ Basic ALU instructions (ADD, SUB, AND, OR, XOR)
- ✅ Load/Store instructions (LDR, STR)
- ✅ Branch instructions (B, BL, BR, RET, B.cond)
- ✅ Syscall emulation (exit, write)
- ✅ Simple memory model (flat address space)
- 🚧 Pipeline stages (Fetch, Decode, Execute, Memory, Writeback)
- 🚧 Instruction timing model

### Roadmap

- M4: Cache hierarchy (L1/L2 caches)
- M5: Advanced features (branch prediction, OoO execution)
- M6: Validation against real M2 hardware

## Installation

### Prerequisites

- Go 1.25 or later

### Build

```bash
git clone https://github.com/sarchlab/m2sim.git
cd m2sim
go build
```

## Usage

```bash
./m2sim [options] <program>
```

*Note: CLI is currently under development.*

## Project Structure

```
m2sim/
├── driver/     # Simulation driver
├── emu/        # Functional emulation
├── insts/      # Instruction definitions and decoder
├── loader/     # Program loader
├── timing/     # Timing model
├── benchmarks/ # Test programs
└── samples/    # Example usage
```

## Testing

```bash
go test ./...
```

## Related Projects

- [Akita](https://github.com/sarchlab/akita) - The simulation framework
- [MGPUSim](https://github.com/sarchlab/mgpusim) - GPU simulator using similar architecture

## License

See LICENSE file for details.
