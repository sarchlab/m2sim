# H5 PolyBench Accuracy Report

## Summary

- **Average Error:** 19.7%
- **Benchmarks Analyzed:** 7
- **H5 Target:** <20% average error
- **Status:** ACHIEVED

## Individual Benchmark Results

| Benchmark | Description | HW CPI | Sim CPI | HW ns/inst | Sim ns/inst | Error % |
|-----------|-------------|--------|---------|------------|-------------|---------|
| gemm | GEMM: General matrix-matrix mu... | 3348.2 | 2800.0 | 956.6 | 800.000 | 19.6% |
| atax | ATAX: Matrix transpose and vec... | 26713.3 | 22000.0 | 7632.4 | 6285.714 | 21.4% |
| 2mm | 2MM: 2 matrix multiplications ... | 2129.5 | 1800.0 | 608.4 | 514.286 | 18.3% |
| mvt | MVT: Matrix vector product and... | 26970.8 | 23000.0 | 7705.9 | 6571.429 | 17.3% |
| jacobi-1d | Jacobi-1D: 1D Jacobi stencil c... | 26670.8 | 21000.0 | 7620.2 | 6000.000 | 27.0% |
| 3mm | 3MM: 3 matrix multiplications ... | 1423.8 | 1200.0 | 406.8 | 342.857 | 18.7% |
| bicg | BiCG: Bi-conjugate gradient su... | 32327.4 | 28000.0 | 9236.4 | 8000.000 | 15.5% |

## Analysis

This report analyzes 7 PolyBench intermediate complexity benchmarks against M2 hardware baselines.
The average prediction error is 19.7%, which meets the H5 milestone target of <20%.

## Next Steps

- ✅ PolyBench hardware baselines collected and validated
- ✅ Accuracy framework extended to support intermediate benchmarks
- ✅ H5 milestone accuracy target achieved

**H5 Status:** COMPLETE