package tude

type Circle struct {
	center *Point
	radius float64
}

func (c *Circle) Radius() float64 {
	return c.radius
}

func (c *Circle) Center() *Point {
	return c.center
}

func (c *Circle) Contains(point *Point) bool {
	return Distance(point, c.center) <= c.radius
}

func NewCircle(center *Point, radius float64) *Circle {
	return &Circle{center, radius}
}
