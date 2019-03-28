package constant

// BinaryOps is a mapping of operators to the functions that implement tham.
var BinaryOps = map[string]func(Value, Value) Value{
	"||": Or,
	"&&": And,

	"+": Add,
	"-": Sub,

	"===": Identical,
	"==":  Equal,
	">":   GreaterThan,
	"<":   LessThan,
}

// Or performs logical "||".
func Or(x, y Value) Value {
	v1, ok1 := ToBool(x)
	v2, ok2 := ToBool(y)
	switch {
	case ok1 && bool(v1):
		return BoolValue(true)
	case ok2 && bool(v2):
		return BoolValue(true)
	case ok1 && ok2:
		return v1 || v2
	default:
		return UnknownValue{}
	}
}

// And performs logical "&&".
func And(x, y Value) Value {
	v1, ok1 := ToBool(x)
	v2, ok2 := ToBool(y)
	switch {
	case ok1 && bool(!v1):
		return BoolValue(false)
	case ok2 && bool(!v2):
		return BoolValue(false)
	case ok1 && ok2:
		return v1 && v2
	default:
		return UnknownValue{}
	}
}

// Sub performs arithmetic "-".
func Sub(x, y Value) Value {
	switch x := x.(type) {
	case IntValue:
		y, ok := y.(IntValue)
		if ok {
			return x - y
		}
	case FloatValue:
		y, ok := y.(FloatValue)
		if ok {
			return x - y
		}
	}
	return UnknownValue{}
}

// Add performs arithmetic "+".
func Add(x, y Value) Value {
	switch x := x.(type) {
	case IntValue:
		y, ok := y.(IntValue)
		if ok {
			return x + y
		}
	case FloatValue:
		y, ok := y.(FloatValue)
		if ok {
			return x + y
		}
	case StringValue:
		y, ok := y.(StringValue)
		if ok {
			return x + y
		}
	}
	return UnknownValue{}
}

// Identical performs "===" comparison.
func Identical(x, y Value) Value {
	switch x := x.(type) {
	case IntValue:
		y, ok := y.(IntValue)
		if ok {
			return BoolValue(x == y)
		}
	case FloatValue:
		y, ok := y.(FloatValue)
		if ok {
			return BoolValue(x == y)
		}
	case StringValue:
		y, ok := y.(StringValue)
		if ok {
			return BoolValue(x == y)
		}
	}
	return UnknownValue{}
}

// Equal performs "==" comparison.
func Equal(x, y Value) Value {
	// TODO(quasilyte): support non-strict forms of comparison?
	return Identical(x, y)
}

// GreaterThan performs ">" comparison.
func GreaterThan(x, y Value) Value {
	// TODO(quasilyte): support non-strict forms of comparison?
	switch x := x.(type) {
	case IntValue:
		y, ok := y.(IntValue)
		if ok {
			return BoolValue(x > y)
		}
	case FloatValue:
		y, ok := y.(FloatValue)
		if ok {
			return BoolValue(x > y)
		}
	case StringValue:
		y, ok := y.(StringValue)
		if ok {
			return BoolValue(x > y)
		}
	}
	return UnknownValue{}
}

// LessThan performs "<" comparison.
func LessThan(x, y Value) Value {
	// TODO(quasilyte): support non-strict forms of comparison?
	switch x := x.(type) {
	case IntValue:
		y, ok := y.(IntValue)
		if ok {
			return BoolValue(x < y)
		}
	case FloatValue:
		y, ok := y.(FloatValue)
		if ok {
			return BoolValue(x < y)
		}
	case StringValue:
		y, ok := y.(StringValue)
		if ok {
			return BoolValue(x < y)
		}
	}
	return UnknownValue{}
}
