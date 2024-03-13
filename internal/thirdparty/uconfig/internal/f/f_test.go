package f

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalerStringSlice(t *testing.T) {
	expect := TextUnmarshalerStringSlice{"a", "b", "c"}
	value := TextUnmarshalerStringSlice{}

	err := value.UnmarshalText([]byte("a.b.c"))
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expect, value)
}
