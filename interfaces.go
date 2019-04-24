package main

type segmentDescription interface {
	FnAllDescription() string
	FnGetSingleItemDescription(s string) string
	FnGetIntervalDescriptionFormat(format, s string) string
	FnGetBetweenDescriptionFormat(format string, s ... string) string
	FnGetDescriptionFormat(format, s string) string
}
