// Package RC contains the cumulative return-code bitmask used by iasi-dev.
package RC

const (
	OK          = 0x00
	NothingToDo = 0x01
	Info        = 0x02
	Warning     = 0x04
	Attention   = 0x08

	Error    = 0x10
	Severe   = 0x20
	Fatal    = 0x40
	Reserved = 0x80

	NoticeMask = 0x0F
	ErrorMask  = 0xF0
	ResultMask = 0xFF

	// Internal workflow marker. Control flags live above the result byte and are
	// never accumulated into, or exposed as, the process return code.
	Skip = 0x100
)

// Stop is used internally to unwind execution while preserving the accumulated RC.
type Stop struct {
	Code int
}

// Result returns only the externally observable RC byte.
func Result(rc int) int {
	return rc & ResultMask
}

func IsErroneous(rc int) bool {
	return Result(rc)&ErrorMask != 0
}

// Has reports whether rc contains flag. It is intended primarily for internal
// control flags such as Skip.
func Has(rc int, flag int) bool {
	return rc&flag != 0
}

// Add accumulates only the result byte. Internal control flags are deliberately
// not part of the cumulative return code.
func Add(current *int, rc int) int {
	value := Result(rc)
	if current == nil {
		return value
	}

	*current = Result(*current) | value
	return *current
}

// Value returns the externally observable RC byte.
func Value(current *int) int {
	if current == nil {
		return OK
	}

	return Result(*current)
}
