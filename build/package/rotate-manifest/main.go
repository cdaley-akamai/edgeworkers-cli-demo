// Command rotate-manifest updates devserver-manifest.json for a new EdgeWorkersDevServer release.
//
// It moves the existing "current" release into "previous" (skipping a zero-value
// placeholder release with no packages, and re-releases of the same version) and
// writes the new release as "current".
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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

	if m.Current.Version != "" && len(m.Current.Packages) > 0 && m.Current.Version != version {
		m.Previous = append([]release{m.Current}, m.Previous...)
	}
	m.Previous = removeVersion(m.Previous, version)
	m.Current = release{Version: version, ReleaseDate: releaseDate, Packages: packages}

	return writeManifest(manifestPath, m)
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
