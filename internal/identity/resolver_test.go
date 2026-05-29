package identity

import (
	"testing"
)

func TestNewResolver(t *testing.T) {
	resolver := NewResolver()
	if resolver == nil {
		t.Fatal("Expected NewResolver() to return a non-nil Resolver")
	}

	if len(resolver.cities) == 0 {
		t.Error("Expected Resolver to load city names from embedded cities.csv")
	}
}

func TestResolveName(t *testing.T) {
	resolver := NewResolver()

	// 1. Check resolving with no taken names
	name1 := resolver.ResolveName(nil)
	if name1 == "" {
		t.Error("Expected resolved name to be non-empty")
	}

	// 2. Check collision avoidance
	taken := make(map[string]bool)
	for _, city := range resolver.cities {
		taken[city] = true
	}

	// Since all cities are taken, it should fall back to any city name in the pool
	name2 := resolver.ResolveName(taken)
	if name2 == "" {
		t.Error("Expected resolved name to be non-empty even when all names are taken")
	}

	// 3. Selective avoidance
	takenSelective := map[string]bool{
		resolver.cities[0]: true,
	}
	name3 := resolver.ResolveName(takenSelective)
	if name3 == resolver.cities[0] {
		t.Errorf("Expected resolved name not to be %q as it is marked taken", resolver.cities[0])
	}
}
