package utils

import (
  "fmt"
)

func Foreground(rgb string, message string) string {
  return fmt.Sprintf("\x1b[38;2;%sm%s\x1b[0m", rgb, message) 
}
