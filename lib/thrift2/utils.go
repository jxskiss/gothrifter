package thrift2

// Helpers that convert from various base types to respective pointer types.

func Float32Ptr(v float32) *float32 { return &v }
func Float64Ptr(v float64) *float64 { return &v }
func IntPtr(v int) *int             { return &v }
func Int8Ptr(v int8) *int8          { return &v }
func Int16Ptr(v int16) *int16       { return &v }
func Int32Ptr(v int32) *int32       { return &v }
func Int64Ptr(v int64) *int64       { return &v }
func StringPtr(v string) *string    { return &v }
func Uint32Ptr(v uint32) *uint32    { return &v }
func Uint64Ptr(v uint64) *uint64    { return &v }
func BoolPtr(v bool) *bool          { return &v }
func ByteSlicePtr(v []byte) *[]byte { return &v }
