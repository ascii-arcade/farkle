package tui

import (
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
		(*p)[i] = rand.IntN(6) + 1
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

func (p *dicePool) renderCharacters() string {
	if len(*p) == 0 {
		return ""
	}

	output := ""
	for _, n := range *p {
		output += dieCharacters[n] + " "
	}

	return strings.TrimSpace(output)
}

func (p *dicePool) render(start int, end int) string {
	if len(*p) == 0 {
		return ""
	}
	if end > len(*p) {
		end = len(*p)
	}
	if start >= end {
		return ""
	}

	dice := make([]string, 0)

	for _, n := range (*p)[start:end] {
		dice = append(dice, die(n))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, dice...)
}
