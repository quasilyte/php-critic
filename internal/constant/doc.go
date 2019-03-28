// Package constant implements basic PHP operations over
// constant values of builtin types.
//
// This package can be used to compute constant expressions
// and to carry around values those values are known.
//
// Special type UnknownValue can be used to express "unknown"
// value that can't be resolved, but is still a valid argument
// to the PHP operators.
//
// Note that operations are implemented in conservative way.
// Some operations may return UnknownValue more often than
// they should.
package constant
