package math2

import "math"

func MaxInt(ns ...int) int {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] > m {
			m = ns[i]
		}
	}

	return m
}

func MinInt(ns ...int) int {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] < m {
			m = ns[i]
		}
	}

	return m
}

func MaxInt64(ns ...int64) int64 {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] > m {
			m = ns[i]
		}
	}

	return m
}

func MinInt64(ns ...int64) int64 {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] < m {
			m = ns[i]
		}
	}

	return m
}

func MaxFloat64(ns ...float64) float64 {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] > m {
			m = ns[i]
		}
	}

	return m
}

func MinFloat64(ns ...float64) float64 {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] < m {
			m = ns[i]
		}
	}

	return m
}

func B2I(v bool) int {
	if v {
		return 1
	}

	return 0
}

func RoundInt(x float64) int {
	return int(math.Floor(x + 0.5))
}

func CeilMode(n, m int) int {
	v := n / m

	if v*m < n {
		return v + 1
	}

	return v
}

func IIfInt(b bool, n, m int) int {
	if b {
		return n
	}

	return m
}

func IIfInt64(b bool, n, m int64) int64 {
	if b {
		return n
	}

	return m
}

func IIfFloat(b bool, n, m float64) float64 {
	if b {
		return n
	}

	return m
}

func Range(start int, end int) []int {
	nums := make([]int, end-start+1)

	for n := start; n <= end; n++ {
		nums[n-start] = n
	}

	return nums
}

func InArray(value int, list []int) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}

	return false
}

func ToInt32List(list []int) []int32 {
	ns := make([]int32, 0, len(list))

	for _, v := range list {
		ns = append(ns, int32(v))
	}

	return ns
}
