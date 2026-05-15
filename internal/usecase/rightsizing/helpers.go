package rightsizing

import (
	"strconv"
	"strings"
)

// formatCPU formats CPU cores to Kubernetes resource string
func formatCPU(cores float64) string {
	if cores == 0 {
		return ""
	}
	// Express in millicores if less than 1
	if cores < 1 {
		return strconv.Itoa(int(cores*1000)) + "m"
	}
	// Round to 2 decimal places for cores
	return strconv.FormatFloat(cores, 'f', 2, 64)
}

// formatMemory formats bytes to Kubernetes resource string
func formatMemory(bytes float64) string {
	if bytes == 0 {
		return ""
	}
	// Convert to MiB
	mib := bytes / (1024 * 1024)
	if mib < 1024 {
		return strconv.Itoa(int(mib)) + "Mi"
	}
	// Convert to GiB
	gib := mib / 1024
	return strconv.FormatFloat(gib, 'f', 2, 64) + "Gi"
}

// parseCPU parses Kubernetes CPU resource string to cores
func parseCPU(s string) float64 {
	if s == "" {
		return 0
	}

	s = strings.TrimSpace(s)

	// Handle millicores
	if strings.HasSuffix(s, "m") {
		val, err := strconv.ParseFloat(strings.TrimSuffix(s, "m"), 64)
		if err != nil {
			return 0
		}
		return val / 1000
	}

	// Handle cores
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}

// parseMemory parses Kubernetes memory resource string to bytes
func parseMemory(s string) float64 {
	if s == "" {
		return 0
	}

	s = strings.TrimSpace(s)

	// Suffix mappings to bytes
	suffixes := map[string]float64{
		"Ki": 1024,
		"Mi": 1024 * 1024,
		"Gi": 1024 * 1024 * 1024,
		"Ti": 1024 * 1024 * 1024 * 1024,
		"K":  1000,
		"M":  1000 * 1000,
		"G":  1000 * 1000 * 1000,
		"T":  1000 * 1000 * 1000 * 1000,
	}

	for suffix, multiplier := range suffixes {
		if strings.HasSuffix(s, suffix) {
			val, err := strconv.ParseFloat(strings.TrimSuffix(s, suffix), 64)
			if err != nil {
				return 0
			}
			return val * multiplier
		}
	}

	// No suffix - assume bytes
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}
