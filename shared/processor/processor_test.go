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
