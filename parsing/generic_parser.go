package parsing

type BaseParser struct {
	source CharSource
	ch     rune // = -1
}

const END rune = 0

func (bp *BaseParser) GetSource() CharSource {
	return bp.source
}

func NewBaseParser(cs CharSource) *BaseParser {
	r := &BaseParser{source: cs}
	r.TakeNext()
	return r
}

func (bp *BaseParser) TakeNext() rune { // Take
	result := bp.ch
	if bp.source.HasNext() {
		bp.ch = bp.source.Next()
	} else {
		bp.ch = END
	}
	return result
}

func (bp *BaseParser) Test(expected rune) bool {
	return bp.ch == expected
}

func (bp *BaseParser) Take(expected rune) bool {
	if bp.Test(expected) {
		bp.TakeNext()
		return true
	}
	return false
}

func (bp *BaseParser) Expect(expected rune) {
	if !bp.Take(expected) {
		panic(bp.Error("expected '" + string(expected) + "', found '" + string(bp.ch) + "'"))
	}
}

func (bp *BaseParser) ExpectString(value string) { // Expect
	for _, c := range value {
		bp.Expect(c)
	}
}

func (bp *BaseParser) Eof() bool {
	return bp.Take(END)
}

func (bp *BaseParser) Error(message string) error {
	return bp.source.Error(message)
}

func (bp *BaseParser) Between(from, to rune) bool {
	return from <= bp.ch && bp.ch <= to
}

func (bp *BaseParser) From(src string) bool {
	for _, c := range src {
		if bp.Test(c) {
			return true
		}
	}
	return false
}
