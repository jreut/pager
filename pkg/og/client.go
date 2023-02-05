// Package og is an OpsGenie client.
package og

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/jreut/pager/v2/pkg/interval"
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
func NewHTTPClient(domain, key string, debug bool) httpclient {
	return httpclient{
		domain: domain,
		key:    key,
		debug:  debug,
	}
	// https://api.opsgenie.com/v2/schedules/:scheduleIdentifier/overrides
	// https://api.eu.opsgenie.com/
}

type httpclient struct {
	domain, key string
	debug       bool
}

func (c httpclient) url(path string, query url.Values) string {
	return (&url.URL{
		Scheme:   "https",
		Host:     c.domain,
		Path:     path,
		RawQuery: query.Encode(),
	}).String()
}

func (c httpclient) Post(ctx context.Context, path string, data interface{}) (*http.Response, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.url(path, nil), &buf)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("", c.key)
	req.Header.Add("Content-Type", "application/json")
	if c.debug {
		out, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, fmt.Errorf("could not dump request: %w", err)
		}
		log.Println(string(out))
	}
	res, err := http.DefaultClient.Do(req)
	if c.debug {
		out, err := httputil.DumpResponse(res, true)
		if err != nil {
			return nil, fmt.Errorf("could not dump response: %w", err)
		}
		log.Println(string(out))
	}
	return res, err
}
func (c httpclient) Get(ctx context.Context, path string, query url.Values) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.url(path, query), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("", c.key)
	if c.debug {
		out, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, fmt.Errorf("could not dump request: %w", err)
		}
		log.Println(string(out))
	}
	res, err := http.DefaultClient.Do(req)
	if c.debug {
		out, err := httputil.DumpResponse(res, true)
		if err != nil {
			return nil, fmt.Errorf("could not dump response: %w", err)
		}
		log.Println(string(out))
	}
	return res, err
}

// Apply implements [./pkg/apply/cmd.Destination]
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
		res, err := c.Post(ctx, fmt.Sprintf("/v2/schedules/%s/overrides", schedule), data)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusCreated {
			return fmt.Errorf("oops")
		}
	}
	return nil
}

// GetTimeline returns this schedule's shifts between the times specified.
//
// See: https://docs.opsgenie.com/docs/schedule-api#get-schedule-timeline
func (c httpclient) GetTimeline(ctx context.Context, schedule string, from, to time.Time) ([]save.Interval, error) {
	// The API expects intervals in the form of "X (days|weeks|months) after Y date", which is absolutely ridiculous.
	// Instead of expecting the rest of our program to carry information like that, we translate the given [from,to) interval.
	// We roughly convert the given interval to a number of weeks, rounding up.
	d := to.Sub(from)
	d = d.Round(7 * 24 * time.Hour)
	weeks := int(d.Hours()/24/7) + 1
	data := url.Values{}
	data.Add("interval", fmt.Sprint(weeks))
	data.Add("intervalUnit", "weeks")
	data.Add("date", strftime(from))
	log.Printf("converted [%s,%s) to '%s plus %d weeks'",
		strftime(from),
		strftime(to),
		strftime(from),
		weeks,
	)
	res, err := c.Get(ctx, fmt.Sprintf("/v2/schedules/%s/timeline", schedule), data)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: try again with -debug=true", res.Status)
	}
	var out timeline
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	intervals, err := out.intervals(schedule)
	if err != nil {
		return nil, err
	}
	if err := interval.WriteCSV(log.Writer(), intervals); err != nil {
		return nil, fmt.Errorf("writing csv")
	}
	return interval.Flatten(intervals), nil
}

func strftime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func strptime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

type timeline struct {
	Data struct {
		FinalTimeline struct {
			Rotations []struct {
				Periods []struct {
					StartDate string `json:"startDate"`
					EndDate   string `json:"endDate"`
					Recipient struct {
						Name string `json:"name"`
					} `json:"recipient"`
				} `json:"periods"`
			} `json:"rotations"`
		} `json:"finalTimeline"`
	} `json:"data"`
}

func (t timeline) intervals(schedule string) ([]save.Interval, error) {
	var out []save.Interval
	for i, r := range t.Data.FinalTimeline.Rotations {
		for j, p := range r.Periods {
			s, err := strptime(p.StartDate)
			if err != nil {
				return nil, fmt.Errorf("parsing start time at rotation %d, period %d: %w", i, j, err)
			}
			e, err := strptime(p.EndDate)
			if err != nil {
				return nil, fmt.Errorf("parsing end time at rotation %d, period %d: %w", i, j, err)
			}
			out = append(out, save.Interval{
				Person:    p.Recipient.Name,
				Schedule:  schedule,
				StartAt:   s,
				EndBefore: e,
				Kind:      save.IntervalKindShift,
			})
		}
	}
	return out, nil
}
