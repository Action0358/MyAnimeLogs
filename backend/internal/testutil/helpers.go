package testutil

// StringPtr converts a string to *string
// Useful for testing optional string fields
func StringPtr(s string) *string {
	return &s
}

// IntPtr converts an int to *int
// Useful for testing optional int fields
func IntPtr(i int) *int {
	return &i
}

// BoolPtr converts a bool to *bool
// Useful for testing optional bool fields
func BoolPtr(b bool) *bool {
	return &b
}
