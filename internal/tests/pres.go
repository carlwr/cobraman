package tests

import (
	"os"
	"testing"
	"time"

	"github.com/carlwr/cobraman/internal/tests/fjoin"
)

// Returns a temporary directory for the test to use.
//
// If the cfg parameter specifies the _never_ policy, this function only calls [`(testing.T).TempDir`](https://pkg.go.dev/testing#T.TempDir), and is hence equivalent to that function.
//
// If the cfg parameter specifies the _always_ policy, this function additionally registers a cleanup function that copies the temporary directory to a permanent location, preserving it. The _failing_ policy is similar, but only preserves the directory if the test has failed.
//
// The `testInvokedAt` parameter can be provided as a pointer to a `time.Time` value to specify the time at which the test run was initiated. This is then used as part of the name of the preserved directory path, by which preserved directories will be sorted under one subdirectory per test run. If `nil` is provided, it will be ignored.
func TempDirWith(t *testing.T, cfg PreserveCfg, testInvokedAt *time.Time) string {

	tmpDir := t.TempDir()

	if cfg.Policy == P_Never {
		return tmpDir
	}

	invokedAtStr := ""
	if testInvokedAt != nil {
		invokedAtStr = (*testInvokedAt).Format("Mon_150405.0000")
	}
	t.Cleanup(func() { Preserve(t, tmpDir, cfg, invokedAtStr) })

	return tmpDir
}

func TempDirFunc(cfg PreserveCfg, testInvokedAt *time.Time) func(t *testing.T) string {
	fun := func(t *testing.T) string {
		return TempDirWith(t, cfg, testInvokedAt)
	}
	return fun
}

type Policy int

const (
	P_Failing Policy = iota
	P_Always
	P_Never
)

type PreserveCfg struct {
	Policy Policy
	Dir    string
}

// func Preserve(t *testing.T, dir string, cfg PreserveCfg, invokedAt time.Time) {
func Preserve(t *testing.T, dir string, cfg PreserveCfg, prefix string) {

	sinceAlways := (cfg.Policy == P_Always)
	sinceFailin := (cfg.Policy == P_Failing) && t.Failed()
	doPreserve := sinceAlways || sinceFailin

	if doPreserve {
		var err error

		toDir, err := fjoin.Join(
			cfg.Dir,
			// invokedAt.Format("Mon_150405.0000"),
			prefix,
			t.Name())
		if err != nil {
			t.Logf("WARNING: failed to filenamify:\n  '%v'", err)
			return
		}

		err = os.CopyFS(toDir, os.DirFS(dir))
		if err != nil {
			t.Logf("WARNING: failed to preserve:\n  '%v'", err)
			return
		}
		t.Logf("info: preserved temp dir:\n  %s", toDir)
	}
}
