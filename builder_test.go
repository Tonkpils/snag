package main

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestNewBuilder(t *testing.T) {
	b, err := NewBuilder(config{})
	assert.NoError(t, err)
	assert.NotNil(t, b)
}
