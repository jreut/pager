package og

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/jreut/pager/v2/pkg/assert"
	"github.com/jreut/pager/v2/pkg/interval"
)

//go:embed fixtures/timeline.json
var fixtures []byte

func TestTimeline(t *testing.T) {
	var timeline timeline
	err := json.Unmarshal(fixtures, &timeline)
	assert.Nil(t, err)

	parsed, err := timeline.intervals("schedule-id")
	assert.Nil(t, err)
	assert.Nil(t, interval.WriteCSV(
		assert.Golden(t, "parsed.csv"),
		parsed,
	))
	assert.Nil(t, interval.WriteCSV(
		assert.Golden(t, "flattened.csv"),
		interval.Flatten(parsed),
	))
}
