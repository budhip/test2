package rule

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	major int
	minor int
	patch int
}

func NewVersion(versionStr string) (Version, error) {
	parts := strings.Split(versionStr, ".")

	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version format: must be x.y.z")
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %w", err)
	}

	if major < 0 || minor < 0 || patch < 0 {
		return Version{}, fmt.Errorf("version numbers cannot be negative")
	}

	return Version{
		major: major,
		minor: minor,
		patch: patch,
	}, nil
}

type VersionBumpType string

const (
	BumpMajor VersionBumpType = "major"
	BumpMinor VersionBumpType = "minor"
	BumpPatch VersionBumpType = "patch"
)

func (v Version) Bump(bumpType VersionBumpType) (Version, error) {
	switch bumpType {
	case BumpMajor:
		return Version{major: v.major + 1, minor: 0, patch: 0}, nil
	case BumpMinor:
		return Version{major: v.major, minor: v.minor + 1, patch: 0}, nil
	case BumpPatch:
		return Version{major: v.major, minor: v.minor, patch: v.patch + 1}, nil
	default:
		return v, fmt.Errorf("invalid bump type: %s", bumpType)
	}
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

func (v Version) IsGreaterThan(other Version) bool {
	if v.major != other.major {
		return v.major > other.major
	}
	if v.minor != other.minor {
		return v.minor > other.minor
	}
	return v.patch > other.patch
}

func (v Version) Equals(other Version) bool {
	return v.major == other.major && v.minor == other.minor && v.patch == other.patch
}
