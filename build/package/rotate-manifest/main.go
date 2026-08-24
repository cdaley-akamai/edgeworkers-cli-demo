// Command rotate-manifest updates devserver-manifest.json for a new EdgeWorkersDevServer release.
//
// It moves the existing "current" release into "previous" (skipping a zero-value
// placeholder release with no packages, and re-releases of the same version) and
// writes the new release as "current". A release older than the current version is
// treated as a patch for an older line: it upserts into "previous" instead of
// replacing "current".
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type devPackage struct {
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

type release struct {
	Version     string       `json:"version"`
	ReleaseDate string       `json:"releaseDate"`
	Packages    []devPackage `json:"packages"`
}

type manifest struct {
	Current  release   `json:"current"`
	Previous []release `json:"previous"`
}

func main() {
	manifestPath := flag.String("manifest", "devserver-manifest.json", "path to devserver-manifest.json")
	version := flag.String("version", "", "new release version")
	releaseDate := flag.String("release-date", "", "new release date (YYYY-MM-DD)")
	packagesPath := flag.String("packages-file", "", "path to a JSON file containing the new release's packages array")
	flag.Parse()

	if *version == "" || *releaseDate == "" || *packagesPath == "" {
		fmt.Fprintln(os.Stderr, "version, release-date, and packages-file are required")
		os.Exit(1)
	}

	if err := run(*manifestPath, *version, *releaseDate, *packagesPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(manifestPath, version, releaseDate, packagesPath string) error {
	m, err := readManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}
	packages, err := readPackages(packagesPath)
	if err != nil {
		return fmt.Errorf("read packages: %w", err)
	}

	newRelease := release{Version: version, ReleaseDate: releaseDate, Packages: packages}
	hasCurrent := m.Current.Version != "" && len(m.Current.Packages) > 0

	switch {
	case !hasCurrent || m.Current.Version == version:
		// No release yet, or a re-release of the current version.
		m.Current = newRelease
	case compareVersions(version, m.Current.Version) > 0:
		// A newer release: demote the existing current into previous.
		m.Previous = append(removeVersion(m.Previous, version), m.Current)
		m.Current = newRelease
	default:
		// An older-line patch release: current is unaffected, upsert into previous.
		m.Previous = append(removeVersion(m.Previous, version), newRelease)
	}
	sortReleasesDescending(m.Previous)

	return writeManifest(manifestPath, m)
}

// sortReleasesDescending orders releases newest-first so "previous" stays consistent
// regardless of the order releases were made in.
func sortReleasesDescending(releases []release) {
	sort.Slice(releases, func(i, j int) bool {
		return compareVersions(releases[i].Version, releases[j].Version) > 0
	})
}

// compareVersions compares two dot-separated numeric version strings, returning a
// negative number if a < b, zero if equal, and a positive number if a > b.
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	for i := 0; i < len(aParts) || i < len(bParts); i++ {
		var aNum, bNum int
		if i < len(aParts) {
			aNum, _ = strconv.Atoi(aParts[i])
		}
		if i < len(bParts) {
			bNum, _ = strconv.Atoi(bParts[i])
		}
		if aNum != bNum {
			return aNum - bNum
		}
	}
	return 0
}

func readManifest(path string) (manifest, error) {
	var m manifest
	data, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, err
	}
	return m, nil
}

func readPackages(path string) ([]devPackage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var packages []devPackage
	if err := json.Unmarshal(data, &packages); err != nil {
		return nil, err
	}
	return packages, nil
}

// removeVersion drops any release matching version, so re-releasing a version doesn't duplicate it.
func removeVersion(releases []release, version string) []release {
	kept := make([]release, 0, len(releases))
	for _, r := range releases {
		if r.Version != version {
			kept = append(kept, r)
		}
	}
	return kept
}

func writeManifest(path string, m manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
