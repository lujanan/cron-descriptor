package main

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	CronDays = map[int]string{
		0: "SUN",
		1: "MON",
		2: "TUE",
		3: "WED",
		4: "THU",
		5: "FRI",
		6: "SAT",
	}

	CronMonths = map[int]string{
		1:  "JAN",
		2:  "FEB",
		3:  "MAR",
		4:  "APR",
		5:  "MAY",
		6:  "JUN",
		7:  "JUL",
		8:  "AUG",
		9:  "SEP",
		10: "OCT",
		11: "NOV",
		12: "DEC",
	}
)

func parse(desc *descriptor) (*cronEntity, error) {
	entity := &cronEntity{}
	if desc.Expression == "" {
		return nil, errors.New("expression is empty")
	}

	expressionPartsTemp := strings.Split(desc.Expression, " ")
	expressionPartsTempLength := len(expressionPartsTemp)
	if expressionPartsTempLength < 5 {
		return nil, errors.New("expression part less than 5")

	} else if expressionPartsTempLength == 5 {
		//5 part cron so shift array past seconds element
		entity = &cronEntity{
			Minutes:    expressionPartsTemp[0],
			Hours:      expressionPartsTemp[1],
			DayOfMonth: expressionPartsTemp[2],
			Month:      expressionPartsTemp[3],
			DayOfWeek:  expressionPartsTemp[4],
		}

	} else if expressionPartsTempLength == 6 {
		//If last element ends with 4 digits, a year element has been
		//supplied and no seconds element
		yearRegexp := regexp.MustCompile(`\d{4}$`)
		if yearRegexp.MatchString(expressionPartsTemp[5]) {
			entity = &cronEntity{
				Minutes:    expressionPartsTemp[0],
				Hours:      expressionPartsTemp[1],
				DayOfMonth: expressionPartsTemp[2],
				Month:      expressionPartsTemp[3],
				DayOfWeek:  expressionPartsTemp[4],
				Year:       expressionPartsTemp[5],
			}

		} else {
			entity = &cronEntity{
				Seconds:    expressionPartsTemp[0],
				Minutes:    expressionPartsTemp[1],
				Hours:      expressionPartsTemp[2],
				DayOfMonth: expressionPartsTemp[3],
				Month:      expressionPartsTemp[4],
				DayOfWeek:  expressionPartsTemp[5],
			}
		}

	} else if expressionPartsTempLength == 7 {
		entity = &cronEntity{
			Seconds:    expressionPartsTemp[0],
			Minutes:    expressionPartsTemp[1],
			Hours:      expressionPartsTemp[2],
			DayOfMonth: expressionPartsTemp[3],
			Month:      expressionPartsTemp[4],
			DayOfWeek:  expressionPartsTemp[5],
			Year:       expressionPartsTemp[6],
		}

	} else {
		return nil, errors.New("expression part more than 7")
	}

	return normalizeExpression(entity, desc.Options), nil
}

func normalizeExpression(entity *cronEntity, opts *options) *cronEntity {
	//convert ? to * only for DOM and DOW
	entity.DayOfMonth = strings.Replace(entity.DayOfMonth, "?", "*", -1)
	entity.DayOfWeek = strings.Replace(entity.DayOfWeek, "?", "*", -1)

	//convert 0/, 1/ to */
	if regexp.MustCompile(`^0/`).MatchString(entity.Seconds) {
		//seconds
		entity.Seconds = strings.Replace(entity.Seconds, "0/", "*/", 1)
	}

	if regexp.MustCompile(`^0/`).MatchString(entity.Minutes) {
		//minutes
		entity.Minutes = strings.Replace(entity.Minutes, "0/", "*/", 1)
	}

	if regexp.MustCompile(`^0/`).MatchString(entity.Hours) {
		//hours
		entity.Hours = strings.Replace(entity.Hours, "0/", "*/", 1)
	}

	if regexp.MustCompile(`^1/`).MatchString(entity.DayOfMonth) {
		//DOM
		entity.DayOfMonth = strings.Replace(entity.DayOfMonth, "1/", "*/", 1)
	}

	if regexp.MustCompile(`^1/`).MatchString(entity.Month) {
		//DescMonth
		entity.Month = strings.Replace(entity.Month, "1/", "*/", 1)
	}

	if regexp.MustCompile(`^1/`).MatchString(entity.DayOfWeek) {
		//DOW
		entity.DayOfWeek = strings.Replace(entity.DayOfWeek, "1/", "*/", 1)
	}

	if regexp.MustCompile(`^1/`).MatchString(entity.Year) {
		//year
		entity.Year = strings.Replace(entity.Year, "1/", "*/", 1)
	}

	//handle DayOfWeekStartIndexZero option where SUN=1 rather than SUN=0
	if !opts.DayOfWeekStartIndexZero {
		entity.DayOfWeek = decreaseDaysOfWeek(entity.DayOfWeek)
	}

	//convert SUN-SAT format to 0-6 format
	for dayNumber, day := range CronDays {
		entity.DayOfWeek = strings.Replace(strings.ToUpper(entity.DayOfWeek), day, strconv.Itoa(dayNumber), -1)
	}

	//convert JAN-DEC format to 1-12 format
	for monthNumber, month := range CronMonths {
		entity.Month = strings.Replace(strings.ToUpper(entity.Month), month, strconv.Itoa(monthNumber), -1)
	}

	//convert 0 second to (empty)
	if entity.Seconds == "0" {
		entity.Seconds = ""
	}

	//Loop through all parts and apply global normalization
	entityValue := reflect.ValueOf(entity).Elem()
	for i := 0; i < entityValue.NumField(); i++ {
		//convert all '*/1' to '*'
		if entityValue.Field(i).String() == "*/1" {
			entityValue.Field(i).SetString("*")
		}

		/**
			Convert DescMonth,DOW,DescYear step values with a starting value (i.e. not '*') to between expressions.
            This allows us to reuse the between expression handling for step values.

            For Example:
            - month part '3/2' will be converted to '3-12/2' (every 2 months between March and December)
            - DOW part '3/2' will be converted to '3-6/2' (every 2 days between Tuesday and Saturday)
		 */

		if strings.Contains(entityValue.Field(i).String(), "/") &&
			!strings.ContainsAny(entityValue.Field(i).String(), "*-,") {

			choices := map[int]string{
				4: "12",
				5: "6",
				6: "9999",
			}
			if stepRangeThrough, ok := choices[i]; ok {
				parts := strings.Split(entityValue.Field(i).String(), "/")
				entityValue.Field(i).SetString(fmt.Sprintf("%s-%s/%s", parts[0], stepRangeThrough, parts[1]))
			}
		}
	}

	return entity
}

func decreaseDaysOfWeek(dayOfWeekExpressionPart string) string {
	dowChars := make([]string, 0)
	for i, dowCharACSII := range dayOfWeekExpressionPart {
		char := string(dowCharACSII)
		dowChars[i] = char
		if i == 0 || dowChars[i-1] != "#" && dowChars[i-1] != "/" {
			if num, err := strconv.Atoi(char); err == nil {
				dowChars[i] = strconv.Itoa(num - 1)
			}
		}
	}
	return strings.Join(dowChars, "")
}
