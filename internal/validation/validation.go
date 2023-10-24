package validation

import (
	"errors"
	"regexp"
)

func IsValidGUID(guid string) bool {
	if guid == "" {
		return false
	}
	re := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	return re.MatchString(guid)
}

func validateEntityFields(fields map[string]interface{}) error {
	guid, ok := fields["guid"].(string)
	if !ok || !IsValidGUID(guid) {
		return errors.New("Invalid or missing GUID")
	}
	// Добавьте здесь другие проверки, если они нужны
	return nil
}
