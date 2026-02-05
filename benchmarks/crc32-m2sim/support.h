/* M2Sim support header for Embench benchmarks */
#ifndef SUPPORT_H
#define SUPPORT_H

#define CPU_MHZ 1

/* Benchmark interface */
void initialise_benchmark(void);
int benchmark(void);
int verify_benchmark(int result);

/* Time tracking (unused in M2Sim) */
#define start_trigger() ((void)0)
#define stop_trigger() ((void)0)

/* Random number generator from beebsc.c */
static long int beebs_seed = 0;

static inline int rand_beebs(void) {
    beebs_seed = (beebs_seed * 1103515245L + 12345) & ((1UL << 31) - 1);
    return (int)(beebs_seed >> 16);
}

static inline void srand_beebs(unsigned int new_seed) {
    beebs_seed = (long int)new_seed;
}

#endif /* SUPPORT_H */
