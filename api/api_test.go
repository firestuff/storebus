package api_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dchest/uniuri"
	"github.com/firestuff/patchy"
	"github.com/firestuff/patchy/api"
	"github.com/firestuff/patchy/patchyc"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

type testAPI struct {
	api *api.API
	srv *http.Server
	rst *resty.Client
	pyc *patchyc.Client
}

func newTestAPI(t *testing.T) *testAPI {
	dbname := fmt.Sprintf("file:%s?mode=memory&cache=shared", uniuri.New())

	a, err := api.NewSQLiteAPI(dbname)
	require.Nil(t, err)

	api.Register[testType](a)

	mux := http.NewServeMux()
	// Test that prefix stripping works
	mux.Handle("/api/", http.StripPrefix("/api", a))

	listener, err := net.Listen("tcp", "[::]:0")
	require.Nil(t, err)

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 1 * time.Second,
	}

	go func() {
		_ = srv.Serve(listener)
	}()

	baseURL := fmt.Sprintf("http://[::1]:%d/api/", listener.Addr().(*net.TCPAddr).Port)

	rst := resty.New().
		SetHeader("Content-Type", "application/json").
		SetBaseURL(baseURL)

	pyc := patchyc.NewClient(baseURL)

	return &testAPI{
		api: a,
		srv: srv,
		rst: rst,
		pyc: pyc,
	}
}

func (ta *testAPI) r() *resty.Request {
	return ta.rst.R()
}

func (ta *testAPI) shutdown(t *testing.T) {
	err := ta.srv.Shutdown(context.Background())
	require.Nil(t, err)

	ta.api.Close()
}

func readEvent(scan *bufio.Scanner, out any) (string, error) {
	eventType := ""
	data := [][]byte{}

	for scan.Scan() {
		line := scan.Text()

		switch {
		case strings.HasPrefix(line, ":"):
			continue

		case strings.HasPrefix(line, "event: "):
			eventType = strings.TrimPrefix(line, "event: ")

		case strings.HasPrefix(line, "data: "):
			data = append(data, bytes.TrimPrefix(scan.Bytes(), []byte("data: ")))

		case line == "":
			var err error

			if out != nil {
				err = json.Unmarshal(bytes.Join(data, []byte("\n")), out)
			}

			return eventType, err
		}
	}

	return "", io.EOF
}

type testType struct {
	api.Metadata
	Text string `json:"text"`
	Num  int64  `json:"num"`
}

func TestCheckSafe(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	defer os.RemoveAll(dir)

	a, err := api.NewFileStoreAPI(dir)
	require.Nil(t, err)

	require.Nil(t, a.IsSafe())

	api.Register[testType](a)

	require.ErrorIs(t, a.IsSafe(), api.ErrMissingAuthCheck)

	require.Panics(t, func() {
		a.CheckSafe()
	})
}

func TestAccept(t *testing.T) {
	// TODO: Break up test
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	created := &testType{}

	resp, err := ta.r().
		SetBody(&testType{
			Text: "foo",
		}).
		SetResult(created).
		Post("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "foo", created.Text)
	require.NotEmpty(t, created.ID)

	read := &testType{}

	resp, err = ta.r().
		SetHeader("Accept", "text/event-stream;q=0.3, text/xml;q=0.1, application/json;q=0.5").
		SetResult(read).
		SetPathParam("id", created.ID).
		Get("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "application/json", resp.Header().Get("Content-Type"))
	require.Equal(t, "foo", read.Text)
	require.Equal(t, created.ID, read.ID)

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream;q=0.7, text/xml;q=0.1, application/json;q=0.5").
		SetPathParam("id", created.ID).
		Get("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "text/event-stream", resp.Header().Get("Content-Type"))
	resp.RawBody().Close()
}

func TestDebug(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	dbg := map[string]any{}

	resp, err := ta.r().
		SetResult(&dbg).
		Get("_debug")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Contains(t, dbg, "server")
	require.Contains(t, dbg, "ip")
	require.Contains(t, dbg, "http")
	require.Contains(t, dbg, "tls")
}

func TestRequestHookError(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	ta.api.SetRequestHook(func(*http.Request, *patchy.API) (*http.Request, error) {
		return nil, fmt.Errorf("test reject")
	})

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.Error(t, err)
	require.Nil(t, created)
}
