// androidqf - Android Quick Forensics
// Copyright (c) 2021-2026 Claudio Guarnieri.
// Use of this software is governed by the MVT License 1.1 that can be found at
//   https://license.mvt.re/1.1/

// SleuthX fork (1.8.3-sleuthx3): a backport of upstream androidqf #112's unattended options
// (mvt-project/androidqf@5522697, modules/options.go) onto the 1.8.3-sleuthx2 tree, which keeps the
// directory-tree output the SleuthX agent, MVT and ALEAPP read. Flag names and values match
// upstream so a later rebase changes no argv. One addition, -backup-timeout, has no upstream
// equivalent: an unanswered `adb backup` otherwise blocks the whole run.

package modules

import (
	"fmt"
	"strings"
	"time"
)

// Options carries the answers given on the command line. An empty string means no answer was
// given, so the module prompts — unless NonInteractive is set, in which case it is an error.
type Options struct {
	NonInteractive bool
	Backup         string // sms | all | none
	Download       string // all | non-system | none
	RemoveTrusted  string // yes | no
	IntrusionLogs  string // yes | no
	BackupTimeout  time.Duration
}

// CurrentOptions is set once by main before any module runs. A package variable rather than a
// new Module.Run parameter keeps this backport to the three modules that prompt.
var CurrentOptions = &Options{}

// DefaultBackupTimeout bounds the wait for the on-device backup confirmation.
const DefaultBackupTimeout = 120 * time.Second

var validValues = map[string][]string{
	"-backup":         {"sms", "all", "none"},
	"-download":       {"all", "non-system", "none"},
	"-remove-trusted": {"yes", "no"},
	"-intrusion-logs": {"yes", "no"},
}

func checkValue(flag, value string) error {
	if value == "" {
		return nil
	}
	for _, v := range validValues[flag] {
		if v == value {
			return nil
		}
	}
	return fmt.Errorf("invalid %s value %q (expected %s)", flag, value, strings.Join(validValues[flag], ", "))
}

// ValidateNonInteractive refuses, before any acquisition starts, a -non-interactive run that leaves
// a prompting module without its answer — so EOF on stdin can never stand in for one. Mirrors
// upstream ValidateNonInteractive for the modules this tree has.
func ValidateNonInteractive(opts *Options, moduleFilter string) error {
	for flag, value := range map[string]string{
		"-backup": opts.Backup, "-download": opts.Download,
		"-remove-trusted": opts.RemoveTrusted, "-intrusion-logs": opts.IntrusionLogs,
	} {
		if err := checkValue(flag, value); err != nil {
			return err
		}
	}
	if !opts.NonInteractive {
		return nil
	}
	enabled := func(name string) bool { return moduleFilter == "" || moduleFilter == name }
	var missing []string
	if enabled("backup") && opts.Backup == "" {
		missing = append(missing, "-backup (sms, all, none)")
	}
	if enabled("packages") {
		if opts.Download == "" {
			missing = append(missing, "-download (all, non-system, none)")
		} else if opts.Download != "none" && opts.RemoveTrusted == "" {
			missing = append(missing, "-remove-trusted (yes, no)")
		}
	}
	if enabled("intrusion_logs") && opts.IntrusionLogs == "" {
		missing = append(missing, "-intrusion-logs (yes, no)")
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("-non-interactive requires: %s", strings.Join(missing, ", "))
}

// resolveOption returns the command-line answer when there is one, refuses when -non-interactive is
// set without one, and otherwise asks the interactive prompt.
func resolveOption(opts *Options, value, flagUsage string, prompt func() (string, error)) (string, error) {
	if value != "" {
		return value, nil
	}
	if opts.NonInteractive {
		return "", fmt.Errorf("-non-interactive is set but %s was not provided", flagUsage)
	}
	return prompt()
}
