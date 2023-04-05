package main

import (
	"regexp"
	"testing"
)



func Test_regexp(t *testing.T) {
	pattern := `(?i)^(Node|BIKE|CAR|CAR_PREMIUM|CYCLE|HELICOPTER|POOL|FF4W|SHARE|TRIKE)$`
	match, err := regexp.Match(pattern, []byte("taxi"))
	t.Log(match)
	t.Log(err)
}
