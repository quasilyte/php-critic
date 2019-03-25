package main

type (
	UnknownConst struct{}
	IntConst     int64
	FloatConst   float64
	StringConst  string
	BoolConst    bool
)

type Constant interface {
	isValid() bool
}

func (c UnknownConst) isValid() bool { return false }
func (c IntConst) isValid() bool     { return true }
func (c FloatConst) isValid() bool   { return true }
func (c StringConst) isValid() bool  { return true }
func (c BoolConst) isValid() bool    { return true }

func constToBool(x Constant) (BoolConst, bool) {
	switch x := x.(type) {
	case BoolConst:
		return x, true
	case IntConst:
		return BoolConst(x != 0), true
	}
	return false, false
}

func constOr(x, y Constant) Constant {
	v1, ok1 := constToBool(x)
	v2, ok2 := constToBool(y)
	switch {
	case ok1 && bool(v1):
		return BoolConst(true)
	case ok2 && bool(v2):
		return BoolConst(true)
	case ok1 && ok2:
		return v1 || v2
	default:
		return UnknownConst{}
	}
}

func constAnd(x, y Constant) Constant {
	v1, ok1 := constToBool(x)
	v2, ok2 := constToBool(y)
	switch {
	case ok1 && bool(!v1):
		return BoolConst(false)
	case ok2 && bool(!v2):
		return BoolConst(false)
	case ok1 && ok2:
		return v1 && v2
	default:
		return UnknownConst{}
	}
}

func constNegate(x Constant) Constant {
	switch x := x.(type) {
	case IntConst:
		return -x
	case FloatConst:
		return -x
	case BoolConst:
		if x {
			return IntConst(-1)
		}
		return IntConst(0)
	}
	return UnknownConst{}
}

func constSub(x, y Constant) Constant {
	switch x := x.(type) {
	case IntConst:
		y, ok := y.(IntConst)
		if ok {
			return x - y
		}
	case FloatConst:
		y, ok := y.(FloatConst)
		if ok {
			return x - y
		}
	}
	return UnknownConst{}
}

func constAdd(x, y Constant) Constant {
	switch x := x.(type) {
	case IntConst:
		y, ok := y.(IntConst)
		if ok {
			return x + y
		}
	case FloatConst:
		y, ok := y.(FloatConst)
		if ok {
			return x + y
		}
	case StringConst:
		y, ok := y.(StringConst)
		if ok {
			return x + y
		}
	}
	return UnknownConst{}
}

func constIdentical(x, y Constant) Constant {
	switch x := x.(type) {
	case IntConst:
		y, ok := y.(IntConst)
		if ok {
			return BoolConst(x == y)
		}
	case FloatConst:
		y, ok := y.(FloatConst)
		if ok {
			return BoolConst(x == y)
		}
	case StringConst:
		y, ok := y.(StringConst)
		if ok {
			return BoolConst(x == y)
		}
	}
	return UnknownConst{}
}

func constEqual(x, y Constant) Constant {
	// TODO(quasilyte): support non-strict forms of comparison?
	return constIdentical(x, y)
}

func constGreaterThan(x, y Constant) Constant {
	// TODO(quasilyte): support non-strict forms of comparison?
	switch x := x.(type) {
	case IntConst:
		y, ok := y.(IntConst)
		if ok {
			return BoolConst(x > y)
		}
	case FloatConst:
		y, ok := y.(FloatConst)
		if ok {
			return BoolConst(x > y)
		}
	case StringConst:
		y, ok := y.(StringConst)
		if ok {
			return BoolConst(x > y)
		}
	}
	return UnknownConst{}
}

func constLessThan(x, y Constant) Constant {
	// TODO(quasilyte): support non-strict forms of comparison?
	switch x := x.(type) {
	case IntConst:
		y, ok := y.(IntConst)
		if ok {
			return BoolConst(x < y)
		}
	case FloatConst:
		y, ok := y.(FloatConst)
		if ok {
			return BoolConst(x < y)
		}
	case StringConst:
		y, ok := y.(StringConst)
		if ok {
			return BoolConst(x < y)
		}
	}
	return UnknownConst{}
}
