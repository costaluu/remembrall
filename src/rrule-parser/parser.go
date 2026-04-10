// Package rruletext parses natural-language recurrence descriptions (English)
// into rrule.ROption, compatible with github.com/teambition/rrule-go.
//
// Example:
//
//	opts, err := rruletext.ParseText("every week on monday, wednesday at 9")
//	rule, err := rrule.NewRRule(*opts)
package rruleparser

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	rrule "github.com/teambition/rrule-go"
)

// =============================================================================
// Regex cache
// =============================================================================

var (
	muRe             sync.RWMutex
	compiledPatterns = map[string]*regexp.Regexp{}
)

func mustCompile(pattern string) *regexp.Regexp {
	muRe.RLock()
	r, ok := compiledPatterns[pattern]
	muRe.RUnlock()
	if ok {
		return r
	}
	r = regexp.MustCompile(pattern)
	muRe.Lock()
	compiledPatterns[pattern] = r
	muRe.Unlock()
	return r
}

// findMatch returns [fullmatch, group1, group2, ...] or nil.
func findMatch(pattern, text string) []string {
	return mustCompile(pattern).FindStringSubmatch(text)
}

// =============================================================================
// Token definitions (English)
// =============================================================================

type token struct {
	name    string
	pattern string
}

// englishTokens lists all lexer tokens. Longer matches take priority.
var englishTokens = []token{
	{name: "SKIP", pattern: `^[ \t\r\n]+`},
	{name: "every", pattern: `^every`},
	// "nth" must appear before "number" so "3rd" is matched as nth, not number
	{name: "nth", pattern: `^(\+|-)?(\d+)(st|nd|rd|th)`},
	{name: "number", pattern: `^\d+`},
	{name: "day(s)", pattern: `^days?`},
	{name: "weekday(s)", pattern: `^weekdays?`},
	{name: "week(s)", pattern: `^weeks?`},
	{name: "hour(s)", pattern: `^hours?`},
	{name: "minute(s)", pattern: `^minutes?`},
	{name: "month(s)", pattern: `^months?`},
	{name: "year(s)", pattern: `^years?`},
	{name: "on", pattern: `^on`},
	{name: "at", pattern: `^at`},
	{name: "the", pattern: `^the`},
	{name: "first", pattern: `^first`},
	{name: "second", pattern: `^second`},
	{name: "third", pattern: `^third`},
	{name: "last", pattern: `^last`},
	{name: "for", pattern: `^for`},
	{name: "until", pattern: `^until`},
	{name: "comma", pattern: `^,`},
	{name: "monday", pattern: `^(monday|mon)`},
	{name: "tuesday", pattern: `^(tuesday|tue)`},
	{name: "wednesday", pattern: `^(wednesday|wed)`},
	{name: "thursday", pattern: `^(thursday|thu)`},
	{name: "friday", pattern: `^(friday|fri)`},
	{name: "saturday", pattern: `^(saturday|sat)`},
	{name: "sunday", pattern: `^(sunday|sun)`},
	{name: "january", pattern: `^(january|jan)`},
	{name: "february", pattern: `^(february|feb)`},
	{name: "march", pattern: `^(march|mar)`},
	{name: "april", pattern: `^(april|apr)`},
	{name: "may", pattern: `^may`},
	{name: "june", pattern: `^(june|jun)`},
	{name: "july", pattern: `^(july|jul)`},
	{name: "august", pattern: `^(august|aug)`},
	{name: "september", pattern: `^(september|sep)`},
	{name: "october", pattern: `^(october|oct)`},
	{name: "november", pattern: `^(november|nov)`},
	{name: "december", pattern: `^(december|dec)`},
}

// =============================================================================
// Parser / lexer
// =============================================================================

type matchResult struct {
	name    string
	matched string   // full matched text
	groups  []string // capture groups starting at index 1
}

type parser struct {
	tokens []token
	text   string
	symbol *matchResult
	done   bool
}

func newParser(tokens []token) *parser {
	return &parser{tokens: tokens}
}

// start initialises the parser and advances to the first non-SKIP symbol.
func (p *parser) start(text string) bool {
	p.text = strings.ToLower(strings.TrimSpace(text))
	p.done = false
	return p.nextSymbol()
}

// isDone returns true when no more symbols are available.
func (p *parser) isDone() bool {
	return p.done && p.symbol == nil
}

// nextSymbol advances to the next non-SKIP symbol.
func (p *parser) nextSymbol() bool {
	p.symbol = nil

	for {
		if p.done {
			return false
		}

		var best *matchResult
		for _, tok := range p.tokens {
			m := findMatch(tok.pattern, p.text)
			if m == nil {
				continue
			}
			if best == nil || len(m[0]) > len(best.matched) {
				best = &matchResult{
					name:    tok.name,
					matched: m[0],
					groups:  m[1:],
				}
			}
		}

		if best == nil {
			p.done = true
			p.symbol = nil
			return false
		}

		p.text = p.text[len(best.matched):]
		if p.text == "" {
			p.done = true
		}

		if best.name == "SKIP" {
			continue
		}

		p.symbol = best
		return true
	}
}

// accept consumes and returns the current symbol if its name matches.
func (p *parser) accept(name string) *matchResult {
	if p.symbol != nil && p.symbol.name == name {
		v := p.symbol
		p.nextSymbol()
		return v
	}
	return nil
}

// acceptNumber is shorthand for accept("number").
func (p *parser) acceptNumber() *matchResult {
	return p.accept("number")
}

// expect consumes the expected symbol or returns an error.
func (p *parser) expect(name string) error {
	if p.accept(name) != nil {
		return nil
	}
	got := "<nil>"
	if p.symbol != nil {
		got = p.symbol.name
	}
	return errors.New("expected " + name + " but found " + got)
}

// sym returns the name of the current symbol (empty string if none).
func (p *parser) sym() string {
	if p.symbol == nil {
		return ""
	}
	return p.symbol.name
}

// =============================================================================
// ParseText – public entry point
// =============================================================================

// ParseText parses a natural-language recurrence description (English) and
// returns an *rrule.ROption ready to be passed to rrule.NewRRule.
// Returns (nil, nil) when the input is empty or cannot be parsed.
func ParseText(text string) (*rrule.ROption, error) {
	opts := &rrule.ROption{}
	ttr := newParser(englishTokens)

	if !ttr.start(text) {
		return nil, nil
	}

	if err := parseS(ttr, opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// =============================================================================
// Grammar functions
// =============================================================================

// parseS parses the top-level "every [n] <freq> ..." production.
func parseS(ttr *parser, opts *rrule.ROption) error {
	if err := ttr.expect("every"); err != nil {
		return err
	}

	if n := ttr.acceptNumber(); n != nil {
		v, _ := strconv.Atoi(n.matched)
		opts.Interval = v
	}

	if ttr.isDone() {
		return errors.New("unexpected end")
	}

	switch ttr.sym() {
	// ------------------------------------------------------------------ daily
	case "day(s)":
		opts.Freq = rrule.DAILY
		if ttr.nextSymbol() {
			if err := parseAT(ttr, opts); err != nil {
				return err
			}
			return parseF(ttr, opts)
		}

	// --------------------------------------------------------------- weekdays
	case "weekday(s)":
		opts.Freq = rrule.WEEKLY
		opts.Byweekday = []rrule.Weekday{rrule.MO, rrule.TU, rrule.WE, rrule.TH, rrule.FR}
		ttr.nextSymbol()
		if err := parseAT(ttr, opts); err != nil {
			return err
		}
		return parseF(ttr, opts)

	// --------------------------------------------------------------- weekly
	case "week(s)":
		opts.Freq = rrule.WEEKLY
		if ttr.nextSymbol() {
			if err := parseON(ttr, opts); err != nil {
				return err
			}
			if err := parseAT(ttr, opts); err != nil {
				return err
			}
			return parseF(ttr, opts)
		}

	// --------------------------------------------------------------- hourly
	case "hour(s)":
		opts.Freq = rrule.HOURLY
		if ttr.nextSymbol() {
			if err := parseON(ttr, opts); err != nil {
				return err
			}
			return parseF(ttr, opts)
		}

	// -------------------------------------------------------------- minutely
	case "minute(s)":
		opts.Freq = rrule.MINUTELY
		if ttr.nextSymbol() {
			if err := parseON(ttr, opts); err != nil {
				return err
			}
			return parseF(ttr, opts)
		}

	// --------------------------------------------------------------- monthly
	case "month(s)":
		opts.Freq = rrule.MONTHLY
		if ttr.nextSymbol() {
			if err := parseON(ttr, opts); err != nil {
				return err
			}
			return parseF(ttr, opts)
		}

	// --------------------------------------------------------------- yearly
	case "year(s)":
		opts.Freq = rrule.YEARLY
		if ttr.nextSymbol() {
			if err := parseON(ttr, opts); err != nil {
				return err
			}
			return parseF(ttr, opts)
		}

	// --------------------------------------------------- named weekday(s)
	case "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday":
		opts.Freq = rrule.WEEKLY
		wkd, ok := weekdayFromSymbol(ttr.sym())
		if !ok {
			return errors.New("unknown weekday: " + ttr.sym())
		}
		opts.Byweekday = []rrule.Weekday{wkd}

		if !ttr.nextSymbol() {
			return nil
		}

		for ttr.accept("comma") != nil {
			if ttr.isDone() {
				return errors.New("unexpected end")
			}
			w, err := decodeWKD(ttr)
			if err != nil {
				return errors.New("unexpected symbol " + ttr.sym() + ", expected weekday")
			}
			opts.Byweekday = append(opts.Byweekday, w)
			ttr.nextSymbol()
		}

		if err := parseAT(ttr, opts); err != nil {
			return err
		}
		if err := parseMDAYs(ttr, opts); err != nil {
			return err
		}
		return parseF(ttr, opts)

	// --------------------------------------------------- named month(s)
	case "january", "february", "march", "april", "may", "june",
		"july", "august", "september", "october", "november", "december":
		opts.Freq = rrule.YEARLY
		m, ok := monthFromSymbol(ttr.sym())
		if !ok {
			return errors.New("unknown month: " + ttr.sym())
		}
		opts.Bymonth = []int{m}

		if !ttr.nextSymbol() {
			return nil
		}

		for ttr.accept("comma") != nil {
			if ttr.isDone() {
				return errors.New("unexpected end")
			}
			month, err := decodeM(ttr)
			if err != nil {
				return errors.New("unexpected symbol " + ttr.sym() + ", expected month")
			}
			opts.Bymonth = append(opts.Bymonth, month)
			ttr.nextSymbol()
		}

		if err := parseON(ttr, opts); err != nil {
			return err
		}
		return parseF(ttr, opts)

	default:
		return errors.New("unknown symbol: " + ttr.sym())
	}

	return nil
}

// parseON handles the "on [the] ..." clause.
func parseON(ttr *parser, opts *rrule.ROption) error {
	if ttr.accept("on") == nil && ttr.accept("the") == nil {
		return nil
	}

	for {
		nth, nthOk, err := decodeNTH(ttr)
		if err != nil {
			return err
		}

		wkd, wkdErr := decodeWKD(ttr)
		m, mErr := decodeM(ttr)

		switch {
		case nthOk:
			if wkdErr == nil {
				// nth <weekday>
				ttr.nextSymbol()
				opts.Byweekday = append(opts.Byweekday, wkd.Nth(nth))
			} else {
				// nth <day>
				opts.Bymonthday = append(opts.Bymonthday, nth)
				ttr.accept("day(s)")
			}

		case wkdErr == nil:
			ttr.nextSymbol()
			opts.Byweekday = append(opts.Byweekday, wkd)

		case ttr.sym() == "weekday(s)":
			ttr.nextSymbol()
			if opts.Byweekday == nil {
				opts.Byweekday = []rrule.Weekday{rrule.MO, rrule.TU, rrule.WE, rrule.TH, rrule.FR}
			}

		case ttr.sym() == "week(s)":
			ttr.nextSymbol()
			n := ttr.acceptNumber()
			if n == nil {
				return errors.New("unexpected symbol " + ttr.sym() + ", expected week number")
			}
			v, _ := strconv.Atoi(n.matched)
			opts.Byweekno = append(opts.Byweekno, v)
			for ttr.accept("comma") != nil {
				n = ttr.acceptNumber()
				if n == nil {
					return errors.New("unexpected symbol " + ttr.sym() + "; expected week number")
				}
				v, _ = strconv.Atoi(n.matched)
				opts.Byweekno = append(opts.Byweekno, v)
			}

		case mErr == nil:
			ttr.nextSymbol()
			opts.Bymonth = append(opts.Bymonth, m)

		default:
			return nil
		}

		if ttr.accept("comma") == nil && ttr.accept("the") == nil && ttr.accept("on") == nil {
			break
		}
	}
	return nil
}

// parseAT handles the "at <hour>[, <hour>...]" clause.
func parseAT(ttr *parser, opts *rrule.ROption) error {
	if ttr.accept("at") == nil {
		return nil
	}

	for {
		n := ttr.acceptNumber()
		if n == nil {
			return errors.New("unexpected symbol " + ttr.sym() + ", expected hour")
		}
		v, _ := strconv.Atoi(n.matched)
		opts.Byhour = append(opts.Byhour, v)

		for ttr.accept("comma") != nil {
			n = ttr.acceptNumber()
			if n == nil {
				return errors.New("unexpected symbol " + ttr.sym() + "; expected hour")
			}
			v, _ = strconv.Atoi(n.matched)
			opts.Byhour = append(opts.Byhour, v)
		}

		if ttr.accept("comma") == nil && ttr.accept("at") == nil {
			break
		}
	}
	return nil
}

// parseMDAYs handles an optional "on the <nth>[, <nth>...]" clause.
func parseMDAYs(ttr *parser, opts *rrule.ROption) error {
	ttr.accept("on")
	ttr.accept("the")

	nth, ok, err := decodeNTH(ttr)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	opts.Bymonthday = append(opts.Bymonthday, nth)
	ttr.nextSymbol()

	for ttr.accept("comma") != nil {
		nth, ok, err = decodeNTH(ttr)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("unexpected symbol " + ttr.sym() + "; expected monthday")
		}
		opts.Bymonthday = append(opts.Bymonthday, nth)
		ttr.nextSymbol()
	}
	return nil
}

// parseF handles the terminal "until <date>" or "for <n> times" clause.
func parseF(ttr *parser, opts *rrule.ROption) error {
	switch ttr.sym() {
	case "until":
		t, err := parseUntilDate(ttr.text)
		if err != nil {
			return errors.New("cannot parse until date: " + ttr.text)
		}
		opts.Until = t

	case "for":
		ttr.nextSymbol() // consume "for"
		if ttr.symbol == nil {
			return errors.New("expected number after 'for'")
		}
		n, _ := strconv.Atoi(ttr.symbol.matched)
		opts.Count = n
		if err := ttr.expect("number"); err != nil {
			return err
		}
	}
	return nil
}

// parseUntilDate tries several common date layouts.
func parseUntilDate(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02",
		"01/02/2006",
		"January 2, 2006",
		"Jan 2, 2006",
		"02 Jan 2006",
		"2 January 2006",
	}
	s = strings.TrimSpace(s)
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("unrecognised date format")
}

// =============================================================================
// Symbol decoders
// =============================================================================

// decodeNTH tries to decode ordinal words ("first", "second", "third", "last",
// or a numeric nth like "3rd"). Returns (value, true, nil) on success.
func decodeNTH(ttr *parser) (int, bool, error) {
	switch ttr.sym() {
	case "last":
		ttr.nextSymbol()
		return -1, true, nil

	case "first":
		ttr.nextSymbol()
		return 1, true, nil

	case "second":
		ttr.nextSymbol()
		if ttr.accept("last") != nil {
			return -2, true, nil
		}
		return 2, true, nil

	case "third":
		ttr.nextSymbol()
		if ttr.accept("last") != nil {
			return -3, true, nil
		}
		return 3, true, nil

	case "nth":
		// groups[0]=sign, groups[1]=digits, groups[2]=suffix (st/nd/rd/th)
		sign := ttr.symbol.groups[0]
		digits := ttr.symbol.groups[1]
		v, err := strconv.Atoi(sign + digits)
		if err != nil {
			return 0, false, err
		}
		if v < -366 || v > 366 {
			return 0, false, errors.New("nth out of range: " + strconv.Itoa(v))
		}
		ttr.nextSymbol()
		if ttr.accept("last") != nil {
			return -v, true, nil
		}
		return v, true, nil
	}
	return 0, false, nil
}

// decodeWKD returns the Weekday for the current symbol, or an error.
func decodeWKD(ttr *parser) (rrule.Weekday, error) {
	wkd, ok := weekdayFromSymbol(ttr.sym())
	if !ok {
		return rrule.Weekday{}, errors.New("expected weekday, got " + ttr.sym())
	}
	return wkd, nil
}

// decodeM returns the month number (1-12) for the current symbol, or an error.
func decodeM(ttr *parser) (int, error) {
	m, ok := monthFromSymbol(ttr.sym())
	if !ok {
		return 0, errors.New("expected month, got " + ttr.sym())
	}
	return m, nil
}

// =============================================================================
// Lookup tables
// =============================================================================

func weekdayFromSymbol(sym string) (rrule.Weekday, bool) {
	switch sym {
	case "monday":
		return rrule.MO, true
	case "tuesday":
		return rrule.TU, true
	case "wednesday":
		return rrule.WE, true
	case "thursday":
		return rrule.TH, true
	case "friday":
		return rrule.FR, true
	case "saturday":
		return rrule.SA, true
	case "sunday":
		return rrule.SU, true
	}
	return rrule.Weekday{}, false
}

func monthFromSymbol(sym string) (int, bool) {
	switch sym {
	case "january":
		return 1, true
	case "february":
		return 2, true
	case "march":
		return 3, true
	case "april":
		return 4, true
	case "may":
		return 5, true
	case "june":
		return 6, true
	case "july":
		return 7, true
	case "august":
		return 8, true
	case "september":
		return 9, true
	case "october":
		return 10, true
	case "november":
		return 11, true
	case "december":
		return 12, true
	}
	return 0, false
}
