package newsapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// Articles calls GET /v1/articles. The "source" query param is required by the API.
func (v *V1) Articles(ctx context.Context, params url.Values, opts ...RequestOption) (*ArticlesV1Result, http.Header, error) {
	ro := collectRequestOptions(opts)
	u := v.c.requestURL("/v1/articles", params)
	hdr, body, err := v.c.getJSON(ctx, u, true, ro)
	if err != nil {
		return nil, headOrNil(hdr, ro.showHeaders), err
	}
	var out ArticlesV1Result
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, headOrNil(hdr, ro.showHeaders), err
	}
	return &out, headOrNil(hdr, ro.showHeaders), nil
}

// Sources calls GET /v1/sources. All query parameters are optional. The reference Node
// client does not send an API key for this request; this method matches that behavior.
func (v *V1) Sources(ctx context.Context, params url.Values, opts ...RequestOption) (*SourcesV1Result, http.Header, error) {
	ro := collectRequestOptions(opts)
	u := v.c.requestURL("/v1/sources", params)
	hdr, body, err := v.c.getJSON(ctx, u, false, ro)
	if err != nil {
		return nil, headOrNil(hdr, ro.showHeaders), err
	}
	var out SourcesV1Result
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, headOrNil(hdr, ro.showHeaders), err
	}
	return &out, headOrNil(hdr, ro.showHeaders), nil
}
