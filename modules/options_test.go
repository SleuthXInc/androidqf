// SleuthX fork (1.8.3-sleuthx3) — the unattended options backported from upstream #112.

package modules

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/mvt-project/androidqf/adb"
	"github.com/mvt-project/androidqf/utils"
)

func neverPrompt(t *testing.T) func() (string, error) {
	return func() (string, error) {
		t.Fatal("the interactive prompt ran although the answer was given on the command line")
		return "", nil
	}
}

func TestBackupFlagSkipsPrompt(t *testing.T) {
	opts := &Options{NonInteractive: true, Backup: "sms"}
	got, err := resolveOption(opts, backupFlagLabel(opts.Backup), "-backup", neverPrompt(t))
	if err != nil || got != backupOnlySMS {
		t.Fatalf("got %q, %v; want %q", got, err, backupOnlySMS)
	}
}

func TestDownloadFlagSkipsPrompt(t *testing.T) {
	opts := &Options{NonInteractive: true, Download: "non-system", RemoveTrusted: "no"}
	got, err := resolveOption(opts, downloadFlagLabel(opts.Download), "-download", neverPrompt(t))
	if err != nil || got != apkNotSystem {
		t.Fatalf("got %q, %v; want %q", got, err, apkNotSystem)
	}
	keep, err := resolveOption(opts, removeTrustedFlagLabel(opts.RemoveTrusted), "-remove-trusted", neverPrompt(t))
	if err != nil || keep != apkKeepAll {
		t.Fatalf("got %q, %v; want %q", keep, err, apkKeepAll)
	}
}

func TestNonInteractiveWithoutFlagFails(t *testing.T) {
	// Before acquisition: every missing answer is named.
	err := ValidateNonInteractive(&Options{NonInteractive: true}, "")
	if err == nil {
		t.Fatal("want an error for a -non-interactive run with no answers")
	}
	for _, want := range []string{"-backup", "-download", "-intrusion-logs"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name %s", err, want)
		}
	}
	// At the prompt itself: no prompt runs, and EOF is never taken as an answer.
	if _, err := resolveOption(&Options{NonInteractive: true}, "", "-backup", neverPrompt(t)); err == nil {
		t.Fatal("want an error when -non-interactive has no answer")
	}
	// A complete answer set passes; -remove-trusted is needed only when APKs are downloaded.
	full := &Options{NonInteractive: true, Backup: "sms", Download: "non-system", RemoveTrusted: "no", IntrusionLogs: "no"}
	if err := ValidateNonInteractive(full, ""); err != nil {
		t.Fatalf("complete answers rejected: %v", err)
	}
	if err := ValidateNonInteractive(&Options{NonInteractive: true, Backup: "none", Download: "none", IntrusionLogs: "no"}, ""); err != nil {
		t.Fatalf("download none should not need -remove-trusted: %v", err)
	}
}

func TestInvalidValueIsRefused(t *testing.T) {
	if err := ValidateNonInteractive(&Options{Backup: "everything"}, ""); err == nil {
		t.Fatal("want an error for -backup everything")
	}
}

func TestBackupTimeoutReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake adb is a shell script")
	}
	fake := filepath.Join(t.TempDir(), "adb")
	// A fake adb that, like a real one awaiting the on-device tap, never returns on its own.
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nexec sleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	client := &adb.ADB{ExePath: fake}
	start := time.Now()
	err := client.BackupWithTimeout(filepath.Join(t.TempDir(), "backup.ab"), "com.android.providers.telephony", 200*time.Millisecond)
	if !errors.Is(err, adb.ErrBackupTimedOut) {
		t.Fatalf("got %v; want ErrBackupTimedOut", err)
	}
	if !strings.Contains(err.Error(), "timed out waiting for on-device confirmation") {
		t.Fatalf("error text %q is not what the agent classifies as timed_out", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("the adb child was not killed at the timeout (took %s)", elapsed)
	}
}

func TestVersion(t *testing.T) {
	if utils.Version != "1.8.3-sleuthx3" {
		t.Fatalf("utils.Version = %q; want 1.8.3-sleuthx3", utils.Version)
	}
}
