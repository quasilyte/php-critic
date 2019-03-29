package constant

// Value is an arbitrary, potentially unresolved (unknown) constant.
type Value interface {
	isValid() bool
}

type (
	// UnknownValue is a value that can't be reliably const-folded.
	//
	// It can be a variable, function call result and other
	// dynamic kinds of values.
	//
	// Can also signify type error that may occur during the operation.
	UnknownValue struct{}

	// IntValue is such an x value that is_integer($x) returns true.
	IntValue int64

	// FloatValue is such an x value that is_float($x) returns true.
	FloatValue float64

	// StringValue is such an x value that is_string($x) returns true.
	StringValue string

	// BoolValue is such an x value that is_bool($x) returns true.
	BoolValue bool
)

// ToBool converts x constant to boolean constants following PHP conversion rules.
// Second bool result tells whether that conversion was successful.
func ToBool(x Value) (BoolValue, bool) {
	switch x := x.(type) {
	case BoolValue:
		return x, true
	case IntValue:
		return BoolValue(x != 0), true
	case FloatValue:
		return BoolValue(x != 0), true
	case StringValue:
		return BoolValue(x != "" && x != "0"), true
	}
	return false, false
}

// ToInt converts x constant to int constants following PHP conversion rules.
// Second bool result tells whether that conversion was successful.
func ToInt(x Value) (IntValue, bool) {
	switch x := x.(type) {
	case BoolValue:
		if x {
			return 1, true
		}
		return 0, true
	case IntValue:
		return x, true
	case FloatValue:
		return IntValue(x), true
	}
	return 0, false
}

func (c UnknownValue) isValid() bool { return false }
func (c IntValue) isValid() bool     { return true }
func (c FloatValue) isValid() bool   { return true }
func (c StringValue) isValid() bool  { return true }
func (c BoolValue) isValid() bool    { return true }
