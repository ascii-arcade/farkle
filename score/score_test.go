package score

import (
	"testing"
)

func TestScore(t *testing.T) {
	tests := []struct {
		dieFaces  []int
		wantScore int
		wantError string
	}{
		{
			dieFaces:  []int{1, 1, 1, 4, 5, 5, 6},
			wantScore: 0,
			wantError: "too many dice",
		},
		{
			dieFaces:  []int{},
			wantScore: 0,
			wantError: "no dice",
		},
		{
			dieFaces:  []int{0, 1},
			wantScore: 0,
			wantError: "invalid die face: 0",
		},
		{
			dieFaces:  []int{1, 7},
			wantScore: 0,
			wantError: "invalid die face: 7",
		},
		{
			dieFaces:  []int{1, 1, 1, 1, 2, 5},
			wantScore: 0,
			wantError: "useless dice detected",
		},
		{
			dieFaces:  []int{4, 4, 4, 4, 4, 4},
			wantScore: 3000,
		},
		{
			dieFaces:  []int{5, 5, 5, 5, 5, 5},
			wantScore: 3000,
		},
		{
			dieFaces:  []int{1, 1, 1, 1, 1, 5},
			wantScore: 2050,
		},
		{
			dieFaces:  []int{1, 1, 1, 1, 5},
			wantScore: 1050,
		},
		{
			dieFaces:  []int{3, 3, 3, 5, 5, 5},
			wantScore: 2500,
		},
		{
			dieFaces:  []int{3, 3, 3, 3, 5, 5},
			wantScore: 1500,
		},
		{
			dieFaces:  []int{1, 1, 3, 3, 4, 4},
			wantScore: 1500,
		},
		{
			dieFaces:  []int{1, 2, 3, 4, 5, 6},
			wantScore: 1500,
		},
		{
			dieFaces:  []int{4, 4, 4, 4, 4, 5},
			wantScore: 2050,
		},
		{
			dieFaces:  []int{4, 4, 4, 4, 4},
			wantScore: 2000,
		},
		{
			dieFaces:  []int{5, 4, 4, 4, 4, 4},
			wantScore: 2050,
		},
		{
			dieFaces:  []int{1, 1, 1, 1, 1},
			wantScore: 2000,
		},
		{
			dieFaces:  []int{1, 1, 1, 1},
			wantScore: 1000,
		},
		{
			dieFaces:  []int{1, 1, 1, 1, 5},
			wantScore: 1050,
		},
		{
			dieFaces:  []int{1, 1, 1, 5},
			wantScore: 350,
		},
		{
			dieFaces:  []int{5, 5, 5},
			wantScore: 500,
		},
		{
			dieFaces:  []int{3, 3, 3, 5},
			wantScore: 350,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got, _, err := Calculate(tt.dieFaces, false)

			if tt.wantError != "" {
				if err == nil {
					t.Errorf("Calculate(%v) = no error, want %v", tt.dieFaces, tt.wantError)
				} else if err.Error() != tt.wantError {
					t.Errorf("Calculate(%v) = %v, want %v", tt.dieFaces, err.Error(), tt.wantError)
				}
			}

			if tt.wantError == "" && err != nil {
				t.Errorf("Calculate(%v) = unexpected error: %v", tt.dieFaces, err)
			}

			if got != tt.wantScore {
				t.Errorf("Calculate(%v) = %d, want %d", tt.dieFaces, got, tt.wantScore)
			}
		})
	}
}

func TestUsedDice(t *testing.T) {
	tests := []struct {
		dieFaces     []int
		wantDieFaces []int
	}{
		{
			dieFaces:     []int{1, 1, 4, 4, 5, 5},
			wantDieFaces: []int{1, 1, 4, 4, 5, 5},
		},
		{
			dieFaces:     []int{1, 1, 4, 5, 5, 6},
			wantDieFaces: []int{1, 1, 5, 5},
		},
		{
			dieFaces:     []int{1, 3, 3, 4, 4},
			wantDieFaces: []int{1},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			_, gotUsedFaces, err := Calculate(tt.dieFaces, true)

			if err != nil {
				t.Errorf("Calculate(%v) = unexpected error: %v", tt.dieFaces, err)
			}

			if !equalSlices(gotUsedFaces, tt.wantDieFaces) {
				t.Errorf("Calculate(%v) = %v, want %v", tt.dieFaces, gotUsedFaces, tt.wantDieFaces)
			}
		})
	}
}

func equalSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	amap := make(map[int]int)
	bmap := make(map[int]int)
	for _, v := range a {
		amap[v]++
	}
	for _, v := range b {
		bmap[v]++
	}
	for k, v := range amap {
		if bmap[k] != v {
			return false
		}
	}
	return true
}
