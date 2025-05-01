package tui

import (
	"math/rand"
	"slices"
	"strings"
)

type dicePool []int

func newDicePool(size int) dicePool {
	p := make(dicePool, size)
	for i := range p {
		p[i] = 1
	}
	return p
}

func (p *dicePool) roll() {
	for i := range *p {
		(*p)[i] = rand.Intn(6) + 1
	}
}

func (p *dicePool) contains(face int) bool {
	return slices.Contains(*p, face)
}

func (p *dicePool) add(face int) {
	*p = append(*p, face)
}

func (p *dicePool) remove(face int) {
	for i, n := range *p {
		if n == face {
			*p = slices.Delete(*p, i, i+1)
			return
		}
	}
}

func (p *dicePool) render() string {
	var lines = make([]string, len(diceFaces[1]))

	for _, n := range *p {
		for i, line := range diceFaces[n] {
			lines[i] += line + "  "
		}
	}

	return strings.Join(lines, "\n")
}
