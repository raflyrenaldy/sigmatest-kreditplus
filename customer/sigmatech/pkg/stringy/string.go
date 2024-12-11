package stringy

import "strings"

type Separator string

const (
	CommaSeparator      Separator = ","
	SpaceSeparator      Separator = " "
	DashSeparator       Separator = "-"
	UnderscoreSeparator Separator = "_"
	SlashSeparator      Separator = "/"
	ColonSeparator      Separator = ":"
	SemiColonSeparator  Separator = ";"
	DotSeparator        Separator = "."
	PipeSeparator       Separator = "|"
	TildeSeparator      Separator = "~"
	PlusSeparator       Separator = "+"
	EqualSeparator      Separator = "="
	QuestionSeparator   Separator = "?"
	EmptySeparator      Separator = ""
)

func (s Separator) ToString() string {
	return string(s)
}

func ToFlatString(values ...string) string {
	return strings.Join(values, EmptySeparator.ToString())
}

func ToFlatStringWithSeparator(separator Separator, values ...string) string {
	return strings.Join(values, separator.ToString())
}
