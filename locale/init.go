package locale

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	EN_US = iota
	ZH_CN
)

var (
	defaultLanguageTag = language.English

	languageTypeList = map[int]language.Tag{
		EN_US: language.AmericanEnglish,
		ZH_CN: language.Chinese,
	}

	localeList = map[int]map[string]string{
		ZH_CN: zhCN,
	}
)

func NewPrinter(localeType int) *message.Printer {
	languageTag, ok := languageTypeList[localeType]
	if !ok {
		return message.NewPrinter(defaultLanguageTag)
	}
	languageList, ok := localeList[localeType]
	if !ok {
		return message.NewPrinter(defaultLanguageTag)
	}

	for key, val := range languageList {
		message.SetString(languageTag, key, val)
	}
	return message.NewPrinter(languageTag)
}
