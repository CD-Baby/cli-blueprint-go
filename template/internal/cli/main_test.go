package cli

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// TestMain registers the mycli binary as a testscript command so scripts under
// testdata/script can invoke `mycli ...` in-process against a temp directory.
func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"mycli": func() { os.Exit(Execute()) },
	})
}

func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
	})
}
