package versioning

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeprecationInfo_IsDeprecated(t *testing.T) {
	tests := []struct {
		name         string
		deprecatedAt time.Time
		want         bool
	}{
		{
			name:         "Deprecated in the past",
			deprecatedAt: time.Now().Add(-24 * time.Hour),
			want:         true,
		},
		{
			name:         "Not yet deprecated",
			deprecatedAt: time.Now().Add(24 * time.Hour),
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &DeprecationInfo{
				DeprecatedAt: tt.deprecatedAt,
			}
			assert.Equal(t, tt.want, info.IsDeprecated())
		})
	}
}

func TestDeprecationInfo_IsSunset(t *testing.T) {
	tests := []struct {
		name       string
		sunsetDate time.Time
		want       bool
	}{
		{
			name:       "Sunset in the past",
			sunsetDate: time.Now().Add(-24 * time.Hour),
			want:       true,
		},
		{
			name:       "Not yet sunset",
			sunsetDate: time.Now().Add(24 * time.Hour),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &DeprecationInfo{
				SunsetDate: tt.sunsetDate,
			}
			assert.Equal(t, tt.want, info.IsSunset())
		})
	}
}

func TestDeprecationInfo_DaysUntilSunset(t *testing.T) {
	tests := []struct {
		name       string
		sunsetDate time.Time
		wantMin    int
		wantMax    int
	}{
		{
			name:       "30 days until sunset",
			sunsetDate: time.Now().Add(30 * 24 * time.Hour),
			wantMin:    29,
			wantMax:    30,
		},
		{
			name:       "Already sunset",
			sunsetDate: time.Now().Add(-24 * time.Hour),
			wantMin:    0,
			wantMax:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &DeprecationInfo{
				SunsetDate: tt.sunsetDate,
			}
			days := info.DaysUntilSunset()
			assert.GreaterOrEqual(t, days, tt.wantMin)
			assert.LessOrEqual(t, days, tt.wantMax)
		})
	}
}

func TestDeprecationInfo_GetWarningMessage(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		info       *DeprecationInfo
		wantEmpty  bool
		wantSubstr string
	}{
		{
			name: "Sunset version",
			info: &DeprecationInfo{
				Version:            V1,
				DeprecatedAt:       now.Add(-60 * 24 * time.Hour),
				SunsetDate:         now.Add(-1 * time.Hour),
				ReplacementVersion: V2,
				Message:            "Please migrate",
			},
			wantEmpty:  false,
			wantSubstr: "has been sunset",
		},
		{
			name: "Deprecated version",
			info: &DeprecationInfo{
				Version:            V1,
				DeprecatedAt:       now.Add(-30 * 24 * time.Hour),
				SunsetDate:         now.Add(30 * 24 * time.Hour),
				ReplacementVersion: V2,
				Message:            "Please migrate",
			},
			wantEmpty:  false,
			wantSubstr: "is deprecated",
		},
		{
			name: "Not deprecated yet",
			info: &DeprecationInfo{
				Version:            V1,
				DeprecatedAt:       now.Add(30 * 24 * time.Hour),
				SunsetDate:         now.Add(90 * 24 * time.Hour),
				ReplacementVersion: V2,
				Message:            "Please migrate",
			},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.info.GetWarningMessage()
			if tt.wantEmpty {
				assert.Empty(t, msg)
			} else {
				assert.NotEmpty(t, msg)
				assert.Contains(t, msg, tt.wantSubstr)
			}
		})
	}
}

func TestAddDeprecation(t *testing.T) {
	// Clean up after test
	defer func() {
		DeprecationSchedule = make(map[APIVersion]*DeprecationInfo)
	}()

	now := time.Now()

	tests := []struct {
		name    string
		info    *DeprecationInfo
		wantErr bool
	}{
		{
			name: "Valid deprecation",
			info: &DeprecationInfo{
				Version:            V1,
				DeprecatedAt:       now,
				SunsetDate:         now.Add(90 * 24 * time.Hour),
				ReplacementVersion: V2,
				Message:            "Please migrate",
			},
			wantErr: false,
		},
		{
			name: "Invalid version",
			info: &DeprecationInfo{
				Version:            APIVersion("v99"),
				DeprecatedAt:       now,
				SunsetDate:         now.Add(90 * 24 * time.Hour),
				ReplacementVersion: V2,
				Message:            "Please migrate",
			},
			wantErr: true,
		},
		{
			name: "Invalid replacement version",
			info: &DeprecationInfo{
				Version:            V1,
				DeprecatedAt:       now,
				SunsetDate:         now.Add(90 * 24 * time.Hour),
				ReplacementVersion: APIVersion("v99"),
				Message:            "Please migrate",
			},
			wantErr: true,
		},
		{
			name: "Sunset before deprecated",
			info: &DeprecationInfo{
				Version:            V1,
				DeprecatedAt:       now.Add(90 * 24 * time.Hour),
				SunsetDate:         now,
				ReplacementVersion: V2,
				Message:            "Please migrate",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AddDeprecation(tt.info)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, DeprecationSchedule[tt.info.Version])
			}
		})
	}
}

func TestGetDeprecationInfo(t *testing.T) {
	// Setup
	now := time.Now()
	testInfo := &DeprecationInfo{
		Version:            V1,
		DeprecatedAt:       now,
		SunsetDate:         now.Add(90 * 24 * time.Hour),
		ReplacementVersion: V2,
		Message:            "Test message",
	}
	DeprecationSchedule[V1] = testInfo

	// Clean up after test
	defer func() {
		DeprecationSchedule = make(map[APIVersion]*DeprecationInfo)
	}()

	tests := []struct {
		name    string
		version APIVersion
		want    bool
	}{
		{
			name:    "Existing deprecation",
			version: V1,
			want:    true,
		},
		{
			name:    "Non-existent deprecation",
			version: V2,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, exists := GetDeprecationInfo(tt.version)
			assert.Equal(t, tt.want, exists)
			if exists {
				assert.NotNil(t, info)
			}
		})
	}
}

func TestIsVersionDeprecated(t *testing.T) {
	// Setup
	now := time.Now()
	DeprecationSchedule[V1] = &DeprecationInfo{
		Version:            V1,
		DeprecatedAt:       now.Add(-24 * time.Hour),
		SunsetDate:         now.Add(90 * 24 * time.Hour),
		ReplacementVersion: V2,
	}

	// Clean up after test
	defer func() {
		DeprecationSchedule = make(map[APIVersion]*DeprecationInfo)
	}()

	tests := []struct {
		name    string
		version APIVersion
		want    bool
	}{
		{
			name:    "Deprecated version",
			version: V1,
			want:    true,
		},
		{
			name:    "Non-deprecated version",
			version: V2,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsVersionDeprecated(tt.version)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestRemoveDeprecation(t *testing.T) {
	// Setup
	now := time.Now()
	DeprecationSchedule[V1] = &DeprecationInfo{
		Version:            V1,
		DeprecatedAt:       now,
		SunsetDate:         now.Add(90 * 24 * time.Hour),
		ReplacementVersion: V2,
	}

	// Verify it exists
	_, exists := GetDeprecationInfo(V1)
	assert.True(t, exists)

	// Remove it
	RemoveDeprecation(V1)

	// Verify it's gone
	_, exists = GetDeprecationInfo(V1)
	assert.False(t, exists)
}

func TestGetAllDeprecations(t *testing.T) {
	// Setup
	now := time.Now()
	DeprecationSchedule[V1] = &DeprecationInfo{
		Version:            V1,
		DeprecatedAt:       now,
		SunsetDate:         now.Add(90 * 24 * time.Hour),
		ReplacementVersion: V2,
	}

	// Clean up after test
	defer func() {
		DeprecationSchedule = make(map[APIVersion]*DeprecationInfo)
	}()

	all := GetAllDeprecations()
	assert.Len(t, all, 1)
	assert.NotNil(t, all[V1])

	// Verify it returns a copy (modifying it shouldn't affect original)
	all[V2] = &DeprecationInfo{Version: V2}
	assert.Len(t, DeprecationSchedule, 1)
}
