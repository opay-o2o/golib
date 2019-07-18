package tude

import (
	"math"
)

type Polygon struct {
	points []*Point
}

func NewPolygon(points []*Point) *Polygon {
	return &Polygon{points}
}

func (p *Polygon) Points() []*Point {
	return p.points
}

func (p *Polygon) Add(point *Point) {
	p.points = append(p.points, point)
}

func (p *Polygon) IsClosed() bool {
	if len(p.points) < 3 {
		return false
	}

	return true
}

func (p *Polygon) Contains(point *Point) bool {
	if !p.IsClosed() {
		return false
	}

	start := len(p.points) - 1
	end := 0

	contains := p.intersectsWithRaycast(point, p.points[start], p.points[end])

	for i := 1; i < len(p.points); i++ {
		if p.intersectsWithRaycast(point, p.points[i-1], p.points[i]) {
			contains = !contains
		}
	}

	return contains
}

func (p *Polygon) intersectsWithRaycast(point *Point, start *Point, end *Point) bool {
	if start.lng > end.lng {
		start, end = end, start
	}

	for point.lng == start.lng || point.lng == end.lng {
		newLng := math.Nextafter(point.lng, math.Inf(1))
		point = &Point{newLng, point.lat}
	}

	if point.lng < start.lng || point.lng > end.lng {
		return false
	}

	if start.lat > end.lat {
		if point.lat > start.lat {
			return false
		}
		if point.lat < end.lat {
			return true
		}

	} else {
		if point.lat > end.lat {
			return false
		}
		if point.lat < start.lat {
			return true
		}
	}

	raySlope := (point.lng - start.lng) / (point.lat - start.lat)
	diagSlope := (end.lng - start.lng) / (end.lat - start.lat)

	return raySlope >= diagSlope
}
