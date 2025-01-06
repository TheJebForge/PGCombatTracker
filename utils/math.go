package utils

import "math"

// AbsInt "if it's simple to write, it shouldn't be in stdlib" my ass!
func AbsInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func LerpF32(a, b, t float32) float32 {
	return a + (b-a)*t
}

func LerpInt(a, b int, t float64) int {
	floatingA, floatingB := float64(a), float64(b)
	floatingResult := floatingA + (floatingB-floatingA)*t
	return int(math.Round(floatingResult))
}

type Interpolatable interface {
	Interpolate(other Interpolatable, t float64) Interpolatable
}
