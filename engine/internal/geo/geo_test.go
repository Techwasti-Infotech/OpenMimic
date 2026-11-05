package geo

import (
	"strings"
	"testing"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

func TestGeoGeneratorCoordinates(t *testing.T) {
	rng := state.NewDeterministicRand(42)
	cfg := &pb.GeoLocationFunction{
		Mode:   "coordinates",
		Region: "EU",
	}

	gen := NewGeoGenerator(cfg)
	coords := gen.Generate(rng).(map[string]any)

	lat, okLat := coords["latitude"].(float64)
	lon, okLon := coords["longitude"].(float64)

	if !okLat || !okLon {
		t.Fatalf("expected float64 lat/lon coordinates")
	}

	// EU bbox: Lat [36, 70], Lon [-10, 40]
	if lat < 36.0 || lat > 70.0 {
		t.Errorf("latitude %f outside EU bounds [36, 70]", lat)
	}
	if lon < -10.0 || lon > 40.0 {
		t.Errorf("longitude %f outside EU bounds [-10, 40]", lon)
	}
}

func TestGeoGeneratorCorrelatedIP(t *testing.T) {
	rng := state.NewDeterministicRand(100)
	cfg := &pb.GeoLocationFunction{
		Mode:   "correlated_ip",
		Region: "NA",
	}

	gen := NewGeoGenerator(cfg)
	ip := gen.Generate(rng).(string)

	if !strings.HasPrefix(ip, "12.148.") {
		t.Errorf("expected NA IP prefix 12.148., got %s", ip)
	}
}

func TestGeoGeneratorCityAndCountry(t *testing.T) {
	rng := state.NewDeterministicRand(77)
	genCountry := NewGeoGenerator(&pb.GeoLocationFunction{Mode: "country_code", Region: "APAC"})
	country := genCountry.Generate(rng).(string)

	if country != "JP" && country != "SG" && country != "IN" && country != "AU" && country != "KR" {
		t.Errorf("unexpected APAC country: %s", country)
	}
}

