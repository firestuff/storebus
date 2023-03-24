package api

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) getList(cfg *config, w http.ResponseWriter, r *http.Request) error {
	opts, err := parseListOpts(r)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrBadRequest, "parse list parameters failed (%w)", err)
	}

	list, err := api.listInt(r.Context(), cfg, opts)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "list failed (%w)", err)
	}

	etag, err := HashList(list)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "hash list failed (%w)", err)
	}

	if opts.IfNoneMatchETag != "" && opts.IfNoneMatchETag == etag {
		w.WriteHeader(http.StatusNotModified)
		return nil
	}

	err = jsrest.WriteList(w, list, etag)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "write list failed (%w)", err)
	}

	return nil
}
