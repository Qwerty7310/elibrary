package domain

import (
	"errors"
	"testing"
)

func TestLocationTypeLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   LocationType
		want int
	}{
		{LocationTypeBuilding, 1},
		{LocationTypeRoom, 2},
		{LocationTypeCabinet, 3},
		{LocationTypeShelf, 4},
		{LocationType("unknown"), -1},
	}

	for _, tt := range tests {
		if got := tt.in.Level(); got != tt.want {
			t.Fatalf("Level(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestLocationTypeIsChildOf(t *testing.T) {
	t.Parallel()

	if !LocationTypeRoom.IsChildOf(LocationTypeBuilding) {
		t.Fatal("room should be child of building")
	}
	if LocationTypeShelf.IsChildOf(LocationTypeBuilding) {
		t.Fatal("shelf should not be direct child of building")
	}
	if LocationTypeBuilding.IsChildOf(LocationTypeShelf) {
		t.Fatal("building should not be child of shelf")
	}
}

func TestParseLocationType(t *testing.T) {
	t.Parallel()

	valid := []LocationType{
		LocationTypeBuilding,
		LocationTypeRoom,
		LocationTypeCabinet,
		LocationTypeShelf,
	}

	for _, want := range valid {
		got, err := ParseLocationType(string(want))
		if err != nil {
			t.Fatalf("ParseLocationType(%q) error = %v", want, err)
		}
		if got != want {
			t.Fatalf("ParseLocationType(%q) = %q, want %q", want, got, want)
		}
	}

	_, err := ParseLocationType("floor")
	if !errors.Is(err, ErrInvalidLocationType) {
		t.Fatalf("ParseLocationType() error = %v, want %v", err, ErrInvalidLocationType)
	}
}
