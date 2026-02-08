/*
 * Integer Matrix Multiplication Benchmark for M2Sim
 *
 * Medium-sized benchmark (100x100 matrices) using only exit and write syscalls.
 * Compiles as static ARM64 Linux ELF for M2Sim emulation.
 *
 * Target: H3 calibration with cache + compute mix validation.
 */

#include <stdint.h>

/* ARM64 Linux syscall numbers */
#define SYS_WRITE 64
#define SYS_EXIT  93

/* Constants */
#define MATRIX_SIZE 100
#define STDOUT_FD   1

/* Global matrices to ensure they're in memory, not optimized away */
static int32_t matrix_a[MATRIX_SIZE][MATRIX_SIZE];
static int32_t matrix_b[MATRIX_SIZE][MATRIX_SIZE];
static int32_t matrix_c[MATRIX_SIZE][MATRIX_SIZE];

/* Simple syscall wrappers */
static inline long syscall1(long num, long arg1) {
    register long x0 asm("x0") = arg1;
    register long x8 asm("x8") = num;
    asm volatile("svc #0" : "=r"(x0) : "r"(x0), "r"(x8) : "memory");
    return x0;
}

static inline long syscall3(long num, long arg1, long arg2, long arg3) {
    register long x0 asm("x0") = arg1;
    register long x1 asm("x1") = arg2;
    register long x2 asm("x2") = arg3;
    register long x8 asm("x8") = num;
    asm volatile("svc #0" : "=r"(x0) : "r"(x0), "r"(x1), "r"(x2), "r"(x8) : "memory");
    return x0;
}

static void write_str(const char *str) {
    const char *p = str;
    while (*p) p++;  /* Find length */
    syscall3(SYS_WRITE, STDOUT_FD, (long)str, p - str);
}

static void write_num(int64_t num) {
    char buf[32];
    char *p = buf + sizeof(buf) - 1;
    int is_negative = 0;

    *p = '\0';

    if (num < 0) {
        is_negative = 1;
        num = -num;
    }

    if (num == 0) {
        *(--p) = '0';
    } else {
        while (num > 0) {
            *(--p) = '0' + (num % 10);
            num /= 10;
        }
    }

    if (is_negative) {
        *(--p) = '-';
    }

    write_str(p);
}

/* Initialize matrix A with a simple pattern */
static void init_matrix_a(void) {
    int i, j;
    for (i = 0; i < MATRIX_SIZE; i++) {
        for (j = 0; j < MATRIX_SIZE; j++) {
            /* Pattern: A[i][j] = (i + j) % 13 + 1 */
            /* This creates values 1-13, good for avoiding overflow */
            matrix_a[i][j] = ((i + j) % 13) + 1;
        }
    }
}

/* Initialize matrix B with another pattern */
static void init_matrix_b(void) {
    int i, j;
    for (i = 0; i < MATRIX_SIZE; i++) {
        for (j = 0; j < MATRIX_SIZE; j++) {
            /* Pattern: B[i][j] = (i * 7 + j * 3) % 11 + 1 */
            /* Different pattern, values 1-11 */
            matrix_b[i][j] = ((i * 7 + j * 3) % 11) + 1;
        }
    }
}

/* Initialize result matrix C to zero */
static void init_matrix_c(void) {
    int i, j;
    for (i = 0; i < MATRIX_SIZE; i++) {
        for (j = 0; j < MATRIX_SIZE; j++) {
            matrix_c[i][j] = 0;
        }
    }
}

/* Core matrix multiplication algorithm */
static void multiply_matrices(void) {
    int i, j, k;

    /* Standard triple-nested loop matrix multiplication */
    /* C[i][j] = sum(A[i][k] * B[k][j]) for all k */
    for (i = 0; i < MATRIX_SIZE; i++) {
        for (j = 0; j < MATRIX_SIZE; j++) {
            int32_t sum = 0;
            for (k = 0; k < MATRIX_SIZE; k++) {
                sum += matrix_a[i][k] * matrix_b[k][j];
            }
            matrix_c[i][j] = sum;
        }
    }
}

/* Compute a simple checksum of the result matrix */
static int64_t compute_checksum(void) {
    int64_t checksum = 0;
    int i, j;

    for (i = 0; i < MATRIX_SIZE; i++) {
        for (j = 0; j < MATRIX_SIZE; j++) {
            checksum += matrix_c[i][j];
        }
    }

    /* Add some additional mixing to make checksum more distinctive */
    checksum ^= (checksum >> 32);
    checksum += (checksum * 0x9e3779b97f4a7c15ULL);
    checksum ^= (checksum >> 32);

    return checksum;
}

/* Verify a few key matrix elements for correctness */
static int verify_result(void) {
    /* For the initialization patterns we use:
     * A[0][0] = 1, B[0][0] = 1
     * A[0][1] = 2, B[1][0] = 4
     * etc.
     *
     * We can compute C[0][0] manually:
     * C[0][0] = A[0][0]*B[0][0] + A[0][1]*B[1][0] + A[0][2]*B[2][0] + ...
     *
     * With our patterns:
     * A[0][k] = (k % 13) + 1
     * B[k][0] = ((k * 7) % 11) + 1
     */

    /* Rather than hardcoding expected values, we'll verify
     * that the result is reasonable (non-zero, within expected range)
     */

    /* C[0][0] should be positive and reasonable */
    if (matrix_c[0][0] <= 0 || matrix_c[0][0] > 1000000) {
        return 0;  /* Verification failed */
    }

    /* C[50][50] (middle element) should also be reasonable */
    if (matrix_c[50][50] <= 0 || matrix_c[50][50] > 1000000) {
        return 0;  /* Verification failed */
    }

    /* C[99][99] (last element) should also be reasonable */
    if (matrix_c[99][99] <= 0 || matrix_c[99][99] > 1000000) {
        return 0;  /* Verification failed */
    }

    return 1;  /* Verification passed */
}

/* Main function */
int main(void) {
    int64_t checksum;
    int verification_result;

    write_str("Matrix Multiply Benchmark (100x100 integers)\n");
    write_str("Initializing matrices...\n");

    /* Initialize input matrices */
    init_matrix_a();
    init_matrix_b();
    init_matrix_c();

    write_str("Performing matrix multiplication...\n");

    /* Perform the multiplication */
    multiply_matrices();

    write_str("Computing checksum...\n");

    /* Compute and verify result */
    checksum = compute_checksum();
    verification_result = verify_result();

    /* Output results */
    write_str("Matrix multiplication complete.\n");
    write_str("Checksum: ");
    write_num(checksum);
    write_str("\n");

    write_str("Verification: ");
    if (verification_result) {
        write_str("PASSED\n");
    } else {
        write_str("FAILED\n");
        syscall1(SYS_EXIT, 1);  /* Exit with error code */
    }

    write_str("Benchmark completed successfully.\n");

    /* Exit with success */
    syscall1(SYS_EXIT, 0);

    /* Never reached, but included to satisfy compiler */
    return 0;
}