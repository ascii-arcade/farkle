package score

import (
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name     string
		dieFaces []int
		want     int
		wantOk   bool
	}{
		{
			name:     "no score",
			dieFaces: []int{4, 6, 4, 2, 3, 3},
			want:     0,
			wantOk:   false,
		},
		{
			name:     "four of a kind of ones and a five",
			dieFaces: []int{1, 1, 1, 1, 2, 5},
			want:     1050,
			wantOk:   true,
		},
		{
			name:     "six of a kind of fours",
			dieFaces: []int{4, 4, 4, 4, 4, 4},
			want:     3000,
			wantOk:   true,
		},
		{
			name:     "six of a kind of fives",
			dieFaces: []int{5, 5, 5, 5, 5, 5},
			want:     3000,
			wantOk:   true,
		},
		{
			name:     "five of a kind of ones with a five",
			dieFaces: []int{1, 1, 1, 1, 1, 5},
			want:     2050,
			wantOk:   true,
		},
		{
			name:     "four of a kind of ones with a five",
			dieFaces: []int{1, 1, 1, 1, 5},
			want:     1050,
			wantOk:   true,
		},
		{
			name:     "three of a kind of threes and fives",
			dieFaces: []int{3, 3, 3, 5, 5, 5},
			want:     2500,
			wantOk:   true,
		},
		{
			name:     "four of a kind of threes and a pair of fives",
			dieFaces: []int{3, 3, 3, 3, 5, 5},
			want:     1500,
			wantOk:   true,
		},
		{
			name:     "three pairs",
			dieFaces: []int{1, 1, 3, 3, 4, 4},
			want:     1500,
			wantOk:   true,
		},
		{
			name:     "three pairs",
			dieFaces: []int{2, 2, 4, 4, 5, 5},
			want:     1500,
			wantOk:   true,
		},
		{
			name:     "straight",
			dieFaces: []int{1, 2, 3, 4, 5, 6},
			want:     1500,
			wantOk:   true,
		},
		{
			name:     "four of a kind of fours and a five",
			dieFaces: []int{4, 4, 4, 4, 4, 5},
			want:     2050,
			wantOk:   true,
		},
		{
			name:     "five of a kind of fours",
			dieFaces: []int{4, 4, 4, 4, 4},
			want:     2000,
			wantOk:   true,
		},
		{
			name:     "five of a kind of fours with a five",
			dieFaces: []int{5, 4, 4, 4, 4, 4},
			want:     2050,
			wantOk:   true,
		},
		{
			name:     "five of a kind of ones",
			dieFaces: []int{1, 1, 1, 1, 1},
			want:     2000,
			wantOk:   true,
		},
		{
			name:     "four of a kind of ones",
			dieFaces: []int{1, 1, 1, 1},
			want:     1000,
			wantOk:   true,
		},
		{
			name:     "four of a kind of ones with a five",
			dieFaces: []int{1, 1, 1, 1, 5},
			want:     1050,
			wantOk:   true,
		},
		{
			name:     "three of a kind of ones and a five",
			dieFaces: []int{1, 1, 1, 5},
			want:     350,
			wantOk:   true,
		},
		{
			name:     "three of a kind of fives",
			dieFaces: []int{5, 5, 5},
			want:     500,
			wantOk:   true,
		},
		{
			name:     "three of a kind of threes and a five",
			dieFaces: []int{3, 3, 3, 5},
			want:     350,
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Calculate(tt.dieFaces)
			if !ok && tt.wantOk {
				t.Errorf("Calculate(%v) = not okay, want ok", tt.dieFaces)
			}
			if ok && !tt.wantOk {
				t.Errorf("Calculate(%v) = ok, want not okay", tt.dieFaces)
			}
			if got != tt.want {
				t.Errorf("Calculate(%v) = %d, want %d", tt.dieFaces, got, tt.want)
			}
		})
	}
}
