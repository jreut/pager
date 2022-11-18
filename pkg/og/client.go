// Package og is an OpsGenie client.
package og

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jreut/pager/v2/pkg/save"
)

const (
	DomainDefault = "api.opsgenie.com"
	DomainEU      = "api.eu.opsgenie.com"
)

// NewHTTPClient constructs a [Client] of the OpsGenie HTTP API.
//
// You probably want to use [DomainDefault] as the domain.
//
// Create an "API Integration" to get a key.
//
// See: https://support.atlassian.com/opsgenie/docs/create-a-default-api-integration/
func NewHTTPClient(domain, key string) httpclient {
	return httpclient{
		domain: domain,
		key:    key,
	}
	// https://api.opsgenie.com/v2/schedules/:scheduleIdentifier/overrides
	// https://api.eu.opsgenie.com/
}

type httpclient struct {
	domain, key string
}

func (c httpclient) url(path string, query url.Values) string {
	return (&url.URL{
		Scheme:   "https",
		Host:     c.domain,
		Path:     path,
		RawQuery: query.Encode(),
	}).String()
}

func (c httpclient) post(ctx context.Context, path string, data interface{}) (*http.Response, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.url(path, nil), &buf)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("GenieKey", c.key)
	req.Header.Add("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}
func (c httpclient) get(ctx context.Context, path string, query url.Values) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.url(path, query), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("GenieKey", c.key)
	return http.DefaultClient.Do(req)
}

// Apply implements [Client]
//
// See: https://docs.opsgenie.com/docs/schedule-override-api#create-schedule-override
func (c httpclient) Apply(ctx context.Context, schedule string, shifts []save.Interval) error {
	for _, s := range shifts {
		data := map[string]interface{}{
			"user": map[string]string{
				"type":     "user",
				"username": s.Person,
			},
			"startDate": strftime(s.StartAt),
			"endDate":   strftime(s.EndBefore),
		}
		res, err := c.post(ctx, fmt.Sprintf("/v2/schedules/%s/overrides", schedule), data)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusCreated {
			return fmt.Errorf("oops")
		}
	}
	return nil
}

func strftime(t time.Time) string {
	return t.Format(time.RFC3339)
}
