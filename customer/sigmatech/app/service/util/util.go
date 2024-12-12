package util

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func ValidatePassword(password string, hashedPassword string) bool {
	// Comparing the password with the hash
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func GenerateHash(password string) (string, error) {
	// Generate "hash" to store from user password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return string(hash), nil
}

func Int(v int) *int { return &v }

func Boolean(v bool) *bool { return &v }

func UnwrapInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func ConvertInterfaceToJSON(jsob []byte) map[string]interface{} {
	var data map[string]interface{}
	err := json.Unmarshal(jsob, &data)
	fmt.Println(err)
	return data
}

// ContainsElement checks if a value is present in a slice.
func ContainsElement(needle interface{}, haystack interface{}) bool {
	haystackValue := reflect.ValueOf(haystack)

	if haystackValue.Kind() != reflect.Slice {
		return false
	}

	for i := 0; i < haystackValue.Len(); i++ {
		if reflect.DeepEqual(haystackValue.Index(i).Interface(), needle) {
			return true
		}
	}

	return false
}

func ExtractConstraintName(errMsg string) string {
	startIndex := strings.Index(errMsg, "constraint \"") + len("constraint \"")
	endIndex := strings.Index(errMsg[startIndex:], "\"") + startIndex

	if startIndex > -1 && endIndex > startIndex {
		return errMsg[startIndex:endIndex]
	}

	return ""
}

// HandleForeignKeyViolation formats the error message for foreign key constraint violation
func HandleForeignKeyViolation(err error, tableName string) (string, bool) {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
		// Regular expression to find the numeric value within parentheses
		re := regexp.MustCompile(`\((\d+)\)`)

		// Find all matches in the input text
		matches := re.FindAllStringSubmatch(pqErr.Detail, -1)

		// Extract the numeric value from the first match
		if len(matches) > 0 && len(matches[0]) > 1 {
			errorMsgTemplate := "The %s with ID %s is referenced by the %s with the referenced ID %s. You need to delete the %s with the referenced ID first before deleting this %s."
			errorMsg := fmt.Sprintf(errorMsgTemplate, tableName, matches[0][1], pqErr.Table, matches[0][1], pqErr.Table, tableName)
			return errorMsg, true
		}

		return "", false
	}
	return "", false
}

// Convert struct to map return error if any
func StructToMap(obj interface{}) map[string]interface{} {
	objType := reflect.TypeOf(obj)
	objValue := reflect.ValueOf(obj)

	data := make(map[string]interface{})

	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		value := objValue.Field(i)
		data[field.Name] = value.Interface()
	}

	return data
}

// Convert string to int and check if it is a valid int
func StringToInt(str string) int {
	var result int
	var err error

	if str != "" {
		result, err = strconv.Atoi(str)
		if err != nil {
			result = 0
		}
	}

	return result
}
