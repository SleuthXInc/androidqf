// androidqf - Android Quick Forensics
// Copyright (c) 2026 kpcyrd.
// Use of this software is governed by the MVT License 1.1 that can be found at
//   https://license.mvt.re/1.1/

//go:build unbundle

package assets

import (
	"os"
	"path/filepath"
)

// SleuthX fork patch (android-forensic-provisioning, 2026-07-08):
// Upstream hardcodes the collector dir to /usr/lib/androidqf/android-collector,
// which is SIP-protected on macOS and absent on Windows — the two SleuthX
// desktop acquire targets. Honor an ANDROIDQF_COLLECTOR_DIR override so the
// desktop agent can point androidqf at the collectors it stages next to the
// binary (see android-acquire.ts androidqfChildEnv). Falls back to the upstream
// FHS path so a package-manager install (Linux) still works unchanged.
const defaultCollectorDir = "/usr/lib/androidqf/android-collector"

func collectorDir() string {
	if dir := os.Getenv("ANDROIDQF_COLLECTOR_DIR"); dir != "" {
		return dir
	}
	return defaultCollectorDir
}

// Read a specific collector binary from filesystem
func ReadCollectorFile(collectorName string) ([]byte, error) {
	path := filepath.Join(collectorDir(), collectorName)
	return os.ReadFile(path)
}

// Assets are expected to be installed by package manager.
func DeployAssets() (string, error) {
	return "", nil
}

// No assets to clean up
func CleanAssets() error {
	return nil
}
