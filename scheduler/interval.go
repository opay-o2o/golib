package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Interval struct {
	minutes  map[int]bool
	hours    map[int]bool
	days     map[int]bool
	months   map[int]bool
	weekdays map[int]bool
}

func parseRule(rule string, minCap, maxCap int) (err error, set map[int]bool) {
	items := strings.Split(rule, ",")
	set = make(map[int]bool, len(items))

	for _, item := range items {
		item = strings.TrimSpace(item)

		if item == "" {
			err = fmt.Errorf("error interval format: '%s'", rule)
			return
		}

		if item == "*" {
			for n := minCap; n <= maxCap; n++ {
				set[n] = true
			}

			return
		}

		elements := strings.Split(item, "/")

		if len(elements) == 2 && elements[0] == "*" {
			v, e := strconv.Atoi(elements[1])

			if e != nil {
				err = fmt.Errorf("error interval format: '%s'", rule)
				return
			}

			for n := minCap; n <= maxCap; n++ {
				if n%v == 0 {
					set[n] = true
				}
			}

			continue
		}

		if len(elements) == 1 {
			v, e := strconv.Atoi(elements[0])

			if e != nil || v < minCap && v > maxCap {
				err = fmt.Errorf("error interval format: '%s'", rule)
				return
			}

			set[v] = true
			continue
		}

		err = fmt.Errorf("error interval format: '%s'", rule)
		return
	}

	return
}

func (o *Interval) Check(t time.Time) bool {
	if len(o.minutes) != 0 && !o.minutes[t.Minute()] {
		return false
	}

	if len(o.hours) != 0 && !o.hours[t.Hour()] {
		return false
	}

	if len(o.days) != 0 && !o.days[t.Day()] {
		return false
	}

	if len(o.months) != 0 && !o.months[int(t.Month())] {
		return false
	}

	if len(o.weekdays) != 0 && !o.weekdays[int(t.Weekday())] {
		return false
	}

	return true
}

type IValue struct {
	Ref    *map[int]bool
	MinCap int
	MaxCap int
}

func NewInterval(rules string) (err error, interval *Interval) {
	items := strings.Split(strings.TrimSpace(rules), " ")

	if len(items) != 5 {
		err = fmt.Errorf("error interval format: '%s'", rules)
		return
	}

	interval = &Interval{}
	params := []IValue{
		{&interval.minutes, 0, 59},
		{&interval.hours, 0, 23},
		{&interval.days, 1, 31},
		{&interval.months, 1, 12},
		{&interval.weekdays, 0, 7},
	}

	for index, param := range params {
		if err, *param.Ref = parseRule(items[index], param.MinCap, param.MaxCap); err != nil {
			err = fmt.Errorf("error interval format: '%s'", rules)
			return
		}
	}

	return
}
