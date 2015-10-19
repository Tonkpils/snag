package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tempLoc() string {
	return filepath.Join(os.TempDir(), "snag/test/stuff")
}

func createFiles(t *testing.T, names []string) {
	for _, name := range names {
		f, err := os.Create(filepath.Join(tempLoc(), name))
		require.NoError(t, err, "Error creating temp file %s", err)
		f.Close()
	}
}

func deleteFiles(t *testing.T, names []string) {
	for _, name := range names {
		err := os.Remove(filepath.Join(tempLoc(), name))
		require.NoError(t, err, "Error removing temp file %s", err)
	}
}

func TestGlobMatch(t *testing.T) {
	v := tempLoc()
	require.NoError(t, os.MkdirAll(v, 0777))
	defer os.RemoveAll(os.TempDir() + "/snag")

	// blank line
	p := ""
	assert.False(t, globMatch(p, v), "Expected %q to NOT match %q", p, v)

	// a comment
	p = "#a comment"
	assert.False(t, globMatch(p, v), "Expected %q to NOT match %q", p, v)

	// regular match no slash
	p = "gitglob.go"
	assert.True(t, globMatch(p, "gitglob.go"), "Expected %q to match %q", p, v)

	// negation no slash
	p = "!gitglob.go"
	assert.False(t, globMatch(p, "gitglob.go"), "Expected %q to NOT match %q", p, v)

	// match with slash
	p = tempLoc() + "/foo.txt"

	tmpFiles := []string{"foo.txt"}
	createFiles(t, tmpFiles)
	assert.True(t, globMatch(p, v+"/foo.txt"), "Expected %q to match %q", p, v)
	deleteFiles(t, tmpFiles)

	// negate match with slash
	p = "!" + tempLoc() + "/foo.txt"

	tmpFiles = []string{"foo.txt"}
	createFiles(t, tmpFiles)
	assert.False(t, globMatch(p, v+"/foo.txt"), "Expected %q to NOT match %q", p, v)
	deleteFiles(t, tmpFiles)

	// directory
	p = tempLoc()
	assert.True(t, globMatch(p, v), "Expected %q to match %q", p, v)

	// directory with trailing slash
	p = tempLoc() + "/"
	assert.True(t, globMatch(p, v), "Expected %q to match %q", p, v)

	// star matching
	p = tempLoc() + "/*.txt"
	tmpFiles = []string{"foo.txt"}
	createFiles(t, tmpFiles)
	assert.True(t, globMatch(p, v+"/foo.txt"), "Expected %q to match %q", p, v)
	assert.False(t, globMatch(p, v+"/somedir/foo.txt"), "Expected %q to NOT match %q", p, v)
	deleteFiles(t, tmpFiles)
}
