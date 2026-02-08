# Makefile for medium-sized benchmarks
# Builds ARM64 Linux ELF binaries for M2Sim emulation

CC = aarch64-linux-musl-gcc
CFLAGS = -static -O1 -DSPEC -mgeneral-regs-only -fno-builtin
OBJDUMP = aarch64-linux-musl-objdump

TARGETS = matmul

.PHONY: all clean

all: $(TARGETS)

matmul: matmul.c
	$(CC) $(CFLAGS) -o $@ $<
	$(OBJDUMP) -d $@ > $@.dis
	@echo "Built $@ - ARM64 Linux ELF ready for M2Sim"
	@ls -la $@

clean:
	rm -f $(TARGETS) *.dis

.SUFFIXES:
.SUFFIXES: .c