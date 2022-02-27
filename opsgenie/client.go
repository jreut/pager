package opsgenie

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/og"
	"github.com/opsgenie/opsgenie-go-sdk-v2/schedule"
)

// Client is a more reasonable wrapper around the subset of things we
// want to do with the OpsGenie API.
type Client struct{ schedules *schedule.Client }

// NewClient initializes a new OpsGenie API client with the given API key.
//
// To get a key, go to https://cockroachlabs.app.opsgenie.com/settings/integration/add/API/
func NewClient(key string) (out *Client, err error) {
	cfg := client.Config{ApiKey: key}
	out = new(Client)
	out.schedules, err = schedule.NewClient(&cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return out, nil
}

// EnsureSchedule fetches the schedule by name. If the schedule does not
// exist, it creates it in the given team.
func (self *Client) EnsureSchedule(
	ctx context.Context, name, team string,
) (*schedule.Schedule, error) {
	s, err := self.schedules.Get(ctx, &schedule.GetRequest{
		IdentifierType:  schedule.Name,
		IdentifierValue: name,
	})
	if err != nil {
		if err, ok := err.(*client.ApiError); ok {
			if err.StatusCode == 404 {
				res, err := self.schedules.Create(ctx, &schedule.CreateRequest{
					Name:        name,
					Description: "a test schedule for reuter@ to test an integration",
					OwnerTeam: &og.OwnerTeam{
						Name: team,
					},
				})
				if err != nil {
					return nil, errors.WithStack(err)
				}
				return &schedule.Schedule{
					Id:      res.Id,
					Name:    res.Name,
					Enabled: res.Enabled,
				}, nil
			} else {
				return nil, errors.WithStack(err)
			}
		} else {
			return nil, errors.WithStack(err)
		}
	}
	return &s.Schedule, nil
}

// Override creates an override on the given schedule for the given user
// across the given span of time.
func (self *Client) Override(
	ctx context.Context, s *schedule.Schedule, who string, from, to time.Time,
) error {
	_, err := self.schedules.CreateScheduleOverride(ctx, &schedule.CreateScheduleOverrideRequest{
		ScheduleIdentifierType: schedule.Id,
		ScheduleIdentifier:     s.Id,
		User: schedule.Responder{
			Type:     schedule.UserResponderType,
			Username: who,
		},
		StartDate: from,
		EndDate:   to,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
