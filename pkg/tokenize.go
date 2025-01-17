package lisp

/**
s_expression = atomic_symbol \
              / "(" s_expression "."s_expression ")" \
              / list

list = "(" s_expression < s_expression > ")"

atomic_symbol = letter atom_part

atom_part = empty / letter atom_part / number atom_part

letter = "a" / "b" / " ..." / "z"

number = "1" / "2" / " ..." / "9"

empty = " "




<s_expression>  ::= <atomic_symbol>
                  | "(" <s_expression> "." <s_expression> ")"
                  | <list> .

<_list>         ::= <s_expression> <_list>
                  | <s_expression> .

<list>          ::= "(" <s_expression> <_list> ")" .

<atomic_symbol> ::= <letter> <atom_part> .

<atom_part>     ::= <empty> | <letter> <atom_part> | <number> <atom_part> .

<letter>        ::= "a" | "b" | "c" | "d" | "e" | "f" | "g" | "h" | "i" | "j"
                  | "k" | "l" | "m" | "n" | "o" | "p" | "q" | "r" | "s" | "t"
                  | "u" | "v" | "w" | "x" | "y" | "z" .

<number>        ::= "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" | "0" .

<empty>         ::= " ".

**/

const (
	EMPTY_TOKEN  = iota
	NUMBER       // 1
	ATOM         // x
	STRING       // "x"
	KEYWORD      // :x
	BRACE_OPEN   // (
	BRACE_CLOSE  // )
	QUOTE        // '(...) / 'SYMBOL
	S_EXPRESSION // (...)
	COMMA        // ,
	COMMA_AT     // ,@
)
