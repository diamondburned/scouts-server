package unmarshal

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestText(t *testing.T) {
	v, err := Text[*unmarshalableText]("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", v.text)
}

type unmarshalableText struct {
	text string
}

func (t *unmarshalableText) UnmarshalText(text []byte) error {
	t.text = string(text)
	return nil
}
