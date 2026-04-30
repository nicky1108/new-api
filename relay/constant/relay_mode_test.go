package constant

import "testing"

func TestPath2RelayModePlaygroundImageRoutes(t *testing.T) {
	tests := []struct {
		name string
		path string
		want int
	}{
		{
			name: "playground image generation",
			path: "/pg/images/generations",
			want: RelayModeImagesGenerations,
		},
		{
			name: "playground image edit",
			path: "/pg/images/edits",
			want: RelayModeImagesEdits,
		},
		{
			name: "playground audio speech",
			path: "/pg/audio/speech",
			want: RelayModeAudioSpeech,
		},
		{
			name: "playground music generation",
			path: "/pg/music/generations",
			want: RelayModeMiniMaxMusic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Path2RelayMode(tt.path); got != tt.want {
				t.Fatalf("Path2RelayMode(%q) = %d, want %d", tt.path, got, tt.want)
			}
		})
	}
}
