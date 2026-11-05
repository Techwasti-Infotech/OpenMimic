package geo

import (
	"fmt"
	"strings"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

// RegionBox defines standard geographic bounding boxes and country metadata.
type RegionBox struct {
	MinLat   float64
	MaxLat   float64
	MinLon   float64
	MaxLon   float64
	Countries []string
	Cities   []string
	IPPrefix string
}

var regionCatalog = map[string]RegionBox{
	"NA": {
		MinLat:    24.396308,
		MaxLat:    49.384358,
		MinLon:    -125.000000,
		MaxLon:    -66.934570,
		Countries: []string{"US", "CA", "MX"},
		Cities:    []string{"New York", "San Francisco", "Toronto", "Chicago", "Vancouver"},
		IPPrefix:  "12.148",
	},
	"EU": {
		MinLat:    36.0,
		MaxLat:    70.0,
		MinLon:    -10.0,
		MaxLon:    40.0,
		Countries: []string{"DE", "FR", "GB", "NL", "IE"},
		Cities:    []string{"Frankfurt", "Paris", "London", "Amsterdam", "Dublin"},
		IPPrefix:  "82.165",
	},
	"APAC": {
		MinLat:    -45.0,
		MaxLat:    55.0,
		MinLon:    68.0,
		MaxLon:    180.0,
		Countries: []string{"JP", "SG", "IN", "AU", "KR"},
		Cities:    []string{"Tokyo", "Singapore", "Mumbai", "Sydney", "Seoul"},
		IPPrefix:  "115.112",
	},
	"GLOBAL": {
		MinLat:    -90.0,
		MaxLat:    90.0,
		MinLon:    -180.0,
		MaxLon:    180.0,
		Countries: []string{"US", "DE", "JP", "GB", "SG", "BR", "AU", "IN"},
		Cities:    []string{"New York", "London", "Tokyo", "Berlin", "Singapore", "Sydney"},
		IPPrefix:  "198.51",
	},
}

// GeoGenerator produces geographically correlated GPS coordinates, country codes, cities, and IPs.
type GeoGenerator struct {
	mode   string
	box    RegionBox
}

// NewGeoGenerator builds a GeoGenerator from configuration.
func NewGeoGenerator(cfg *pb.GeoLocationFunction) *GeoGenerator {
	regKey := strings.ToUpper(cfg.GetRegion())
	box, exists := regionCatalog[regKey]
	if !exists {
		box = regionCatalog["GLOBAL"]
	}

	// Custom override bounds if specified
	if cfg.GetMinLatitude() != 0 || cfg.GetMaxLatitude() != 0 {
		box.MinLat = cfg.GetMinLatitude()
		box.MaxLat = cfg.GetMaxLatitude()
	}
	if cfg.GetMinLongitude() != 0 || cfg.GetMaxLongitude() != 0 {
		box.MinLon = cfg.GetMinLongitude()
		box.MaxLon = cfg.GetMaxLongitude()
	}

	mode := cfg.GetMode()
	if mode == "" {
		mode = "coordinates"
	}

	return &GeoGenerator{
		mode: mode,
		box:  box,
	}
}

// Generate produces the requested geo attribute.
func (g *GeoGenerator) Generate(rng *state.DeterministicRand) any {
	switch g.mode {
	case "country_code":
		idx := rng.Intn(len(g.box.Countries))
		return g.box.Countries[idx]

	case "city":
		idx := rng.Intn(len(g.box.Cities))
		return g.box.Cities[idx]

	case "correlated_ip":
		octet3 := rng.Intn(254) + 1
		octet4 := rng.Intn(254) + 1
		return fmt.Sprintf("%s.%d.%d", g.box.IPPrefix, octet3, octet4)

	case "coordinates":
		fallthrough
	default:
		lat := g.box.MinLat + rng.Float64()*(g.box.MaxLat-g.box.MinLat)
		lon := g.box.MinLon + rng.Float64()*(g.box.MaxLon-g.box.MinLon)
		return map[string]any{
			"latitude":  lat,
			"longitude": lon,
		}
	}
}

