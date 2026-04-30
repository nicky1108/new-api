package minimax

import "testing"

func TestFlattenVoiceOptionsKeepsCategories(t *testing.T) {
	response := VoiceListResponse{
		SystemVoice: []VoiceItem{
			{
				VoiceID:     "Chinese (Mandarin)_Reliable_Executive",
				VoiceName:   "沉稳高管",
				Description: []string{"一位沉稳可靠的中年男性高管声音"},
				CreatedTime: "1970-01-01",
			},
		},
		VoiceCloning: []VoiceItem{
			{VoiceID: "clone-voice", CreatedTime: "2025-08-20"},
		},
		VoiceGeneration: []VoiceItem{
			{VoiceID: "generated-voice", CreatedTime: "2025-08-21"},
		},
	}

	voices := FlattenVoiceOptions(response)

	if len(voices) != 3 {
		t.Fatalf("len(voices) = %d, want 3", len(voices))
	}
	if voices[0].VoiceID != "Chinese (Mandarin)_Reliable_Executive" {
		t.Fatalf("voices[0].VoiceID = %q", voices[0].VoiceID)
	}
	if voices[0].VoiceName != "沉稳高管" {
		t.Fatalf("voices[0].VoiceName = %q", voices[0].VoiceName)
	}
	if voices[0].Category != "system_voice" {
		t.Fatalf("voices[0].Category = %q, want system_voice", voices[0].Category)
	}
	if voices[0].Label != "沉稳高管 (Chinese (Mandarin)_Reliable_Executive)" {
		t.Fatalf("voices[0].Label = %q", voices[0].Label)
	}
	if voices[1].Category != "voice_cloning" {
		t.Fatalf("voices[1].Category = %q, want voice_cloning", voices[1].Category)
	}
	if voices[2].Category != "voice_generation" {
		t.Fatalf("voices[2].Category = %q, want voice_generation", voices[2].Category)
	}
}

func TestNormalizeVoiceType(t *testing.T) {
	got, ok := NormalizeVoiceType("")
	if !ok || got != "all" {
		t.Fatalf("normalize empty = %q, %v; want all, true", got, ok)
	}

	got, ok = NormalizeVoiceType("voice_cloning")
	if !ok || got != "voice_cloning" {
		t.Fatalf("normalize voice_cloning = %q, %v", got, ok)
	}

	got, ok = NormalizeVoiceType("music")
	if ok || got != "" {
		t.Fatalf("normalize invalid = %q, %v; want empty, false", got, ok)
	}
}
