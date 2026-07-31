package utils

import (
	"encoding/base32"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

const salt = "6rx7aXuTnVtM1ZMXOgJNHX9cmibQs3vAOApvT9KPQev+v2Sa5G9QnH+E483CwmTi+gBCHZWTYLN8EMCjX94B3RPMh44WQUCS/24o75Tr97sLzz5z5ZLtFbQQDtvLgj2n"

var b32encoder = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

func Encode(s string) string {
	if s == "" {
		return ""
	}
	if err := validate(s); err != nil {
		panic(err)
	}
	return b32encoder.EncodeToString(mangle([]byte(toAbbr(s))))
}

func Decode(s string) string {
	if s == "" {
		return ""
	}
	d, err := b32encoder.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return fromAbbr(string(demangle(d)))
}

func DecodeUnsafe(s string) (decoded string) {
	if r := recover(); r != nil {
		return
	}

	if s == "" {
		return ""
	}
	d, err := b32encoder.DecodeString(s)
	if err != nil || len(d) == 0 {
		return
	}
	v := fromAbbr(string(demangle(d)))
	if isValidRaw(v) {
		return v
	} else {
		return ""
	}
}

func isValidRaw(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9',
			c >= 'a' && c <= 'z',
			c == '-', c == '_', c == ':', c == '=', c == '@', c == '.', c == '|':
			continue
		default:
			return false
		}
	}
	return true
}

// validate 对接时校验, 后续移除
func validate(i string) error {
	if strings.Contains(i, ":") {
		return nil
	}
	// FIXME: ao outcomeId, e.g. 1-2
	if strings.Contains(i, "-") {
		return nil
	}
	if _, err := strconv.Atoi(i); err == nil {
		return nil
	}
	return fmt.Errorf("invalid raw id: %s", i)
}

func pivot(s string) string {
	n := len(s)
	m := n / 2
	m1 := (n + 1) / 2
	return s[m1:n] + s[m:m1] + s[0:m]
}

func mangle(s []byte) []byte {
	n := len(s)
	b := make([]byte, n)
	lastPos := n - 1
	lastByte := s[lastPos]
	for i := 0; i < lastPos; i++ {
		b[i] = s[i] ^ lastByte ^ salt[i]
	}
	b[lastPos] = s[lastPos]
	return b
}

func demangle(s []byte) []byte {
	return mangle(s)
}

var WellknownAbbreviations = map[string]string{
	"sr:match:":             "{1}",
	"sr:competitor:":        "(2)", // FIXME: typo
	"sr:tournament:":        "{3}",
	"sr:simple_tournament:": "{4}",
	"sr:stage:":             "{5}",
	"sr:season:":            "{6}",
	"sr:category:":          "{7}",
	"sr:player:":            "{8}",
	"sr:venue:":             "{9}",
	"sr:sport_event:":       "{10}",
	"sr:race_event:":        "{11}",
	"sr:race_tournament:":   "{12}",
	"sr:h2h_tournament:":    "{13}",
	"sr:outright:":          "{14}",
	"sr:sport:":             "{15}",
	"sr:team:":              "{16}",
	"sr:simpleteam:":        "{17}",
	"sr:simple_team:":       "{18}",
	"sr:referee:":           "{19}",
	"sr:market:":            "{20}",
	"sr:lottery:":           "{21}",
	"sr:draw:":              "{22}",
	"sr:competition_group:": "{23}",

	"codds:competition_group:": "{24}",

	"mkt:":                               "{81}",
	"variant=":                           "{82}",
	"pre:markettext:":                    "{83}",
	"sr:winning_margin_no_draw:":         "{84}",
	"sr:exact_goals:":                    "{85}",
	"sr:decided_by_extra_points:bestof:": "{86}",
	"sr:point_range:":                    "{87}",
	"sr:winning_margin:":                 "{88}",
	"sr:correct_score:bestof:":           "{89}",
	"sr:correct_score:below:":            "{90}",
	"sr:goalscorer:fieldplayers_nogoal_owngoal_other|": "{91}",
	"sr:winner_and_rounds:":                            "{92}",
	"sr:winning_method:":                               "{93}",

	"ts:match:":      "[1]",
	"ts:competitor:": "[2]",
	"ts:tournament:": "[3]",
	"ts:utour:":      "[4]",
	"ts:season:":     "[5]",
	"ts:player:":     "[6]",

	"ao:match:":      "<1>",
	"ao:competitor:": "<2>",
	"ao:tournament:": "<3>",
	"ao:utour:":      "<4>",
	"ao:season:":     "<5>",
	"ao:player:":     "<6>",
}

var (
	toAbbrReplacer   *strings.Replacer
	fromAbbrReplacer *strings.Replacer
	once             sync.Once
)

func initReplacers() {
	once.Do(func() {
		toAbbrValues := make([]string, 0, len(WellknownAbbreviations)*2)
		fromAbbrValues := make([]string, 0, len(WellknownAbbreviations)*2)
		for k, v := range WellknownAbbreviations {
			toAbbrValues = append(toAbbrValues, k, v)
			fromAbbrValues = append(fromAbbrValues, v, k)
		}
		toAbbrReplacer = strings.NewReplacer(toAbbrValues...)
		fromAbbrReplacer = strings.NewReplacer(fromAbbrValues...)
	})
}

func toAbbr(s string) string {
	initReplacers()
	s = toAbbrReplacer.Replace(s)
	return s
}

func fromAbbr(s string) string {
	initReplacers()
	s = fromAbbrReplacer.Replace(s)
	return s
}
