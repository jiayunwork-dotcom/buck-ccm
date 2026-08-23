package engine

import (
	"errors"
	"fmt"
	"math"

	"buck-ccm/internal/spec"
)

// DCM 电压比求解器的收敛契约：二分法在 (D, 1) 内找
// f(M) = K·M² + D²·M − D² = 0 的根，误差低于 DCMConvergenceTol 或
// 区间宽度低于 tol 即收敛，最多迭代 DCMMaxIterations 次。
const (
	// DCMConvergenceTol 是默认残差容差。
	DCMConvergenceTol = 1e-10
	// DCMMaxIterations 是默认最大迭代次数。
	DCMMaxIterations = 120
)

// dcmResidual 计算 DCM 电压比方程在 M 处的残差。
//
// 推导：稳态时平均电感电流等于负载电流，联立伏秒平衡得到
// K·M² + D²·M − D² = 0。M = Vout/Vin 在 (D, 1) 内取到。
func dcmResidual(k, d, m float64) float64 {
	return k*m*m + d*d*m - d*d
}

// SolveDCMRatio 用二分法求解 DCM 电压比 M = Vout/Vin。
//
// 求解契约：
//
//	前置条件：s 已通过 spec.Validate（D∈(0,1)、K>0），且调用方确认 DCM。
//	根区间：  [D, 1]（f(D) = D²(K+D−1) ≤ 0，f(1) = K > 0）
//	收敛：    |f(mid)| ≤ tol 或区间宽度 ≤ tol
//	上限：    maxIter 次迭代内必须收敛，否则返回错误
//
// 该契约被测试盯住：迭代上限、容差与根区间一起决定了 DCM 输出的
// 确定性与精度。
func SolveDCMRatio(s spec.Spec, maxIter int, tol float64) (float64, error) {
	if err := spec.Validate(&s); err != nil {
		return 0, err
	}
	if maxIter <= 0 {
		return 0, errors.New("DCM 求解迭代上限必须为正")
	}
	if tol <= 0 || math.IsNaN(tol) {
		return 0, errors.New("DCM 求解容差必须为正且有限")
	}
	k := ParameterK(s)
	d := s.D
	lo, hi := d, 1.0
	flo := dcmResidual(k, d, lo)
	if flo > 0 {
		return 0, errors.New("DCM 求解根区间错误：f(D) 非负，参数不满足 DCM 前提")
	}
	fhi := dcmResidual(k, d, hi)
	if fhi <= 0 {
		return 0, errors.New("DCM 求解根区间错误：f(1) 非正，参数不满足 DCM 前提")
	}
	for i := 0; i < maxIter; i++ {
		mid := (lo + hi) / 2
		fm := dcmResidual(k, d, mid)
		if math.Abs(fm) <= tol || (hi-lo) <= tol {
			return mid, nil
		}
		if flo*fm < 0 {
			hi = mid
		} else {
			lo = mid
			flo = fm
		}
	}
	return 0, fmt.Errorf("%w：%d 次迭代内未达到容差 %g", errNoConvergence, maxIter, tol)
}

// SolveDCMRatioDefault 用默认容差与迭代上限求解 DCM 电压比。
func SolveDCMRatioDefault(s spec.Spec) (float64, error) {
	return SolveDCMRatio(s, DCMMaxIterations, DCMConvergenceTol)
}
