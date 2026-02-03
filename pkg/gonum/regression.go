package gonum

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

// LinearRegression простая линейная регрессия y = a*x + b
type LinearRegression struct {
	a float64 // наклон
	b float64 // смещение
}

// Fit обучает модель на данных: days — дни от начала, progress — выработка (0..1)
func (lr *LinearRegression) Fit(days []float64, progress []float64) error {
	n := len(days)
	if n != len(progress) || n < 2 {
		return nil // недостаточно данных
	}

	// Матрица признаков [1, x]
	X := mat.NewDense(n, 2, nil)
	for i := 0; i < n; i++ {
		X.Set(i, 0, 1.0)
		X.Set(i, 1, days[i])
	}

	// Вектор целевых значений
	y := mat.NewVecDense(n, progress)

	// (X^T * X)^-1 * X^T * y
	XT := mat.DenseCopyOf(X.T())
	XTX := mat.NewDense(2, 2, nil)
	XTX.Mul(XT, X)

	var XTXInv mat.Dense
	if err := XTXInv.Inverse(XTX); err != nil {
		return err
	}

	XTy := mat.NewVecDense(2, nil)
	XTy.MulVec(XT, y)

	coeffs := mat.NewVecDense(2, nil)
	coeffs.MulVec(&XTXInv, XTy)

	lr.b = coeffs.AtVec(0)
	lr.a = coeffs.AtVec(1)
	return nil
}

// Predict прогнозирует выработку на день x
func (lr *LinearRegression) Predict(day float64) float64 {
	return lr.a*day + lr.b
}

// DaysToCompletion оценивает дни до завершения (при 100% выработке)
func (lr *LinearRegression) DaysToCompletion() float64 {
	if lr.a <= 0 {
		return math.Inf(1) // никогда не завершится
	}
	return (1.0 - lr.b) / lr.a
}
