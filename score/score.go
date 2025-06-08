package score

import (
	"errors"
	"fmt"
)

type roll map[int]int

func Calculate(dieFaces []int, ignoreUselessDie bool) (int, []int, error) {
	roll, err := newRoll(dieFaces)
	used := make([]int, 0)
	if err != nil {
		return 0, []int{}, err
	}

	if roll.ofAKind(6) != 0 {
		return 3000, dieFaces, nil
	}

	if roll.hasTriplets() {
		return 2500, dieFaces, nil
	}

	if roll.hasFourOfAKindAndAPair() || roll.hasThreePairs() || roll.hasStraight() {
		return 1500, dieFaces, nil
	}

	score := 0

	if face := roll.ofAKind(5); face != 0 {
		used = append(used, face, face, face, face, face)
		score += 2000
		roll[face] = 0
	}

	if face := roll.ofAKind(4); face != 0 {
		used = append(used, face, face, face, face)
		score += 1000
		roll[face] = 0
	}

	if face := roll.ofAKind(3); face != 0 {
		used = append(used, face, face, face)
		if face == 1 {
			score += 300
		} else {
			score += face * 100
		}
		roll[face] = 0
	}

	score += roll[1] * 100
	for range roll[1] {
		used = append(used, 1)
	}
	roll[1] = 0

	score += roll[5] * 50
	for range roll[5] {
		used = append(used, 5)
	}
	roll[5] = 0

	if (roll[0] > 0 || roll[2] > 0 || roll[3] > 0 || roll[4] > 0 || roll[6] > 0) && !ignoreUselessDie {
		return 0, []int{}, errors.New("useless dice detected")
	}

	if score == 0 && len(dieFaces) > 0 {
		return 0, []int{}, errors.New("no score")
	}

	return score, used, nil
}

func newRoll(dieFaces []int) (roll, error) {
	if err := validateDieFaces(dieFaces); err != nil {
		return nil, err
	}

	roll := make(roll)
	for _, die := range dieFaces {
		roll[die]++
	}

	return roll, nil
}

func validateDieFaces(dieFaces []int) error {
	if len(dieFaces) == 0 {
		return errors.New("no dice")
	}

	if len(dieFaces) > 6 {
		return errors.New("too many dice")
	}

	for _, die := range dieFaces {
		if die < 1 || die > 6 {
			return fmt.Errorf("invalid die face: %d", die)
		}
	}

	return nil
}

func (r roll) ofAKind(targetCount int) int {
	for face, count := range r {
		if count == targetCount {
			return face
		}
	}

	return 0
}

func (r roll) hasTriplets() bool {
	tripletCount := 0
	for _, count := range r {
		if count == 3 {
			tripletCount++
		}
	}

	return tripletCount == 2
}

func (r roll) hasFourOfAKindAndAPair() bool {
	return r.ofAKind(4) != 0 && r.ofAKind(2) != 0
}

func (r roll) hasThreePairs() bool {
	pairCount := 0
	for _, count := range r {
		if count == 2 {
			pairCount++
		}
	}

	return pairCount == 3
}

func (r roll) hasStraight() bool {
	singlesCount := 0
	for i := 1; i <= 6; i++ {
		if r[i] == 1 {
			singlesCount++
		}
	}

	return singlesCount == 6
}
