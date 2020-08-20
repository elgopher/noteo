package tag_test

import (
	"testing"

	"github.com/jacekolszak/noteo/tag"
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
				tagg, err := tag.New(name)
				assert.Error(t, err)
				assert.Equal(t, tag.Tag(""), tagg)
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
