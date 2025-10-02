package main

import (
	"encoding/json"
	"net/http"
)

// VersionInfoResponse describes the API version capabilities
type VersionInfoResponse struct {
	APIVersion                string                              `json:"apiVersion"`
	SupportedResourceVersions map[string][]string                 `json:"supportedResourceVersions"`
	DefaultVersions           map[string]string                   `json:"defaultVersions"`
	VersionDetails            map[string]map[string]VersionDetail `json:"versionDetails"`
}

// VersionDetail provides metadata about a specific version
type VersionDetail struct {
	Version    string   `json:"version"`
	Stability  string   `json:"stability"`
	Deprecated bool     `json:"deprecated"`
	Package    string   `json:"package"`
	Transforms []string `json:"transforms,omitempty"`
}

// GetVersionInfo returns information about supported API versions
func GetVersionInfo(w http.ResponseWriter, r *http.Request) {
	if versionRegistry == nil {
		respondError(w, http.StatusInternalServerError, &ErrMsg{Message: "version registry not initialized"})
		return
	}

	kinds := versionRegistry.ListKinds()

	supportedVersions := make(map[string][]string)
	defaultVersions := make(map[string]string)
	versionDetails := make(map[string]map[string]VersionDetail)

	for _, kind := range kinds {
		versions := versionRegistry.ListVersions(kind)
		supportedVersions[kind] = versions
		defaultVersions[kind] = versionRegistry.GetDefaultVersion(kind)

		// Get detailed information for each version
		versionInfo := versionRegistry.GetVersionInfo(kind)
		kindDetails := make(map[string]VersionDetail)

		for version, metadata := range versionInfo {
			kindDetails[version] = VersionDetail{
				Version:    metadata.Version,
				Stability:  metadata.Stability,
				Deprecated: metadata.Deprecated,
				Package:    metadata.Package,
				Transforms: metadata.Transforms,
			}
		}

		versionDetails[kind] = kindDetails
	}

	response := &VersionInfoResponse{
		APIVersion:                "inventory/v1",
		SupportedResourceVersions: supportedVersions,
		DefaultVersions:           defaultVersions,
		VersionDetails:            versionDetails,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ErrMsg is a simple error message wrapper
type ErrMsg struct {
	Message string `json:"message"`
}

func (e *ErrMsg) Error() string {
	return e.Message
}
