// Code generated by "stringer -type=Enforcement"; DO NOT EDIT.

package idkeydigest

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Unknown-0]
	_ = x[Equal-1]
	_ = x[MAAFallback-2]
	_ = x[WarnOnly-3]
}

const _Enforcement_name = "UnknownEqualMAAFallbackWarnOnly"

var _Enforcement_index = [...]uint8{0, 7, 12, 23, 31}

func (i Enforcement) String() string {
	if i >= Enforcement(len(_Enforcement_index)-1) {
		return "Enforcement(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Enforcement_name[_Enforcement_index[i]:_Enforcement_index[i+1]]
}
