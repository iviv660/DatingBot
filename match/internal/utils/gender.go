package utils

import "strings"

func OppositeGender(g string) string {
	switch strings.ToLower(g) {
	case "Девушка":
		return "Парень"
	case "Парень":
		return "Девушка"
	default:
		return ""
	}
}
