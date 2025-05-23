package dice

import (
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/ascii-arcade/farkle/score"
)

type DicePool []int

func NewDicePool(size int) DicePool {
	p := make(DicePool, size)
	for i := range p {
		p[i] = 1
	}
	return p
}

func (p *DicePool) Roll() {
	for i := range *p {
		(*p)[i] = rand.IntN(6) + 1
	}
}

func (p *DicePool) Contains(face int) bool {
	return slices.Contains(*p, face)
}

func (p *DicePool) Add(face int) {
	*p = append(*p, face)
}

func (p *DicePool) Remove(face int) {
	for i, n := range *p {
		if n == face {
			*p = slices.Delete(*p, i, i+1)
			return
		}
	}
}

func (p *DicePool) Score() (int, error) {
	return score.Calculate(*p)
}

func (p *DicePool) RenderCharacters() string {
	if len(*p) == 0 {
		return ""
	}

	output := ""
	for _, n := range *p {
		output += diceCharacters[n] + " "
	}

	return strings.TrimSpace(output)
}

func (p *DicePool) Render(start int, end int) string {
	if len(*p) == 0 {
		return ""
	}
	if end > len(*p) {
		end = len(*p)
	}
	if start >= end {
		return ""
	}

	var lines = make([]string, len(diceFaces[1]))

	for i, n := range (*p)[start:end] {
		for j, line := range diceFaces[n] {
			if i == len((*p)[start:end])-1 {
				lines[j] += line
			} else {
				lines[j] += line + "  "
			}
		}
	}

	return strings.Join(lines, "\n")
}
