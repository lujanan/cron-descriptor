package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
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

	parsed = false
)

func GetDescription(expression string, descriptionType int) string {
	entity, err := parse(expression)
	if err != nil {
		return err.Error()
	}
	//fmt.Println("cron struct:", entity)
	description := ""

	switch descriptionType {
	case DescFull:
		description, err = getFullDescription(entity)
		if err != nil {
			description = err.Error()
		}

	case DescTimeOfDay:
		description, err = getTimeOfDayDescription(entity)
		if err != nil {
			description = err.Error()
		}

	case DescHours:
		description = getHoursDescription(entity)

	case DescMinutes:
		description = getMinutesDescription(entity)

	case DescSeconds:
		description = getSecondsDescription(entity)

	case DescDayOfMonth:
		description = getDayOfMonthDescription(entity)

	case DescMonth:
		description = getMonthDescription(entity)

	case DescDayOfWeek:
		description = getDayOfWeekDescription(entity)

	case DescYear:
		description = getYearDescription(entity)

	default:
		description = "error type"
	}

	return description
}

func getFullDescription(entity *cronEntity) (string, error) {
	timeSegment, err := getTimeOfDayDescription(entity)
	if err != nil {
		return "", err
	}

	dayOfMonthDesc := getDayOfMonthDescription(entity)
	//fmt.Println("dayOfMonthDesc:", dayOfMonthDesc)
	monthDesc := getMonthDescription(entity)
	//fmt.Println("monthDesc:", monthDesc)
	dayOfWeekDesc := getDayOfWeekDescription(entity)
	//fmt.Println("dayOfWeekDesc:", dayOfWeekDesc)
	yearDesc := getYearDescription(entity)
	//fmt.Println("yearDesc:", yearDesc)

	description := fmt.Sprintf("%s%s%s%s%s", timeSegment, dayOfMonthDesc, dayOfWeekDesc, monthDesc, yearDesc)
	description = transformVerbosity(description, Verbose)
	description = transformCase(description, CasingType)

	return description, nil
}

func getTimeOfDayDescription(entity *cronEntity) (string, error) {
	secondsExp := entity.Seconds
	minutesExp := entity.Minutes
	hoursExp := entity.Hours

	description := make([]string, 0)

	//handle special cases first
	if !strings.ContainsAny(secondsExp, specialCharacters) &&
		!strings.ContainsAny(minutesExp, specialCharacters) &&
		!strings.ContainsAny(hoursExp, specialCharacters) {

		//specific time of day (i.e. 10 14)
		formatTimeStr, err := formatTime(hoursExp, minutesExp, secondsExp)
		if err != nil {
			return "", err
		}
		description = append(description, "At ", formatTimeStr)

	} else if strings.Contains(minutesExp, "-") &&
		!strings.Contains(minutesExp, ",") &&
		!strings.ContainsAny(hoursExp, specialCharacters) {

		//minute range in single hour (i.e. 0-10 11)
		minuteParts := strings.Split(minutesExp, "-")
		minuteBtw0, err := formatTime(hoursExp, minuteParts[0], "")
		if err != nil {
			return "", err
		}
		minuteBtw1, err := formatTime(hoursExp, minuteParts[1], "")
		if err != nil {
			return "", err
		}
		description = append(description, fmt.Sprintf("Every minute between %s and %s", minuteBtw0, minuteBtw1))

	} else if strings.Contains(hoursExp, ",") &&
		!strings.Contains(hoursExp, "-") &&
		!strings.ContainsAny(minutesExp, specialCharacters) {

		//hours list with single minute (o.e. 30 6,14,16)
		hourParts := strings.Split(hoursExp, ",")
		hourPartsLength := len(hourParts)
		description = append(description, "At")
		for i, hourPart := range hourParts {
			hourFormat, err := formatTime(hourPart, minutesExp, "")
			if err != nil {
				return "", err
			}
			description = append(description, " ", hourFormat)
			if i < (hourPartsLength - 2) {
				description = append(description, ",")
			}
			if i == (hourPartsLength - 2) {
				description = append(description, " and")
			}
		}

	} else {
		//default time description
		secondsDescription := getSecondsDescription(entity)
		//fmt.Println("secondsDescription:", secondsDescription)
		minutesDescription := getMinutesDescription(entity)
		//fmt.Println("minutesDescription:", minutesDescription)
		hoursDescription := getHoursDescription(entity)
		//fmt.Println("hoursDescription:", hoursDescription)

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

func getSecondsDescription(entity *cronEntity) string {
	expression := entity.Seconds
	fnAllDescription := func() string {
		return "every second"
	}
	fnGetSingleItemDescription := func(s string) string {
		return s
	}
	fnGetIntervalDescriptionFormat := func(format, _ string) string {
		return fmt.Sprintf("every %s seconds", format)
	}
	fnGetBetweenDescriptionFormat := func(_ string, s ...string) string {
		maxParams := 2
		objectList := make([]interface{}, 0)
		for _, val := range s {
			objectList = append(objectList, val)
			if len(objectList) >= maxParams {
				break
			}
		}
		return fmt.Sprintf("seconds %s through %s past the minute", objectList...)
	}
	fnGetDescriptionFormat := func(_, s string) string {
		return fmt.Sprintf("at %s seconds past the minute", s)
	}

	return getSegmentDescription(
		expression,
		fnAllDescription,
		fnGetSingleItemDescription,
		fnGetIntervalDescriptionFormat,
		fnGetBetweenDescriptionFormat,
		fnGetDescriptionFormat)
}

func getMinutesDescription(entity *cronEntity) string {
	expression := entity.Minutes
	fnAllDescription := func() string {
		return "every minute"
	}
	fnGetSingleItemDescription := func(s string) string {
		return s
	}
	fnGetIntervalDescriptionFormat := func(format, _ string) string {
		return fmt.Sprintf("every %s minutes", format)
	}
	fnGetBetweenDescriptionFormat := func(_ string, s ...string) string {
		maxParams := 2
		objectList := make([]interface{}, 0)
		for _, val := range s {
			objectList = append(objectList, val)
			if len(objectList) >= maxParams {
				break
			}
		}
		return fmt.Sprintf("minutes %s through %s past the hour", objectList...)
	}
	fnGetDescriptionFormat := func(_, s string) string {
		if s == "0" {
			return ""
		} else {
			return fmt.Sprintf("at %s minutes past the hour", s)
		}
	}

	return getSegmentDescription(
		expression,
		fnAllDescription,
		fnGetSingleItemDescription,
		fnGetIntervalDescriptionFormat,
		fnGetBetweenDescriptionFormat,
		fnGetDescriptionFormat)
}

func getHoursDescription(entity *cronEntity) string {
	expression := entity.Hours

	fnAllDescription := func() string {
		return "every hour"
	}
	fnGetSingleItemDescription := func(s string) string {
		hourStr, err := formatTime(s, "0", "")
		if err != nil {
			hourStr = ""
		}
		return hourStr
	}
	fnGetIntervalDescriptionFormat := func(format, _ string) string {
		return fmt.Sprintf("every %s hours", format)
	}
	fnGetBetweenDescriptionFormat := func(_ string, s ...string) string {
		maxParams := 2
		objectList := make([]interface{}, 0)
		for _, val := range s {
			objectList = append(objectList, val)
			if len(objectList) >= maxParams {
				break
			}
		}
		return fmt.Sprintf("between %s and %s", objectList...)
	}
	fnGetDescriptionFormat := func(_, s string) string {
		return fmt.Sprintf("at %s", s)
	}

	return getSegmentDescription(
		expression,
		fnAllDescription,
		fnGetSingleItemDescription,
		fnGetIntervalDescriptionFormat,
		fnGetBetweenDescriptionFormat,
		fnGetDescriptionFormat)
}

func getDayOfWeekDescription(entity *cronEntity) string {
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
		return numberToDay(expNum)
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

			formated = ", on the " + dayOfWeekOfMonthDescription + " %s of the month"

		} else if strings.Contains(s, "L") {
			formated = ", on the last %s of the month"

		} else {
			formated = ", only on %s"
		}
		return formated
	}

	fnAllDescription := func() string {
		return ", every day"
	}
	fnGetSingleItemDescription := func(s string) string {
		return getDayName(s)
	}
	fnGetIntervalDescriptionFormat := func(format, _ string) string {
		return fmt.Sprintf(", every %s days of the week", format)
	}
	fnGetBetweenDescriptionFormat := func(_ string, s ...string) string {
		maxParams := 2
		objectList := make([]interface{}, 0)
		for _, val := range s {
			objectList = append(objectList, val)
			if len(objectList) >= maxParams {
				break
			}
		}
		return fmt.Sprintf(", %s through %s", objectList...)
	}
	fnGetDescriptionFormat := func(format, s string) string {
		return fmt.Sprintf(getFormat(format), s)
	}

	return getSegmentDescription(
		entity.DayOfWeek,
		fnAllDescription,
		fnGetSingleItemDescription,
		fnGetIntervalDescriptionFormat,
		fnGetBetweenDescriptionFormat,
		fnGetDescriptionFormat)
}

func getMonthDescription(entity *cronEntity) string {
	fnAllDescription := func() string {
		return ""
	}
	fnGetSingleItemDescription := func(s string) string {
		month, err := strconv.Atoi(s)
		if err != nil || month < 1 || month > 12 {
			return ""
		}
		return MonthName[month]
	}
	fnGetIntervalDescriptionFormat := func(format, _ string) string {
		return fmt.Sprintf(", every %s months", format)
	}
	fnGetBetweenDescriptionFormat := func(_ string, s ...string) string {
		maxParams := 2
		objectList := make([]interface{}, 0)
		for _, val := range s {
			objectList = append(objectList, val)
			if len(objectList) >= maxParams {
				break
			}
		}
		return fmt.Sprintf(", %s through %s", objectList...)
	}
	fnGetDescriptionFormat := func(_, s string) string {
		return fmt.Sprintf(", only in %s", s)
	}

	return getSegmentDescription(
		entity.Month,
		fnAllDescription,
		fnGetSingleItemDescription,
		fnGetIntervalDescriptionFormat,
		fnGetBetweenDescriptionFormat,
		fnGetDescriptionFormat)
}

func getDayOfMonthDescription(entity *cronEntity) string {
	expression := entity.DayOfMonth
	expression = strings.Replace(expression, "?", "*", -1)
	description := ""

	if expression == "L" {
		description = ", on the last day of the month"

	} else if expression == "LW" || expression == "WL" {
		description = ", on the last weekday of the month"

	} else {
		regexRule := regexp.MustCompile(`(\d{1,2}W)|(W\d{1,2})`)
		if regexRule.MatchString(expression) {
			dayString := ""
			dayNum, err := strconv.Atoi(strings.Replace(expression, "W", "", -1))
			if err == nil {
				if dayNum == 1 {
					dayString = "first weekday"
				} else {
					dayString = fmt.Sprintf("weekday nearest day %s", strconv.Itoa(dayNum))
				}
			}
			description = fmt.Sprintf(", on the %s of the month", dayString)

		} else {
			fnAllDescription := func() string {
				return ", every day"
			}
			fnGetSingleItemDescription := func(s string) string {
				return s
			}
			fnGetIntervalDescriptionFormat := func(format, s string) string {
				if format == "1" {
					return ", every day"
				} else {
					return fmt.Sprintf(", every %s days", s)
				}
			}
			fnGetBetweenDescriptionFormat := func(_ string, s ...string) string {
				maxParams := 2
				objectList := make([]interface{}, 0)
				for _, val := range s {
					objectList = append(objectList, val)
					if len(objectList) >= maxParams {
						break
					}
				}
				return fmt.Sprintf(", between day %s and %s of the month", objectList...)
			}
			fnGetDescriptionFormat := func(_, s string) string {
				return fmt.Sprintf(", on day %s of the month", s)
			}
			description = getSegmentDescription(
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

func getYearDescription(entity *cronEntity) string {
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

	fnAllDescription := func() string {
		return ""
	}
	fnGetSingleItemDescription := func(s string) string {
		return formatYear(s)
	}
	fnGetIntervalDescriptionFormat := func(format, _ string) string {
		return fmt.Sprintf(", every %s years", format)
	}
	fnGetBetweenDescriptionFormat := func(_ string, s ...string) string {
		maxParams := 2
		objectList := make([]interface{}, 0)
		for _, val := range s {
			objectList = append(objectList, val)
			if len(objectList) >= maxParams {
				break
			}
		}
		return fmt.Sprintf(", %s through %s", objectList...)
	}
	fnGetDescriptionFormat := func(_, s string) string {
		return fmt.Sprintf(", only in %s", s)
	}

	return getSegmentDescription(
		entity.Year,
		fnAllDescription,
		fnGetSingleItemDescription,
		fnGetIntervalDescriptionFormat,
		fnGetBetweenDescriptionFormat,
		fnGetDescriptionFormat)
}

func getSegmentDescription(
	expression string,
	fnAllDescription func() string,
	fnGetSingleItemDescription func(s string) string,
	fnGetIntervalDescriptionFormat func(format, s string) string,
	fnGetBetweenDescriptionFormat func(format string, s ... string) string,
	fnGetDescriptionFormat func(format, s string) string,
) string {

	description := ""
	if expression == "" {

	} else if expression == "*" {
		description = fnAllDescription()

	} else if !strings.ContainsAny(expression, "/-,") {
		description = fnGetDescriptionFormat(expression, fnGetSingleItemDescription(expression))

	} else if strings.Contains(expression, "/") {
		segments := strings.Split(expression, "/")
		description = fnGetIntervalDescriptionFormat(segments[1], fnGetSingleItemDescription(segments[1]))

		//interval contains 'between' piece (i.e. 2-59/3 )
		if strings.Contains(segments[0], "-") {
			betweenSegmentDescription := generateBetweenSegmentDescription(segments[0], fnGetBetweenDescriptionFormat, fnGetSingleItemDescription)
			if !regexp.MustCompile(`^(,\s)`).MatchString(betweenSegmentDescription) {
				// not start with ", "
				description += ", "
			}
			description += betweenSegmentDescription

		} else if !strings.ContainsAny(segments[0], "*,") {
			rangeItemDescription := fnGetDescriptionFormat(segments[0], fnGetSingleItemDescription(segments[0]))
			rangeItemDescription = strings.Replace(rangeItemDescription, ", ", "", -1)
			description += fmt.Sprintf(", starting %s", rangeItemDescription)
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
				descriptionContent += " and "
			}

			if strings.Contains(segment, "-") {
				betweenDescription := generateBetweenSegmentDescription(
					segment,
					func(_ string, s ...string) string {
						maxParams := 2
						objectList := make([]interface{}, 0)
						for _, val := range s {
							objectList = append(objectList, val)
							if len(objectList) >= maxParams {
								break
							}
						}
						return fmt.Sprintf(", %s through %s", objectList...)
					},
					fnGetSingleItemDescription)
				betweenDescription = strings.Replace(betweenDescription, ", ", "", -1)
				descriptionContent += betweenDescription

			} else {
				descriptionContent += fnGetSingleItemDescription(segment)
			}
		}

		description = fnGetDescriptionFormat(expression, descriptionContent)

	} else if strings.Contains(expression, "-") {
		description = generateBetweenSegmentDescription(expression, fnGetBetweenDescriptionFormat, fnGetSingleItemDescription)
	}

	//fmt.Println("description:", description)
	return description
}

func generateBetweenSegmentDescription(
	betweenExpression string,
	fnGetBetweenDescritionFormat func(format string, s ...string) string,
	fnGetSingleItemDescription func(s string) string,
) string {
	description := ""
	betweenSegments := strings.Split(betweenExpression, "-")
	betweenSegment1Description := fnGetSingleItemDescription(betweenSegments[0])
	betweenSegment2Description := fnGetSingleItemDescription(betweenSegments[1])
	betweenSegment2Description = strings.Replace(betweenSegment2Description, ":00", ":59", -1)

	description += fnGetBetweenDescritionFormat(betweenExpression, betweenSegment1Description, betweenSegment2Description)
	return description
}

func formatTime(hourExp, minuteExp, secondsExp string) (string, error) {
	hour, err := strconv.Atoi(hourExp)
	if err != nil {
		return "", err
	}
	if hour < 0 || hour > 24 {
		return "", errors.New("error hours")
	}

	period := ""
	if !Use24hourTimeFormat {
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

func transformVerbosity(description string, useVerboseFormat bool) string {
	if !useVerboseFormat {
		description = strings.Replace(description, ", every minute", "", -1)
		description = strings.Replace(description, ", every hour", "", -1)
		description = strings.Replace(description, ", every day", "", -1)
	}
	return description
}

func transformCase(description string, caseType int) string {
	switch caseType {
	case CasingSentence:
		for i, v := range description {
			description = string(unicode.ToUpper(v)) + description[i+1:]
			break
		}

	case CasingTitle:
		description = strings.Title(description)

	default:
		description = strings.ToLower(description)
	}
	return description
}

func numberToDay(dayNumber int) string {
	if dayNumber < 0 || dayNumber >= len(WeekDayName) {
		return ""
	}
	return WeekDayName[dayNumber]
}
