package main

import "cron-descriptor/locale"

type options struct {
	DescriptionType         int
	CasingType              int
	Verbose                 bool
	DayOfWeekStartIndexZero bool
	Use24hourTimeFormat     bool
	Language                int
}

func NewDefaultOptions() *options {
	return &options{
		DescriptionType:         DescFull,
		CasingType:              CasingSentence,
		Verbose:                 false,
		DayOfWeekStartIndexZero: true,
		Use24hourTimeFormat:     false,
		Language:                locale.EN_US,
	}
}
