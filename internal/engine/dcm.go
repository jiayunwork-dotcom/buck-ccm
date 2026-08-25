package engine

import (
	"errors"
	"fmt"
	"math"

	"buck-ccm/internal/spec"
)

const (
	DCMConvergenceTol = 1e-10
	DCMMaxIterations  = 120
)

func dcmResidual(k, d, m float64) float64 {
	return k*m*m + d*d*m - d*d
}

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

func SolveDCMRatioDefault(s spec.Spec) (float64, error) {
	return SolveDCMRatio(s, DCMMaxIterations, DCMConvergenceTol)
}
