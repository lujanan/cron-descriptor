package main

import (
	"cron-descriptor/locale"
	"errors"
	"fmt"
	"golang.org/x/text/message"
	"regexp"
	"strconv"
	"strings"
)

var (
	WeekDayName = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	MonthName   = map[int]string{
		1:  "January",
		2:  "February",
		3:  "March",
		4:  "April",
		5:  "May",
		6:  "June",
		7:  "July",
		8:  "August",
		9:  "September",
		10: "October",
		11: "November",
		12: "December",
	}

	specialCharactersList = []string{"/", "-", ",", "*"}
	specialCharacters     = strings.Join(specialCharactersList, "")
)

type descriptor struct {
	Expression string
	Printer    *message.Printer
	Options    *options
}

func DefaultDescription(expression string) string {
	if expression == "" {
		return ""
	}

	opts := NewDefaultOptions()
	return NewDescriptor(expression, opts).GetDescription()
}

func NewDescriptor(expression string, opts *options) *descriptor {
	printer := locale.NewPrinter(opts.Language)
	return &descriptor{
		Expression: expression,
		Printer:    printer,
		Options:    opts,
	}
}

func (self *descriptor) GetDescription() string {
	entity, err := parse(self)
	if err != nil {
		return err.Error()
	}
	description := ""

	switch self.Options.DescriptionType {
	case DescFull:
		description, err = self.getFullDescription(entity)
		if err != nil {
			description = err.Error()
		}

	case DescTimeOfDay:
		description, err = self.getTimeOfDayDescription(entity)
		if err != nil {
			description = err.Error()
		}

	case DescHours:
		description = self.getHoursDescription(entity)

	case DescMinutes:
		description = self.getMinutesDescription(entity)

	case DescSeconds:
		description = self.getSecondsDescription(entity)

	case DescDayOfMonth:
		description = self.getDayOfMonthDescription(entity)

	case DescMonth:
		description = self.getMonthDescription(entity)

	case DescDayOfWeek:
		description = self.getDayOfWeekDescription(entity)

	case DescYear:
		description = self.getYearDescription(entity)

	default:
		description = "error type"
	}

	return description
}

func (self *descriptor) getFullDescription(entity *cronEntity) (string, error) {
	timeSegment, err := self.getTimeOfDayDescription(entity)
	if err != nil {
		return "", err
	}

	dayOfMonthDesc := self.getDayOfMonthDescription(entity)
	monthDesc := self.getMonthDescription(entity)
	dayOfWeekDesc := self.getDayOfWeekDescription(entity)
	yearDesc := self.getYearDescription(entity)

	description := fmt.Sprintf("%s%s%s%s%s", timeSegment, dayOfMonthDesc, dayOfWeekDesc, monthDesc, yearDesc)
	description = self.transformVerbosity(description)
	description = self.transformCase(description)

	return description, nil
}

func (self *descriptor) getTimeOfDayDescription(entity *cronEntity) (string, error) {
	secondsExp := entity.Seconds
	minutesExp := entity.Minutes
	hoursExp := entity.Hours

	description := make([]string, 0)

	//handle special cases first
	if !strings.ContainsAny(secondsExp, specialCharacters) &&
		!strings.ContainsAny(minutesExp, specialCharacters) &&
		!strings.ContainsAny(hoursExp, specialCharacters) {

		//specific time of day (i.e. 10 14)
		formatTimeStr, err := self.formatTime(hoursExp, minutesExp, secondsExp)
		if err != nil {
			return "", err
		}
		description = append(description, self.Printer.Sprintf("At "), formatTimeStr)

	} else if strings.Contains(minutesExp, "-") &&
		!strings.Contains(minutesExp, ",") &&
		!strings.ContainsAny(hoursExp, specialCharacters) {

		//minute range in single hour (i.e. 0-10 11)
		minuteParts := strings.Split(minutesExp, "-")
		minuteBtw0, err := self.formatTime(hoursExp, minuteParts[0], "")
		if err != nil {
			return "", err
		}
		minuteBtw1, err := self.formatTime(hoursExp, minuteParts[1], "")
		if err != nil {
			return "", err
		}
		description = append(description, self.Printer.Sprintf("Every minute between %s and %s", minuteBtw0, minuteBtw1))

	} else if strings.Contains(hoursExp, ",") &&
		!strings.Contains(hoursExp, "-") &&
		!strings.ContainsAny(minutesExp, specialCharacters) {

		//hours list with single minute (o.e. 30 6,14,16)
		hourParts := strings.Split(hoursExp, ",")
		hourPartsLength := len(hourParts)
		description = append(description, self.Printer.Sprintf("At"))
		for i, hourPart := range hourParts {
			hourFormat, err := self.formatTime(hourPart, minutesExp, "")
			if err != nil {
				return "", err
			}
			description = append(description, " ", hourFormat)
			if i < (hourPartsLength - 2) {
				description = append(description, ",")
			}
			if i == (hourPartsLength - 2) {
				description = append(description, self.Printer.Sprintf(" and"))
			}
		}

	} else {
		//default time description
		secondsDescription := self.getSecondsDescription(entity)
		minutesDescription := self.getMinutesDescription(entity)
		hoursDescription := self.getHoursDescription(entity)

		if secondsDescription != "" {
			description = append(description, secondsDescription)
		}

		if len(description) > 0 {
			description = append(description, ", ")
		}
		if minutesDescription != "" {
			description = append(description, minutesDescription)
		}

		if len(description) > 0 {
			description = append(description, ", ")
		}
		if hoursDescription != "" {
			description = append(description, hoursDescription)
		}
	}

	return strings.Join(description, ""), nil
}

func (self *descriptor) getSecondsDescription(entity *cronEntity) string {
	expression := entity.Seconds
	fnAllDescription := func(printer *message.Printer) string {
		return printer.Sprintf("every second")
	}
	fnGetSingleItemDescription := func(_ *message.Printer, s string) string {
		return s
	}
	fnGetIntervalDescriptionFormat := func(printer *message.Printer, format, _ string) string {
		return printer.Sprintf("every %s seconds", format)
	}
	fnGetBetweenDescriptionFormat := func(printer *message.Printer, _ string, s ...string) string {
		maxParams := 2
		objectList := make([]interface{}, 0)
		for _, val := range s {
			objectList = append(objectList, val)
			if len(objectList) >= maxParams {
				break
			}
		}
		return printer.Sprintf("seconds %s through %s past the minute", objectList...)
	}
	fnGetDescriptionFormat := func(printer *message.Printer, _, s string) string {
		return printer.Sprintf("at %s seconds past the minute", s)
	}

	return self.getSegmentDescription(
		expression,
		fnAllDescription,
		fnGetSingleItemDescription,
		fnGetIntervalDescriptionFormat,
		fnGetBetweenDescriptionFormat,
		fnGetDescriptionFormat)
}

func (self *descriptor) getMinutesDescription(entity *cronEntity) string {
	expression := entity.Minutes
	fnAllDescription := func(printer *message.Printer) string {
		return printer.Sprintf("every minute")
	}
	fnGetSingleItemDescription := func(printer *message.Printer, s string) string {
		return s
	}
	fnGetIntervalDescriptionFormat := func(printer *message.Printer, format, _ string) string {
		return printer.Sprintf("every %s minutes", format)
	}
	fnGetBetweenDescriptionFormat := func(printer *message.Printer, _ string, s ...string) string {
		maxParams := 2
		objectList := make([]interface{}, 0)
		for _, val := range s {
			objectList = append(objectList, val)
			if len(objectList) >= maxParams {
				break
			}
		}
		return printer.Sprintf("minutes %s through %s past the hour", objectList...)
	}
	fnGetDescriptionFormat := func(printer *message.Printer, _, s string) string {
		if s == "0" {
			return ""
		} else {
			return printer.Sprintf("at %s minutes past the hour", s)
		}
	}

	return self.getSegmentDescription(
		expression,
		fnAllDescription,
		fnGetSingleItemDescription,
		fnGetIntervalDescriptionFormat,
		fnGetBetweenDescriptionFormat,
		fnGetDescriptionFormat)
}

func (self *descriptor) getHoursDescription(entity *cronEntity) string {
	expression := entity.Hours

	fnAllDescription := func(printer *message.Printer, ) string {
		return printer.Sprintf("every hour")
	}
	fnGetSingleItemDescription := func(_ *message.Printer, s string) string {
		hourStr, err := self.formatTime(s, "0", "")
		if err != nil {
			hourStr = ""
		}
		return hourStr
	}
	fnGetIntervalDescriptionFormat := func(printer *message.Printer, format, _ string) string {
		return printer.Sprintf("every %s hours", format)
	}
	fnGetBetweenDescriptionFormat := func(printer *message.Printer, _ string, s ...string) string {
		maxParams := 2
		objectList := make([]interface{}, 0)
		for _, val := range s {
			objectList = append(objectList, val)
			if len(objectList) >= maxParams {
				break
			}
		}
		return printer.Sprintf("between %s and %s", objectList...)
	}
	fnGetDescriptionFormat := func(printer *message.Printer, _, s string) string {
		return printer.Sprintf("at %s", s)
	}

	return self.getSegmentDescription(
		expression,
		fnAllDescription,
		fnGetSingleItemDescription,
		fnGetIntervalDescriptionFormat,
		fnGetBetweenDescriptionFormat,
		fnGetDescriptionFormat)
}

func (self *descriptor) getDayOfWeekDescription(entity *cronEntity) string {
	if entity.DayOfWeek == "*" && entity.DayOfMonth != "*" {
		return ""
	}

	getDayName := func(s string) string {
		exp := s
		if strings.Contains(s, "#") {
			expList := strings.SplitN(s, "#", 2)
			exp = expList[0]
		} else if strings.Contains(s, "L") {
			exp = strings.Replace(exp, "L", "", -1)
		}
		expNum, err := strconv.Atoi(exp)
		if err != nil {
			return ""
		}
		return self.numberToDay(expNum)
	}

	getFormat := func(s string) string {
		formated := ""
		if strings.Contains(s, "#") {
			dayOfWeekOfMonthList := strings.SplitN(s, "#", 2)
			dayOfWeekOfMonth := ""
			dayOfWeekOfMonthDescription := ""
			if len(dayOfWeekOfMonthList) == 2 {
				dayOfWeekOfMonth = dayOfWeekOfMonthList[1]
			}

			if num, err := strconv.Atoi(dayOfWeekOfMonth); err == nil && num >= 1 && num <= 5 {
				choices := map[int]string{
					1: "first",
					2: "second",
					3: "third",
					4: "forth",
					5: "fifth",
				}
				dayOfWeekOfMonthDescription = choices[num]
			}

			formated = self.Printer.Sprintf(", on the ") + self.Printer.Sprintf(dayOfWeekOfMonthDescription) + self.Printer.Sprintf(" %s of the month")

		} else if strings.Contains(s, "L") {
			formated = self.Printer.Sprintf(", on the last %s of the month")

		} else {
			formated = self.Printer.Sprintf(", only on %s")
		}
		return formated
	}

	fnAllDescription := func(printer *message.Printer) string {
		return printer.Sprintf(", every day")
	}
	fnGetSingleItemDescription := func(printer *message.Printer, s string) string {
		return getDayName(s)
	}
	fnGetIntervalDescriptionFormat := func(printer *message.Printer, format, _ string) string {
		return printer.Sprintf(", every %s days of the week", format)
	}
	fnGetBetweenDescriptionFormat := func(printer *message.Printer, _ string, s ...string) string {
		maxParams := 2
		objectList := make([]interface{}, 0)
		for _, val := range s {
			objectList = append(objectList, val)
			if len(objectList) >= maxParams {
				break
			}
		}
		return printer.Sprintf(", %s through %s", objectList...)
	}
	fnGetDescriptionFormat := func(printer *message.Printer, format, s string) string {
		fmt.Println(getFormat(format))
		return printer.Sprintf(getFormat(format), s)
	}

	return self.getSegmentDescription(
		entity.DayOfWeek,
		fnAllDescription,
		fnGetSingleItemDescription,
		fnGetIntervalDescriptionFormat,
		fnGetBetweenDescriptionFormat,
		fnGetDescriptionFormat)
}

func (self *descriptor) getMonthDescription(entity *cronEntity) string {
	fnAllDescription := func(_ *message.Printer) string {
		return ""
	}
	fnGetSingleItemDescription := func(printer *message.Printer, s string) string {
		month, err := strconv.Atoi(s)
		if err != nil || month < 1 || month > 12 {
			return ""
		}
		return MonthName[month]
	}
	fnGetIntervalDescriptionFormat := func(printer *message.Printer, format, _ string) string {
		return printer.Sprintf(", every %s months", format)
	}
	fnGetBetweenDescriptionFormat := func(printer *message.Printer, _ string, s ...string) string {
		maxParams := 2
		objectList := make([]interface{}, 0)
		for _, val := range s {
			objectList = append(objectList, val)
			if len(objectList) >= maxParams {
				break
			}
		}
		return printer.Sprintf(", %s through %s", objectList...)
	}
	fnGetDescriptionFormat := func(printer *message.Printer, _, s string) string {
		return printer.Sprintf(", only in %s", s)
	}

	return self.getSegmentDescription(
		entity.Month,
		fnAllDescription,
		fnGetSingleItemDescription,
		fnGetIntervalDescriptionFormat,
		fnGetBetweenDescriptionFormat,
		fnGetDescriptionFormat)
}

func (self *descriptor) getDayOfMonthDescription(entity *cronEntity) string {
	expression := entity.DayOfMonth
	expression = strings.Replace(expression, "?", "*", -1)
	description := ""

	if expression == "L" {
		description = self.Printer.Sprintf(", on the last day of the month")

	} else if expression == "LW" || expression == "WL" {
		description = self.Printer.Sprintf(", on the last weekday of the month")

	} else {
		regexRule := regexp.MustCompile(`(\d{1,2}W)|(W\d{1,2})`)
		if regexRule.MatchString(expression) {
			dayString := ""
			dayNum, err := strconv.Atoi(strings.Replace(expression, "W", "", -1))
			if err == nil {
				if dayNum == 1 {
					dayString = self.Printer.Sprintf("first weekday")
				} else {
					dayString = self.Printer.Sprintf("weekday nearest day %s", strconv.Itoa(dayNum))
				}
			}
			description = self.Printer.Sprintf(", on the %s of the month", dayString)

		} else {
			fnAllDescription := func(printer *message.Printer) string {
				return printer.Sprintf(", every day")
			}
			fnGetSingleItemDescription := func(_ *message.Printer, s string) string {
				return s
			}
			fnGetIntervalDescriptionFormat := func(printer *message.Printer, format, s string) string {
				if format == "1" {
					return printer.Sprintf(", every day")
				} else {
					return printer.Sprintf(", every %s days", s)
				}
			}
			fnGetBetweenDescriptionFormat := func(printer *message.Printer, _ string, s ...string) string {
				maxParams := 2
				objectList := make([]interface{}, 0)
				for _, val := range s {
					objectList = append(objectList, val)
					if len(objectList) >= maxParams {
						break
					}
				}
				return printer.Sprintf(", between day %s and %s of the month", objectList...)
			}
			fnGetDescriptionFormat := func(printer *message.Printer, _, s string) string {
				return printer.Sprintf(", on day %s of the month", s)
			}
			description = self.getSegmentDescription(
				expression,
				fnAllDescription,
				fnGetSingleItemDescription,
				fnGetIntervalDescriptionFormat,
				fnGetBetweenDescriptionFormat,
				fnGetDescriptionFormat)
		}
	}

	return description
}

func (self *descriptor) getYearDescription(entity *cronEntity) string {
	formatYear := func(s string) string {
		regexpRule := regexp.MustCompile(`^\d+$`)
		if regexpRule.MatchString(s) {
			year, err := strconv.Atoi(s)
			if err != nil {
				return s
			}
			return strconv.Itoa(year)

		} else {
			return s
		}
	}

	fnAllDescription := func(_ *message.Printer) string {
		return ""
	}
	fnGetSingleItemDescription := func(_ *message.Printer, s string) string {
		return formatYear(s)
	}
	fnGetIntervalDescriptionFormat := func(printer *message.Printer, format, _ string) string {
		return printer.Sprintf(", every %s years", format)
	}
	fnGetBetweenDescriptionFormat := func(printer *message.Printer, _ string, s ...string) string {
		maxParams := 2
		objectList := make([]interface{}, 0)
		for _, val := range s {
			objectList = append(objectList, val)
			if len(objectList) >= maxParams {
				break
			}
		}
		return printer.Sprintf(", %s through %s", objectList...)
	}
	fnGetDescriptionFormat := func(printer *message.Printer, _, s string) string {
		return printer.Sprintf(", only in %s", s)
	}

	return self.getSegmentDescription(
		entity.Year,
		fnAllDescription,
		fnGetSingleItemDescription,
		fnGetIntervalDescriptionFormat,
		fnGetBetweenDescriptionFormat,
		fnGetDescriptionFormat)
}

func (self *descriptor) getSegmentDescription(
	expression string,
	fnAllDescription func(printer *message.Printer) string,
	fnGetSingleItemDescription func(printer *message.Printer, s string) string,
	fnGetIntervalDescriptionFormat func(printer *message.Printer, format, s string) string,
	fnGetBetweenDescriptionFormat func(printer *message.Printer, format string, s ... string) string,
	fnGetDescriptionFormat func(printer *message.Printer, format, s string) string,
) string {

	description := ""
	if expression == "" {

	} else if expression == "*" {
		description = fnAllDescription(self.Printer)

	} else if !strings.ContainsAny(expression, "/-,") {
		description = fnGetDescriptionFormat(self.Printer, expression, fnGetSingleItemDescription(self.Printer, expression))

	} else if strings.Contains(expression, "/") {
		segments := strings.Split(expression, "/")
		description = fnGetIntervalDescriptionFormat(self.Printer, segments[1], fnGetSingleItemDescription(self.Printer, segments[1]))

		//interval contains 'between' piece (i.e. 2-59/3 )
		if strings.Contains(segments[0], "-") {
			betweenSegmentDescription := self.generateBetweenSegmentDescription(segments[0], fnGetBetweenDescriptionFormat, fnGetSingleItemDescription)
			if !regexp.MustCompile(`^(,\s)`).MatchString(betweenSegmentDescription) {
				// not start with ", "
				description += ", "
			}
			description += betweenSegmentDescription

		} else if !strings.ContainsAny(segments[0], "*,") {
			rangeItemDescription := fnGetDescriptionFormat(self.Printer, segments[0], fnGetSingleItemDescription(self.Printer, segments[0]))
			rangeItemDescription = strings.Replace(rangeItemDescription, ", ", "", -1)
			description += self.Printer.Sprintf(", starting %s", rangeItemDescription)
		}

	} else if strings.Contains(expression, ",") {
		segments := strings.Split(expression, ",")
		segmentsLength := len(segments)
		descriptionContent := ""

		for i, segment := range segments {
			if i > 0 && segmentsLength > 2 {
				descriptionContent += ","
				if i < (segmentsLength - 1) {
					descriptionContent += " "
				}
			}

			if i > 0 && segmentsLength > 1 && (i == (segmentsLength-1) || segmentsLength == 2) {
				descriptionContent += self.Printer.Sprintf(" and ")
			}

			if strings.Contains(segment, "-") {
				betweenDescription := self.generateBetweenSegmentDescription(
					segment,
					func(printer *message.Printer, _ string, s ...string) string {
						maxParams := 2
						objectList := make([]interface{}, 0)
						for _, val := range s {
							objectList = append(objectList, val)
							if len(objectList) >= maxParams {
								break
							}
						}
						return printer.Sprintf(", %s through %s", objectList...)
					},
					fnGetSingleItemDescription)
				betweenDescription = strings.Replace(betweenDescription, ", ", "", -1)
				descriptionContent += betweenDescription

			} else {
				descriptionContent += fnGetSingleItemDescription(self.Printer, segment)
			}
		}

		description = fnGetDescriptionFormat(self.Printer, expression, descriptionContent)

	} else if strings.Contains(expression, "-") {
		description = self.generateBetweenSegmentDescription(expression, fnGetBetweenDescriptionFormat, fnGetSingleItemDescription)
	}

	return description
}

func (self *descriptor) generateBetweenSegmentDescription(
	betweenExpression string,
	fnGetBetweenDescritionFormat func(printer *message.Printer, format string, s ...string) string,
	fnGetSingleItemDescription func(printer *message.Printer, s string) string,
) string {
	description := ""
	betweenSegments := strings.Split(betweenExpression, "-")
	betweenSegment1Description := fnGetSingleItemDescription(self.Printer, betweenSegments[0])
	betweenSegment2Description := fnGetSingleItemDescription(self.Printer, betweenSegments[1])
	betweenSegment2Description = strings.Replace(betweenSegment2Description, ":00", ":59", -1)

	description += fnGetBetweenDescritionFormat(self.Printer, betweenExpression, betweenSegment1Description, betweenSegment2Description)
	return description
}

func (self *descriptor) formatTime(hourExp, minuteExp, secondsExp string) (string, error) {
	hour, err := strconv.Atoi(hourExp)
	if err != nil {
		return "", err
	}
	if hour < 0 || hour > 24 {
		return "", errors.New("error hours")
	}

	period := ""
	if !self.Options.Use24hourTimeFormat {
		if hour >= 12 {
			period = " PM"
		} else {
			period = " AM"
		}

		if hour > 12 {
			hour -= 12
		}
	}
	hourExp = fmt.Sprintf("%02d", hour)

	minute := 0
	if minute, err = strconv.Atoi(minuteExp); err != nil {
		return "", err
	}
	if minute < 0 || minute > 59 {
		return "", errors.New("error minutes")
	}
	minuteExp = fmt.Sprintf("%02d", minute)

	if secondsExp != "" {
		if seconds, err := strconv.Atoi(secondsExp); err != nil {
			return "", err
		} else {
			if seconds < 0 || seconds > 59 {
				return "", errors.New("error seconds")
			}
			secondsExp = fmt.Sprintf(":%02d", seconds)
		}
	}

	return fmt.Sprintf("%s:%s%s%s", hourExp, minuteExp, secondsExp, period), nil
}

func (self *descriptor) transformVerbosity(description string) string {
	if !self.Options.Verbose {
		description = strings.Replace(description, self.Printer.Sprintf(", every minute"), "", -1)
		description = strings.Replace(description, self.Printer.Sprintf(", every hour"), "", -1)
		description = strings.Replace(description, self.Printer.Sprintf(", every day"), "", -1)
	}
	return description
}

func (self *descriptor) transformCase(description string) string {
	switch self.Options.CasingType {
	case CasingSentence:
		descriptionParts := strings.Split(description, " ")
		descriptionParts[0] = strings.Title(descriptionParts[0])
		description = strings.Join(descriptionParts, " ")

	case CasingTitle:
		description = strings.Title(description)

	default:
		description = strings.ToLower(description)
	}
	return description
}

func (self *descriptor) numberToDay(dayNumber int) string {
	if dayNumber < 0 || dayNumber >= len(WeekDayName) {
		return ""
	}
	return self.Printer.Sprintf(WeekDayName[dayNumber])
}
