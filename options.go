package main

type Options struct {
	ThrowExceptionOnParseError bool
	CasingType                 int
	Verbose                    bool
	DayOfWeekStartIndexZero    bool
	Use24hourTimeFormat        bool
}

var options *Options

func init() {
	options = &Options{
		ThrowExceptionOnParseError: false,
		CasingType:                 CasingSentence,
		Verbose:                    false,
		DayOfWeekStartIndexZero:    true,
		Use24hourTimeFormat:        false,
	}
}
