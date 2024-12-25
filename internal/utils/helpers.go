package utils

import "strconv"

// IntPtr returns a pointer to the given int value
func IntPtr(i int) *int {
	return &i
}

// StringToUint converts a string to uint
func StringToUint(s string) uint {
	i, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint(i)
}

// StringToInt converts a string to int with a default value of 0 if conversion fails
func StringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
