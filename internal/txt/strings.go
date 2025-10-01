package txt

import "strings"

func Repeat(char rune, length int) string {
	return strings.Repeat(string(char), length)

}

func Line(length int) string {
	return Repeat('─', length)
}

func Separator() string {
	return Line(40)
}
