package relay

import "testing"

func TestIsOfficialMiniMaxMusicModel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		model string
		want  bool
	}{
		{model: "music-2.6", want: true},
		{model: "music-cover", want: true},
		{model: "music-2.6-free", want: true},
		{model: "music-cover-free", want: true},
		{model: "custom-music-alias", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.model, func(t *testing.T) {
			t.Parallel()

			got := isOfficialMiniMaxMusicModel(tt.model)
			if got != tt.want {
				t.Fatalf("isOfficialMiniMaxMusicModel(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}
