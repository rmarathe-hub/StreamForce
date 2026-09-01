package processor

import "testing"

func TestTargetHeights(t *testing.T) {
	tests := []struct {
		source int
		want   []int
	}{
		{1080, []int{1080, 720, 480}},
		{720, []int{720, 480}},
		{480, []int{480}},
		{360, []int{360}},
	}

	for _, tt := range tests {
		got := targetHeights(tt.source)
		if len(got) != len(tt.want) {
			t.Fatalf("targetHeights(%d) = %v, want %v", tt.source, got, tt.want)
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Fatalf("targetHeights(%d) = %v, want %v", tt.source, got, tt.want)
			}
		}
	}
}

func TestOverallProgress(t *testing.T) {
	tests := []struct {
		variantIndex   int
		variantCount   int
		variantPercent int
		want           int
	}{
		{0, 3, 0, 5},
		{0, 3, 50, 20},
		{1, 3, 0, 35},
		{2, 3, 100, 95},
		{0, 1, 100, 95},
	}

	for _, tt := range tests {
		got := overallProgress(tt.variantIndex, tt.variantCount, tt.variantPercent)
		if got != tt.want {
			t.Fatalf(
				"overallProgress(%d, %d, %d) = %d, want %d",
				tt.variantIndex,
				tt.variantCount,
				tt.variantPercent,
				got,
				tt.want,
			)
		}
	}
}
