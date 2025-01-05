package utils

import "gioui.org/f32"

func squaredDistance(a, b f32.Point) float32 {
	dx := a.X - b.X
	dy := a.Y - b.Y

	return dx*dx + dy*dy
}

func SimplifyRadialDistance(points []f32.Point, tolerance float32) []f32.Point {
	if len(points) <= 1 {
		return points
	}

	squaredTolerance := tolerance * tolerance

	previousPoint := points[0]
	newPoints := []f32.Point{previousPoint}
	var point f32.Point

	for i := 1; i < len(points); i++ {
		point = points[i]

		if squaredDistance(point, previousPoint) > squaredTolerance {
			newPoints = append(newPoints, point)
			previousPoint = point
		}
	}

	if previousPoint != point {
		newPoints = append(newPoints, point)
	}

	return newPoints
}

func squaredPerpendicularDistance(p, a, b f32.Point) float32 {
	x, y := a.X, a.Y
	dx, dy := b.X-x, b.Y-y

	if dx != 0 || dy != 0 {
		t := ((p.X-x)*dx + (p.Y-y)*dy) / (dx*dx + dy*dy)

		if t > 1 {
			x, y = b.X, b.Y
		} else if t > 0 {
			x += dx * t
			y += dy * t
		}
	}

	dx = p.X - x
	dy = p.Y - y

	return dx*dx + dy*dy
}

func douglasPeuckerStep(points []f32.Point, start, end int, squaredTolerance float32, simplifiedPoints *[]f32.Point) {
	maxSquaredDistance := squaredTolerance
	var index int

	for i := start + 1; i < end; i++ {
		sqrDistance := squaredPerpendicularDistance(points[i], points[start], points[end])

		if sqrDistance > maxSquaredDistance {
			index = i
			maxSquaredDistance = sqrDistance
		}
	}

	if maxSquaredDistance > squaredTolerance {
		if index-start > 1 {
			douglasPeuckerStep(points, start, index, squaredTolerance, simplifiedPoints)
		}
		*simplifiedPoints = append(*simplifiedPoints, points[index])
		if end-index > 1 {
			douglasPeuckerStep(points, index, end, squaredTolerance, simplifiedPoints)
		}
	}
}

func SimplifyDouglasPeucker(points []f32.Point, tolerance float32) []f32.Point {
	if len(points) <= 1 {
		return points
	}

	squaredTolerance := tolerance * tolerance

	end := len(points) - 1

	simplified := []f32.Point{points[0]}
	douglasPeuckerStep(points, 0, end, squaredTolerance, &simplified)
	simplified = append(simplified, points[end])

	return simplified
}

func SimplifyPoints(points []f32.Point, tolerance float32) []f32.Point {
	return SimplifyDouglasPeucker(SimplifyRadialDistance(points, tolerance), tolerance)
}
