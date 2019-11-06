package h3zone

import "github.com/uber/h3-go"

const (
	MinLevel = 0
	MaxLevel = 15
)

var edgeLengthKm = []float64{
	1107.712591, 418.6760055, 158.2446558, 59.81085794,
	22.6063794, 8.544408276, 3.229482772, 1.220629759,
	0.461354684, 0.174375668, 0.065907807, 0.024910561,
	0.009415526, 0.003559893, 0.001348575, 0.000509713,
}

type Zone struct {
	Lng    float64
	Lat    float64
	Length float64
	Level  int
	Hash   string
}

func LoadHash(hash string) *Zone {
	index := h3.FromString(hash)
	level, center := h3.Resolution(index), h3.ToGeo(index)
	length := edgeLengthKm[level]

	return &Zone{Lng: center.Longitude, Lat: center.Latitude, Level: level, Length: length, Hash: hash}
}

func LoadGeo(lng, lat float64, level int) *Zone {
	if level < MinLevel || level > MaxLevel {
		return nil
	}

	index := h3.FromGeo(h3.GeoCoord{Latitude: lat, Longitude: lng}, level)
	center, hash := h3.ToGeo(index), h3.ToString(index)
	length := edgeLengthKm[level]

	return &Zone{Lng: center.Longitude, Lat: center.Latitude, Level: level, Length: length, Hash: hash}
}
