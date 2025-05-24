package score

type roll map[int]int

func Calculate(dieFaces []int) (int, bool) {
	roll := newRoll(dieFaces)

	if roll.ofAKind(6) != 0 {
		return 3000, true
	}

	if roll.hasTriplets() {
		return 2500, true
	}

	if roll.hasFourOfAKindAndAPair() || roll.hasThreePairs() || roll.hasStraight() {
		return 1500, true
	}

	score := 0

	if face := roll.ofAKind(5); face != 0 {
		score += 2000
		delete(roll, face)
	}

	if face := roll.ofAKind(4); face != 0 {
		score += 1000
		delete(roll, face)
	}

	if face := roll.ofAKind(3); face != 0 {
		if face == 1 {
			score += 300
		} else {
			score += face * 100
		}
		delete(roll, face)
	}

	score += roll[1] * 100
	delete(roll, 1)

	score += roll[5] * 50
	delete(roll, 5)

	return score, score != 0
}

func newRoll(dieFaces []int) roll {
	roll := make(roll)
	for _, die := range dieFaces {
		roll[die]++
	}

	return roll
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
