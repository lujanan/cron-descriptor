package main

import (
	"fmt"
	"testing"
)

func TestGetDescription(t *testing.T) {
	cron := "0 15 10 * 6L 2002-2006"
	descriptionType := DescFull

	desc := GetDescription(cron, descriptionType)
	fmt.Println(desc)

	cronList := []string{
		"0 1 */4 * * *",
		"0/2 * * * * ?",
		"0 0/2 * * * ?",
		"0 0 2 1 * ?",
		"0 15 10 ? * MON-FRI",
		"0 0 10,14,16 * * ?",
		"0 0/30 9-17 * * ?",
		"0 0 12 ? * WED",
		"0 0 12 * * ?",
		"0 15 10 ? * *",
		"0 15 10 * * ?",
		"0 15 10 * * ? 2005",
		"0 * 14 * * ?",
		"0 0/5 14 * * ?",
		"0 0/5 14,18 * * ?",
		"0 0-5 14 * * ?",
		"0 10,44 14 ? 3 WED",
		"0 15 10 ? * MON-FRI",
		"0 15 10 15 * ?",
		"0 15 10 L * ?",
		"0 15 10 ? * 6L",
		"0 15 10 ? * 6L 2002-2005",
		"0 15 10 ? * 6#3",
	}

	for _, val := range cronList {
		desc := GetDescription(val, descriptionType)
		fmt.Printf("%s:: \n %s \n\n", val, desc)
	}
}
