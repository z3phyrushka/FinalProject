package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	if nowStr == "" || date == "" || repeat == "" {
		http.Error(w, "params required", http.StatusBadRequest)
		return
	}

	now, err := parseDate(nowStr)
	if err != nil {
		http.Error(w, "invalid now date", http.StatusBadRequest)
		return
	}

	next, err := NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(next))
}

func parseDate(s string) (time.Time, error) {
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

func formatDate(t time.Time) string {
	return t.Format(dateFormat)
}

func afterNow(date, now time.Time) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := now.Date()
	if y1 != y2 {
		return y1 > y2
	}
	if m1 != m2 {
		return m1 > m2
	}
	return d1 > d2
}

func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func isLeap(year int) bool {
	return (year%4 == 0 && year%100 != 0) || (year%400 == 0)
}

func advanceDaysUntil(t time.Time, step int, now time.Time) time.Time {
	next := t.AddDate(0, 0, step)
	for !afterNow(next, now) {
		next = next.AddDate(0, 0, step)
	}
	return next
}

func advanceYearsUntil(t time.Time, now time.Time) time.Time {
	next := t
	for {
		ny := next.Year() + 1
		m := next.Month()
		d := next.Day()
		if m == 2 && d == 29 && !isLeap(ny) {
			m = 3
			d = 1
		}
		next = time.Date(ny, m, d, 0, 0, 0, 0, time.UTC)
		if afterNow(next, now) {
			return next
		}
	}
}

func advanceByOneDayWithPredicate(t time.Time, now time.Time, pred func(time.Time) bool) time.Time {
	cur := t
	for {
		cur = cur.AddDate(0, 0, 1)
		if afterNow(cur, now) && pred(cur) {
			return cur
		}
	}
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("repeat is empty")
	}

	start, err := parseDate(dstart)
	if err != nil {
		return "", errors.New("invalid start date")
	}

	parts := strings.SplitN(strings.TrimSpace(repeat), " ", 3)
	typ := parts[0]

	switch typ {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("invalid d format")
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil || n < 1 || n > 400 {
			return "", errors.New("invalid d value")
		}
		next := advanceDaysUntil(start, n, now)
		return formatDate(next), nil

	case "y":
		if len(parts) != 1 {
			return "", errors.New("invalid y format")
		}
		next := advanceYearsUntil(start, now)
		return formatDate(next), nil

	case "w":
		if len(parts) != 2 {
			return "", errors.New("invalid w format")
		}
		dayStrs := strings.Split(parts[1], ",")
		weekOK := map[time.Weekday]bool{}
		for _, s := range dayStrs {
			if s == "" {
				return "", errors.New("invalid w value")
			}
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 || n > 7 {
				return "", errors.New("invalid w value")
			}
			wd := time.Weekday(n % 7)
			weekOK[wd] = true
		}
		pred := func(t time.Time) bool { return weekOK[t.Weekday()] }
		next := advanceByOneDayWithPredicate(start, now, pred)
		return formatDate(next), nil

	case "m":
		if len(parts) < 2 {
			return "", errors.New("invalid m format")
		}
		dayTokens := strings.Split(parts[1], ",")
		dayOK := map[int]bool{}
		for _, tok := range dayTokens {
			if tok == "" {
				return "", errors.New("invalid m day")
			}
			n, err := strconv.Atoi(tok)
			if err != nil || n == 0 || n < -2 || n > 31 {
				return "", errors.New("invalid m day")
			}
			dayOK[n] = true
		}
		monthOK := map[int]bool{}
		if len(parts) > 2 && strings.TrimSpace(parts[2]) != "" {
			for _, tok := range strings.Split(parts[2], ",") {
				if tok == "" {
					return "", errors.New("invalid m month")
				}
				n, err := strconv.Atoi(tok)
				if err != nil || n < 1 || n > 12 {
					return "", errors.New("invalid m month")
				}
				monthOK[n] = true
			}
		}
		pred := func(t time.Time) bool {
			m := int(t.Month())
			if len(monthOK) > 0 && !monthOK[m] {
				return false
			}
			day := t.Day()
			last := lastDayOfMonth(t.Year(), t.Month())
			if dayOK[day] {
				return true
			}
			if dayOK[-1] && day == last {
				return true
			}
			if dayOK[-2] && day == last-1 {
				return true
			}
			return false
		}
		next := advanceByOneDayWithPredicate(start, now, pred)
		return formatDate(next), nil

	default:
		return "", errors.New("unsupported repeat format")
	}
}
