package api_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTSNode(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	t.Cleanup(func() {
		ta.shutdown(t)

		paths, err := filepath.Glob("ts_test/node/*.js")
		require.NoError(t, err)

		for _, path := range paths {
			os.Remove(path)
		}

		os.Remove("ts_test/node/client.ts")
	})

	ctx := context.Background()

	tc, err := ta.pyc.TSClient(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, tc)

	err = os.WriteFile("ts_test/node/client.ts", []byte(tc), 0o600)
	require.NoError(t, err)

	runNoError(t, "ts_test/node", nil, "tsc", "--pretty")

	paths, err := filepath.Glob("ts_test/node/*_test.js")
	require.NoError(t, err)

	for _, path := range paths {
		path := path

		t.Run(
			path,
			func(t *testing.T) {
				t.Parallel()
				testPath(t, path)
			},
		)
	}
}

func testPath(t *testing.T, path string) {
	ta := newTestAPI(t)
	defer ta.shutdown(t)

	env := map[string]string{
		"NODE_DEBUG":       os.Getenv("NODE_DEBUG"),
		"NODE_NO_WARNINGS": "1",
		"BASE_URL":         ta.baseURL,
	}

	runNoError(t, "ts_test/node", env, "node", "--test", "--enable-source-maps", filepath.Base(path))

	require.NoError(t, os.Remove(path))
}
