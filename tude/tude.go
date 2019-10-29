package tude

import (
	"math"
)

const (
	R = 6371000
)

type Shape interface {
	Contains(point *Point) bool
}

type Point struct {
	lng, lat float64
}

func NewPoint(lng, lat float64) *Point {
	return &Point{lng, lat}
}

type TimePoint struct {
	point     *Point
	timestamp int64
}

func (p *TimePoint) GetPoint() *Point {
	return p.point
}

func NewTimePoint(lng, lat float64, timestamp int64) *TimePoint {
	return &TimePoint{&Point{lng, lat}, timestamp}
}

func Radians(x float64) float64 {
	return x * math.Pi / 180
}

func Distance(p1, p2 *Point) float64 {
	avgLat := Radians(p1.lat+p2.lat) / 2
	disLat := R * math.Cos(avgLat) * Radians(p1.lng-p2.lng)
	disLon := R * Radians(p1.lat-p2.lat)
	return math.Sqrt(disLat*disLat + disLon*disLon)
}

func Angle(p1, p2 *Point) float64 {
	numerator := math.Sin(Radians(p2.lng-p1.lng)) * math.Cos(Radians(p2.lat))
	denominator := math.Cos(Radians(p1.lat))*math.Sin(Radians(p2.lat)) - math.Sin(Radians(p1.lat))*math.Cos(Radians(p2.lat))*math.Cos(Radians(p2.lng-p1.lng))
	angle := math.Atan2(math.Abs(numerator), math.Abs(denominator))

	if p2.lng > p1.lng {
		if p2.lat < p1.lat {
			angle = math.Pi - angle
		} else if p2.lat == p1.lat {
			angle = math.Pi / 2
		}
	} else if p2.lng < p1.lng {
		if p2.lat > p1.lat {
			angle = 2*math.Pi - angle
		} else if p2.lat < p1.lat {
			angle = math.Pi + angle
		} else {
			angle = math.Pi * 3 / 2
		}
	} else {
		if p2.lat >= p1.lat {
			angle = 0
		} else {
			angle = math.Pi
		}
	}

	return angle * 180 / math.Pi
}
