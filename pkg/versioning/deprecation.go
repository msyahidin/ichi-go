package versioning

import (
	"fmt"
	"time"
)

// DeprecationInfo contains information about a deprecated API version
type DeprecationInfo struct {
	Version            APIVersion
	DeprecatedAt       time.Time
	SunsetDate         time.Time
	ReplacementVersion APIVersion
	Message            string
}

// IsDeprecated checks if the version is currently deprecated
func (d *DeprecationInfo) IsDeprecated() bool {
	return time.Now().After(d.DeprecatedAt)
}

// IsSunset checks if the version has reached its sunset date
func (d *DeprecationInfo) IsSunset() bool {
	return time.Now().After(d.SunsetDate)
}

// DaysUntilSunset returns the number of days until the sunset date
func (d *DeprecationInfo) DaysUntilSunset() int {
	if d.IsSunset() {
		return 0
	}
	return int(time.Until(d.SunsetDate).Hours() / 24)
}

// GetWarningMessage returns a formatted warning message
func (d *DeprecationInfo) GetWarningMessage() string {
	if d.IsSunset() {
		return fmt.Sprintf("API version %s has been sunset. Please use %s", d.Version, d.ReplacementVersion)
	}

	if d.IsDeprecated() {
		daysLeft := d.DaysUntilSunset()
		return fmt.Sprintf(
			"API version %s is deprecated and will be sunset in %d days. Please migrate to %s. %s",
			d.Version,
			daysLeft,
			d.ReplacementVersion,
			d.Message,
		)
	}

	return ""
}

// DeprecationSchedule holds the deprecation information for all versions
// This should be configured based on your deprecation policy
var DeprecationSchedule = map[APIVersion]*DeprecationInfo{
	// Example: Uncomment when you want to deprecate V1
	// V1: {
	// 	Version:            V1,
	// 	DeprecatedAt:       time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
	// 	SunsetDate:         time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
	// 	ReplacementVersion: V2,
	// 	Message:            "V1 API will be removed on December 1, 2024. Please migrate to V2.",
	// },
}

// GetDeprecationInfo returns deprecation info for a version
func GetDeprecationInfo(version APIVersion) (*DeprecationInfo, bool) {
	info, exists := DeprecationSchedule[version]
	return info, exists
}

// IsVersionDeprecated checks if a version is deprecated
func IsVersionDeprecated(version APIVersion) bool {
	if info, exists := DeprecationSchedule[version]; exists {
		return info.IsDeprecated()
	}
	return false
}

// IsVersionSunset checks if a version is sunset
func IsVersionSunset(version APIVersion) bool {
	if info, exists := DeprecationSchedule[version]; exists {
		return info.IsSunset()
	}
	return false
}

// AddDeprecation adds a deprecation entry to the schedule
// This is useful for dynamic configuration
func AddDeprecation(info *DeprecationInfo) error {
	if !info.Version.IsValid() {
		return fmt.Errorf("invalid version: %s", info.Version)
	}

	if !info.ReplacementVersion.IsValid() {
		return fmt.Errorf("invalid replacement version: %s", info.ReplacementVersion)
	}

	if info.SunsetDate.Before(info.DeprecatedAt) {
		return fmt.Errorf("sunset date must be after deprecation date")
	}

	DeprecationSchedule[info.Version] = info
	return nil
}

// RemoveDeprecation removes a deprecation entry from the schedule
func RemoveDeprecation(version APIVersion) {
	delete(DeprecationSchedule, version)
}

// GetAllDeprecations returns all deprecation information
func GetAllDeprecations() map[APIVersion]*DeprecationInfo {
	// Return a copy to prevent modification
	result := make(map[APIVersion]*DeprecationInfo, len(DeprecationSchedule))
	for k, v := range DeprecationSchedule {
		result[k] = v
	}
	return result
}
