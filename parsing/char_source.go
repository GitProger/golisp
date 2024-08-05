package parsing

type CharSource interface {
	HasNext() bool
	Next() rune // char
	Error(string) error
}
