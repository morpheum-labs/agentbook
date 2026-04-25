package newapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// TopHeadlines calls GET /v2/top-headlines. As in the Node client, a nil [url.Values] becomes
// `language=en` by default; an empty (non-nil) [url.Values] is sent as-is, which the API can reject
// if no filters are provided.
func (v *V2) TopHeadlines(ctx context.Context, params url.Values, opts ...RequestOption) (*TopHeadlinesResult, http.Header, error) {
	ro := collectRequestOptions(opts)
	p := applyTopHeadlinesParams(params)
	u := v.c.requestURL("/v2/top-headlines", p)
	hdr, body, err := v.c.getJSON(ctx, u, true, ro)
	if err != nil {
		return nil, headOrNil(hdr, ro.showHeaders), err
	}
	var out TopHeadlinesResult
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, headOrNil(hdr, ro.showHeaders), err
	}
	return &out, headOrNil(hdr, ro.showHeaders), nil
}

// Everything calls GET /v2/everything. At least one of `q`, `sources`, or `domains` (see NewsAPI docs) is required.
func (v *V2) Everything(ctx context.Context, params url.Values, opts ...RequestOption) (*EverythingResult, http.Header, error) {
	ro := collectRequestOptions(opts)
	u := v.c.requestURL("/v2/everything", params)
	hdr, body, err := v.c.getJSON(ctx, u, true, ro)
	if err != nil {
		return nil, headOrNil(hdr, ro.showHeaders), err
	}
	var out EverythingResult
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, headOrNil(hdr, ro.showHeaders), err
	}
	return &out, headOrNil(hdr, ro.showHeaders), nil
}

// Sources calls GET /v2/sources. All query parameters are optional.
func (v *V2) Sources(ctx context.Context, params url.Values, opts ...RequestOption) (*SourcesV2Result, http.Header, error) {
	ro := collectRequestOptions(opts)
	u := v.c.requestURL("/v2/sources", params)
	hdr, body, err := v.c.getJSON(ctx, u, true, ro)
	if err != nil {
		return nil, headOrNil(hdr, ro.showHeaders), err
	}
	var out SourcesV2Result
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, headOrNil(hdr, ro.showHeaders), err
	}
	return &out, headOrNil(hdr, ro.showHeaders), nil
}
