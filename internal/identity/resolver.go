package identity

import (
	"bytes"
	"embed"
	"encoding/csv"
	"io"
	"log"
	"math/rand/v2"
)

//go:embed cities.csv
var citiesFS embed.FS

// Resolver handles device name resolution using a pool of city names,
// avoiding collisions with names already taken by other peers on the network.
type Resolver struct {
	cities []string
}

// NewResolver creates a name resolver with the embedded city name pool.
func NewResolver() *Resolver {
	cities, err := loadCities()
	if err != nil {
		log.Println("Failed to load embedded cities.csv, using default fallbacks:", err)
		cities = []string{"Damascus", "Aleppo", "Homs", "Latakia"}
	}
	return &Resolver{cities: cities}
}

// ResolveName picks a random city name that isn't already taken by existing peers.
func (r *Resolver) ResolveName(takenNames map[string]bool) string {
	// Filter untaken city names
	var available []string
	for _, city := range r.cities {
		if !takenNames[city] {
			available = append(available, city)
		}
	}

	if len(available) == 0 {
		available = r.cities
	}

	// math/rand/v2 auto-seeds — no manual seeding needed
	return available[rand.IntN(len(available))]
}

func loadCities() ([]string, error) {
	data, err := citiesFS.ReadFile("cities.csv")
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(bytes.NewReader(data))
	// Read header
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	var cities []string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) > 0 && record[0] != "" {
			cities = append(cities, record[0])
		}
	}
	return cities, nil
}
