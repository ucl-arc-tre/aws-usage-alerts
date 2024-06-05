package efs

// Value if the pointer is not nil, otherwise the value
func valueOrZero(value *int64) float64 {
	if value == nil {
		return 0
	} else {
		return float64(*value)
	}
}
