package example_test

import "testing"

func TestExample(t *testing.T) {
	tests := []struct {
		name string
		give string // TODO: replace with actual input type
		want string // TODO: replace with actual output type
	}{
		{
			name: "basic case",
			give: "",
			want: "",
		},
		// TODO: add more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Example(tt.give)
			if got != tt.want {
				t.Errorf("Example(%q) = %q, want %q", tt.give, got, tt.want)
			}
			// For richer diffs, consider:
			//   if diff := cmp.Diff(tt.want, got); diff != "" {
			//       t.Errorf("Example() mismatch (-want +got):\n%s", diff)
			//   }
		})
	}
}
