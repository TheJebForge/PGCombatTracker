package utils

// Stringed So they have Stringer interface, but strings don't fking implement that?
type Stringed string

func (s Stringed) String() string {
	return string(s)
}
