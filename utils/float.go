package utils

import (
	"fmt"
)

func FormatPercent(val float64) string {
	return fmt.Sprintf("%.2f", val)
}
