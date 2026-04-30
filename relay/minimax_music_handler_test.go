package relay

import "testing"

func TestPublicPlaygroundMiniMaxMusicModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		model     string
		wantModel string
		wantOK    bool
	}{
		{
			name:      "normalizes paid text-to-music model",
			model:     "music-2.6",
			wantModel: "music-2.6-free",
			wantOK:    true,
		},
		{
			name:      "normalizes paid cover model",
			model:     "music-cover",
			wantModel: "music-cover-free",
			wantOK:    true,
		},
		{
			name:      "keeps public model unchanged",
			model:     "music-2.6-free",
			wantModel: "music-2.6-free",
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotModel, gotOK := publicPlaygroundMiniMaxMusicModel(tt.model)
			if gotModel != tt.wantModel || gotOK != tt.wantOK {
				t.Fatalf("publicPlaygroundMiniMaxMusicModel(%q) = %q, %v; want %q, %v", tt.model, gotModel, gotOK, tt.wantModel, tt.wantOK)
			}
		})
	}
}
