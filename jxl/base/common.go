package base

type Numeric interface {
	~int | ~int32 | ~int64
}

// DivCeil calculates the ceiling of a division using generic types.
func DivCeil[T1 Numeric, T2 Numeric](a T1, b T2) T1 {
	// Convert T2 to T1 to perform the division, assuming that T2 is numeric and can be converted to T1
	var bConverted T1
	switch v := any(b).(type) {
	case int:
		bConverted = T1(v)
	case int32:
		bConverted = T1(v)
	case int64:
		bConverted = T1(v)
	default:
		panic("unsupported type for b")
	}

	// Perform the division and ceiling operation
	result := (a + bConverted - 1) / bConverted
	return result
}
