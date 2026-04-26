// Package sleep implements the `sleep` applet: pause for a duration.
package sleep

import (
	"errors"
	"strconv"
	"time"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "sleep",
		Help:  "delay for a specified amount of time",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: sleep NUMBER[s|m|h|d]...
Pause for NUMBER seconds (default), or with a unit suffix:
  s   seconds  m   minutes  h   hours  d   days

If multiple NUMBERs are given, sleep the sum of all durations.
`

// Main is the applet entry point.
func Main(argv []string) int {
	if len(argv) < 2 {
		ioutil.Errf("sleep: missing operand")
		return 2
	}
	var total time.Duration
	for _, a := range argv[1:] {
		d, err := parseDuration(a)
		if err != nil {
			ioutil.Errf("sleep: invalid time interval: %s", a)
			return 2
		}
		total += d
	}
	time.Sleep(total)
	return 0
}

func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, errors.New("empty duration")
	}
	mult := time.Second
	switch s[len(s)-1] {
	case 's':
		s = s[:len(s)-1]
	case 'm':
		mult = time.Minute
		s = s[:len(s)-1]
	case 'h':
		mult = time.Hour
		s = s[:len(s)-1]
	case 'd':
		mult = 24 * time.Hour
		s = s[:len(s)-1]
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, errors.New("negative duration")
	}
	return time.Duration(n * float64(mult)), nil
}
