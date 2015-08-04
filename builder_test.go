package main

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestNewBuilder(t *testing.T) {
	b, err := NewBuilder(nil, nil, nil, nil, false)
	assert.NoError(t, err)
	assert.NotNil(t, b)
}
