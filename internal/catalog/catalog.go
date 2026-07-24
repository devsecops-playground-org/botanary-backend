// Package catalog holds the plant listings the API serves.
package catalog

import (
	"errors"
	"sort"
	"strings"
	"sync"
)

var (
	ErrNotFound = errors.New("listing not found")
	ErrInvalid  = errors.New("listing is invalid")
)

type Listing struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Species  string `json:"species"`
	PriceCts int    `json:"price_cts"`
	InStock  bool   `json:"in_stock"`
}

type Catalog struct {
	mu     sync.RWMutex
	items  map[int]Listing
	nextID int
}

func New() *Catalog {
	return &Catalog{items: make(map[int]Listing), nextID: 1}
}

func NewSeeded() *Catalog {
	c := New()
	for _, l := range []Listing{
		{Name: "Monstera Deliciosa", Species: "monstera", PriceCts: 4500, InStock: true},
		{Name: "Fiddle Leaf Fig", Species: "ficus", PriceCts: 8900, InStock: true},
		{Name: "Snake Plant", Species: "sansevieria", PriceCts: 2400, InStock: false},
	} {
		_, _ = c.Add(l)
	}
	return c
}

func (c *Catalog) Add(l Listing) (Listing, error) {
	if strings.TrimSpace(l.Name) == "" {
		return Listing{}, ErrInvalid
	}
	if l.PriceCts < 0 {
		return Listing{}, ErrInvalid
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	l.ID = c.nextID
	c.nextID++
	c.items[l.ID] = l
	return l, nil
}

func (c *Catalog) Get(id int) (Listing, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	l, ok := c.items[id]
	if !ok {
		return Listing{}, ErrNotFound
	}
	return l, nil
}

// List returns listings sorted by ID, optionally filtered to those in stock.
func (c *Catalog) List(inStockOnly bool) []Listing {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]Listing, 0, len(c.items))
	for _, l := range c.items {
		if inStockOnly && !l.InStock {
			continue
		}
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
