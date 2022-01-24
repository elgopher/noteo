package tag_test

import (
	"testing"
	"time"

	"github.com/elgopher/noteo/date"
	"github.com/elgopher/noteo/tag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("should return error for invalid tag", func(t *testing.T) {
		names := []string{
			"", " ", "foo bar", "\t", "foo\tbar", "\n", "foo\n",
		}
		for _, name := range names {
			t.Run(name, func(t *testing.T) {
				newTag, err := tag.New(name)
				assert.Error(t, err)
				assert.Equal(t, tag.Tag{}, newTag)
			})
		}
	})
}

func TestName(t *testing.T) {
	t.Run("diacritic characters", func(t *testing.T) {
		tg, err := tag.New("ąę:1")
		require.NoError(t, err)
		assert.Equal(t, "ąę", tg.Name())
	})
}

func TestTag_MakeDateAbsolute(t *testing.T) {
	tests := map[string]struct {
		tag         string
		expectedTag string
	}{
		"now": {
			tag:         "deadline:now",
			expectedTag: "deadline:2020-09-10T16:30:11+02:00",
		},
		"today": {
			tag:         "deadline:today",
			expectedTag: "deadline:2020-09-10",
		},
	}
	date.SetNow(func() time.Time {
		return time.Date(2020, 9, 10, 16, 30, 11, 0, time.FixedZone("CEST", 60*60*2))
	})
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tg, err := tag.New(test.tag)
			require.NoError(t, err)
			tg, err = tg.MakeDateAbsolute()
			require.NoError(t, err)
			expectedTag, err := tag.New(test.expectedTag)
			require.NoError(t, err)
			assert.Equal(t, expectedTag, tg)
		})
	}
}
