package utils

import "math"

func Euclidean(x1 int, y1 int, x2 int, y2 int) float64 {
  return math.Sqrt(math.Pow(float64((x1 - x2)), 2) + math.Pow(float64((y1 - y2)), 2))
}

func Abs(n int) int {
  if (n < 0) {
    return n * -1
  }
  return n
}

