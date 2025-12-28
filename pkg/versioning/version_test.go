package versioning

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAPIVersion_String(t *testing.T) {
	tests := []struct {
		name    string
		version APIVersion
		want    string
	}{
		{"V1", V1, "v1"},
		{"V2", V2, "v2"},
		{"V3", V3, "v3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.version.String())
		})
	}
}

func TestAPIVersion_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		version APIVersion
		want    bool
	}{
		{"V1 is valid", V1, true},
		{"V2 is valid", V2, true},
		{"V3 is valid", V3, true},
		{"Invalid version", APIVersion("v99"), false},
		{"Empty version", APIVersion(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.version.IsValid())
		})
	}
}

func TestNewVersionedRoute(t *testing.T) {
	vr := NewVersionedRoute("ichi-go", V1, "auth")

	assert.Equal(t, V1, vr.Version)
	assert.Equal(t, "auth", vr.Domain)
	assert.Equal(t, "ichi-go", vr.ServiceName)
}

func TestVersionedRoute_BuildPath(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		version     APIVersion
		domain      string
		want        string
	}{
		{
			name:        "V1 auth path",
			serviceName: "ichi-go",
			version:     V1,
			domain:      "auth",
			want:        "/ichi-go/api/v1/auth",
		},
		{
			name:        "V2 user path",
			serviceName: "myapp",
			version:     V2,
			domain:      "user",
			want:        "/myapp/api/v2/user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := NewVersionedRoute(tt.serviceName, tt.version, tt.domain)
			assert.Equal(t, tt.want, vr.BuildPath())
		})
	}
}

func TestVersionedRoute_BuildPathWithEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		version     APIVersion
		domain      string
		endpoint    string
		want        string
	}{
		{
			name:        "Login endpoint",
			serviceName: "ichi-go",
			version:     V1,
			domain:      "auth",
			endpoint:    "login",
			want:        "/ichi-go/api/v1/auth/login",
		},
		{
			name:        "Nested endpoint",
			serviceName: "myapp",
			version:     V2,
			domain:      "user",
			endpoint:    "profile/settings",
			want:        "/myapp/api/v2/user/profile/settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := NewVersionedRoute(tt.serviceName, tt.version, tt.domain)
			assert.Equal(t, tt.want, vr.BuildPathWithEndpoint(tt.endpoint))
		})
	}
}

func TestVersionedRoute_Group(t *testing.T) {
	e := echo.New()
	vr := NewVersionedRoute("ichi-go", V1, "auth")
	group := vr.Group(e)

	assert.NotNil(t, group)
	// Echo group prefix includes trailing slash
	assert.Equal(t, "/ichi-go/api/v1/auth", group.Prefix)
}

func TestBuildAPIPath(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		version     string
		domain      string
		want        string
	}{
		{
			name:        "Standard path",
			serviceName: "ichi-go",
			version:     "v1",
			domain:      "auth",
			want:        "/ichi-go/api/v1/auth",
		},
		{
			name:        "V2 path",
			serviceName: "myapp",
			version:     "v2",
			domain:      "payment",
			want:        "/myapp/api/v2/payment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildAPIPath(tt.serviceName, tt.version, tt.domain)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    APIVersion
		wantErr bool
	}{
		{
			name:    "Valid V1",
			input:   "v1",
			want:    V1,
			wantErr: false,
		},
		{
			name:    "Valid V2",
			input:   "v2",
			want:    V2,
			wantErr: false,
		},
		{
			name:    "Invalid version",
			input:   "v99",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAllVersions(t *testing.T) {
	versions := AllVersions()
	assert.Len(t, versions, 3)
	assert.Contains(t, versions, V1)
	assert.Contains(t, versions, V2)
	assert.Contains(t, versions, V3)
}

func TestGetLatestVersion(t *testing.T) {
	latest := GetLatestVersion()
	assert.Equal(t, V1, latest)
}
