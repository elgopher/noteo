package date_test

import (
	"testing"
	"time"

	"github.com/jacekolszak/noteo/date"
	"github.com/stretchr/testify/assert"
)

func TestFormatRFC2822(t *testing.T) {
	given := time.Date(2020, 10, 15, 16, 30, 10, 30, time.FixedZone("CEST", 60*60*2))
	// when
	f := date.FormatRFC2822(given)
	// then
	assert.Equal(t, "Thu, 15 Oct 2020 16:30:10 +0200", f)
}

func TestFormatISO8601(t *testing.T) {
	given := time.Date(2020, 10, 15, 16, 30, 10, 30, time.FixedZone("CEST", 60*60*2))
	// when
	f := date.FormatISO8601(given)
	// then
	assert.Equal(t, "2020-10-15 16:30:10 +0200", f)
}
