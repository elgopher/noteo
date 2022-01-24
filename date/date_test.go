package date_test

import (
	"testing"
	"time"

	"github.com/elgopher/noteo/date"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAbsolute(t *testing.T) {
	t.Run("should parse RFC2822", func(t *testing.T) {
		given := "Thu, 15 Oct 2020 16:30:10 +0200"
		// when
		ti, err := date.ParseAbsolute(given)
		// then
		require.NoError(t, err)
		year, month, day := ti.Date()
		assert.Equal(t, 2020, year)
		assert.Equal(t, time.October, month)
		assert.Equal(t, 15, day)
		assert.Equal(t, 16, ti.Hour())
		assert.Equal(t, 30, ti.Minute())
		assert.Equal(t, 10, ti.Second())
		_, offset := ti.Zone()
		assert.Equal(t, 60*60*2, offset)
	})
	t.Run("should parse ISO8601", func(t *testing.T) {
		given := "2020-10-15 16:30:10 +0200"
		// when
		ti, err := date.ParseAbsolute(given)
		// then
		require.NoError(t, err)
		year, month, day := ti.Date()
		assert.Equal(t, 2020, year)
		assert.Equal(t, time.October, month)
		assert.Equal(t, 15, day)
		assert.Equal(t, 16, ti.Hour())
		assert.Equal(t, 30, ti.Minute())
		assert.Equal(t, 10, ti.Second())
		_, offset := ti.Zone()
		assert.Equal(t, 60*60*2, offset)
	})
	t.Run("should parse strict ISO8601", func(t *testing.T) {
		given := "2020-10-15T16:30:10+02:00"
		// when
		ti, err := date.ParseAbsolute(given)
		// then
		require.NoError(t, err)
		year, month, day := ti.Date()
		assert.Equal(t, 2020, year)
		assert.Equal(t, time.October, month)
		assert.Equal(t, 15, day)
		assert.Equal(t, 16, ti.Hour())
		assert.Equal(t, 30, ti.Minute())
		assert.Equal(t, 10, ti.Second())
		_, offset := ti.Zone()
		assert.Equal(t, 60*60*2, offset)
	})
}

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
