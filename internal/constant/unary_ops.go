package constant

// UnaryOps is a mapping of operators to the functions that implement tham.
var UnaryOps = map[string]func(Value) Value{
	"!": Not,
	"-": Neg,
}

// Not performs logical "!".
func Not(x Value) Value {
	v, ok := ToBool(x)
	if !ok {
		return UnknownValue{}
	}
	return BoolValue(!v)
}

// Neg performs arithmetic unary "-".
func Neg(x Value) Value {
	switch x := x.(type) {
	case IntValue:
		return -x
	case FloatValue:
		return -x
	case BoolValue:
		if x {
			return IntValue(-1)
		}
		return IntValue(0)
	}
	return UnknownValue{}
}
