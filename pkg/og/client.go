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

type Client interface {
	Apply(context.Context, []save.Interval) error
	List(_ context.Context, startAt, endBefore time.Time) ([]save.Interval, error)
}

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
func NewHTTPClient(domain, key, schedule string) Client {
	return httpclient{
		domain:   domain,
		key:      key,
		schedule: schedule,
	}
	// https://api.opsgenie.com/v2/schedules/:scheduleIdentifier/overrides
	// https://api.eu.opsgenie.com/
}

type httpclient struct {
	domain, key, schedule string
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
func (c httpclient) Apply(ctx context.Context, shifts []save.Interval) error {
	for _, s := range shifts {
		data := map[string]interface{}{
			"user": map[string]string{
				"type":     "user",
				"username": s.Person,
			},
			"startDate": strftime(s.StartAt),
			"endDate":   strftime(s.EndBefore),
		}
		res, err := c.post(ctx, fmt.Sprintf("/v2/schedules/%s/overrides", c.schedule), data)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusCreated {
			return fmt.Errorf("oops")
		}
	}
	return nil
}

// List implements [Client]
func (httpclient) List(_ context.Context, startAt time.Time, endBefore time.Time) ([]save.Interval, error) {
	panic("unimplemented")
}

func strftime(t time.Time) string {
	return t.Format(time.RFC3339)
}
