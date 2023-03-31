package dao

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_getInsolvencyPractitionersDetails(t *testing.T) {
	t.Parallel()

	t.Run("no practitioners to retrieve", func(t *testing.T) {
		got, err := getInsolvencyPractitionersDetails(nil, "12345", nil)
		assert.Equal(t, fmt.Errorf("no practitioners to retrieve"), err)
		assert.Nil(t, got)
	})
}
