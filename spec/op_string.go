// Code generated by "stringer -linecomment -output op_string.go -type CompOp"; DO NOT EDIT.

package spec

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[EqualTo-1]
	_ = x[NotEqualTo-2]
	_ = x[LessThan-3]
	_ = x[GreaterThan-4]
	_ = x[LessThanEqualTo-5]
	_ = x[GreaterThanEqualTo-6]
}

const _CompOp_name = "==!=<><=>="

var _CompOp_index = [...]uint8{0, 2, 4, 5, 6, 8, 10}

func (i CompOp) String() string {
	i -= 1
	if i >= CompOp(len(_CompOp_index)-1) {
		return "CompOp(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _CompOp_name[_CompOp_index[i]:_CompOp_index[i+1]]
}
