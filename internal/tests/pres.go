package tests

import (
	"os"
	"testing"
	"time"

	"github.com/carlwr/cobraman/internal/tests/fjoin"
)

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

func Preserve(t *testing.T, dir string, cfg PreserveCfg, invokedAt time.Time) {

	sinceAlways := (cfg.Policy == P_Always)
	sinceFailin := (cfg.Policy == P_Failing) && t.Failed()
	doPreserve := sinceAlways || sinceFailin

	if doPreserve {
		var err error

		toDir, err := fjoin.Join(
			cfg.Dir,
			invokedAt.Format("Mon_150405.0000"),
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
