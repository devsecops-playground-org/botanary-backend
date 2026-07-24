package catalog

import (
	"errors"
	"sync"
	"testing"
)

func TestAddAssignsSequentialIDs(t *testing.T) {
	c := New()

	first, err := c.Add(Listing{Name: "Aloe", PriceCts: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, _ := c.Add(Listing{Name: "Basil", PriceCts: 200})

	if first.ID != 1 || second.ID != 2 {
		t.Fatalf("got IDs %d and %d, want 1 and 2", first.ID, second.ID)
	}
}

func TestAddRejectsInvalidListings(t *testing.T) {
	c := New()

	for _, l := range []Listing{
		{Name: "   ", PriceCts: 100},
		{Name: "Negative", PriceCts: -1},
	} {
		if _, err := c.Add(l); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Add(%+v) = %v, want ErrInvalid", l, err)
		}
	}
}

func TestGetUnknownID(t *testing.T) {
	if _, err := New().Get(7); !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestListFiltersOutOfStock(t *testing.T) {
	c := NewSeeded()

	if got := len(c.List(false)); got != 3 {
		t.Fatalf("got %d listings, want 3", got)
	}
	inStock := c.List(true)
	if len(inStock) != 2 {
		t.Fatalf("got %d in stock, want 2", len(inStock))
	}
	if inStock[0].ID > inStock[1].ID {
		t.Fatal("listings are not sorted by ID")
	}
}

func TestConcurrentAddsDoNotRace(t *testing.T) {
	c := New()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.Add(Listing{Name: "Fern", PriceCts: 1})
		}()
	}
	wg.Wait()

	if got := len(c.List(false)); got != 50 {
		t.Fatalf("got %d listings after concurrent adds, want 50", got)
	}
}
