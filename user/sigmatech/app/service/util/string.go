package util

import (
	"math/rand"
	"regexp"
)

func GenerateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func ReverseString(s string) string {
	// Convert the string to a slice of bytes
	strBytes := []byte(s)

	// Initialize two indices: one at the beginning and one at the end of the slice
	start := 0
	end := len(strBytes) - 1

	// Swap characters from the start and end towards the middle
	for start < end {
		strBytes[start], strBytes[end] = strBytes[end], strBytes[start]
		start++
		end--
	}

	// Convert the byte slice back to a string
	return string(strBytes)
}

// ContainsString checks if a string exists in a slice of strings.
func ContainsString(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// RemoveString removes a string from a slice of strings.
func RemoveString(slice []string, item string) []string {
	var result []string
	for _, v := range slice {
		if v != item {
			result = append(result, v)
		}
	}
	return result
}

// IsValidEmail checks if the email is valid or not.
func IsValidEmail(email string) bool {
	// Define a regular expression pattern for a valid email address
	// This pattern allows for a wide range of valid email formats
	// Note that email validation can be complex, and this pattern may not cover all cases.
	// For a more comprehensive solution, consider using a dedicated email validation library.
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// Use the regexp package to match the email against the pattern
	match, _ := regexp.MatchString(emailPattern, email)

	return match
}
