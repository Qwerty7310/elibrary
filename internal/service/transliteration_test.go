package service

import "testing"

func TestTransliterateRuToEn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty string",
			in:   "",
			want: "",
		},
		{
			name: "mixed text",
			in:   "Привет, world!",
			want: "Privet, world!",
		},
		{
			name: "hard and soft signs",
			in:   "Подъезд и вьюга",
			want: "Podezd i vyuga",
		},
		{
			name: "upper case letters",
			in:   "ЁЖИК",
			want: "YoZhIK",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := TransliterateRuToEn(tt.in); got != tt.want {
				t.Fatalf("TransliterateRuToEn(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
