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

	Skip = 0x100
)

type Stop struct {
	Code int
}

func Result(rc int) int {
	return rc & ResultMask
}

func IsErroneous(rc int) bool {
	return Result(rc)&ErrorMask != 0
}

func Has(rc int, flag int) bool {
	return rc&flag != 0
}

func Add(current *int, rc int) int {
	value := Result(rc)
	if current == nil {
		return value
	}

	*current = Result(*current) | value
	return *current
}

func Value(current *int) int {
	if current == nil {
		return OK
	}

	return Result(*current)
}
