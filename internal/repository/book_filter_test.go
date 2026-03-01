package repository

import "testing"

func TestBookFilterLimitOr(t *testing.T) {
	t.Parallel()

	limit := 15
	if got := (BookFilter{Limit: &limit}).LimitOr(20); got != 15 {
		t.Fatalf("LimitOr() = %d, want %d", got, 15)
	}
	if got := (BookFilter{}).LimitOr(20); got != 20 {
		t.Fatalf("LimitOr() default = %d, want %d", got, 20)
	}
}

func TestBookFilterOffsetOr(t *testing.T) {
	t.Parallel()

	offset := 7
	if got := (BookFilter{Offset: &offset}).OffsetOr(0); got != 7 {
		t.Fatalf("OffsetOr() = %d, want %d", got, 7)
	}
	if got := (BookFilter{}).OffsetOr(0); got != 0 {
		t.Fatalf("OffsetOr() default = %d, want %d", got, 0)
	}
}
