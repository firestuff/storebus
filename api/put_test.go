package api_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/firestuff/patchy/patchyc"
	"github.com/stretchr/testify/require"
)

func TestReplace(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo", Num: 1})
	require.NoError(t, err)
	require.EqualValues(t, 0, created.Generation)

	replaced, err := patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
	require.NoError(t, err)
	require.NotNil(t, replaced)
	require.Equal(t, "bar", replaced.Text)
	require.EqualValues(t, 0, replaced.Num)
	require.EqualValues(t, 1, replaced.Generation)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "bar", get.Text)
	require.EqualValues(t, 0, get.Num)
	require.EqualValues(t, 1, get.Generation)
}

func TestReplaceIfMatchETagSuccess(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	ta.pyc.SetHeader("If-Match", fmt.Sprintf(`"%s"`, created.ETag))

	replaced, err := patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
	require.NoError(t, err)
	require.Equal(t, "bar", replaced.Text)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "bar", get.Text)
}

func TestReplaceIfMatchETagMismatch(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	ta.pyc.SetHeader("If-Match", `"etag:doesnotmatch"`)

	replaced, err := patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
	require.Error(t, err)
	require.ErrorContains(t, err, "etag mismatch")
	require.Nil(t, replaced)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "foo", get.Text)
}

func TestReplaceIfMatchGenerationSuccess(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	ta.pyc.SetHeader("If-Match", fmt.Sprintf(`"generation:%d"`, created.Generation))

	replaced, err := patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
	require.NoError(t, err)
	require.Equal(t, "bar", replaced.Text)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "bar", get.Text)
}

func TestReplaceIfMatchGenerationMismatch(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	ta.pyc.SetHeader("If-Match", `"generation:50"`)

	replaced, err := patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
	require.Error(t, err)
	require.ErrorContains(t, err, "generation mismatch")
	require.Nil(t, replaced)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "foo", get.Text)
}

func TestReplaceIfMatchInvalid(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	ta.pyc.SetHeader("If-Match", `"foobar"`)

	replaced, err := patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid If-Match")
	require.Nil(t, replaced)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "foo", get.Text)
}
