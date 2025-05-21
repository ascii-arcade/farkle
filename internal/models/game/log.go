package game

import "strings"

type log []string

func (l *log) entries() string {
	if len(*l) <= 15 {
		return strings.Join(*l, "\n")
	}

	recent := (*l)[len(*l)-15:]

	return strings.Join(recent, "\n")
}
