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
	assert.False(t, globMatch("", v))

	// a comment
	assert.False(t, globMatch("#a comment", v))

	// regular match no slash
	assert.True(t, globMatch("gitglob.go", "gitglob.go"))

	// negation no slash
	assert.False(t, globMatch("!gitglob.go", "gitglob.go"))

	// match with slash
	tmpFiles := []string{"foo.txt"}
	createFiles(t, tmpFiles)
	assert.True(t, globMatch(tempLoc()+"/foo.txt", v+"/foo.txt"))
	deleteFiles(t, tmpFiles)

	// negate match with slash
	tmpFiles = []string{"foo.txt"}
	createFiles(t, tmpFiles)
	assert.False(t, globMatch("!"+tempLoc()+"/foo.txt", v+"/foo.txt"))
	deleteFiles(t, tmpFiles)

	// directory
	assert.True(t, globMatch(tempLoc(), v))

	// directory with trailing slash
	assert.True(t, globMatch(tempLoc()+"/", v))

	// star matching
	tmpFiles = []string{"foo.txt"}
	createFiles(t, tmpFiles)
	assert.True(t, globMatch(tempLoc()+"/*.txt", v+"/foo.txt"))
	assert.False(t, globMatch(tempLoc()+"/*.txt", v+"/somedir/foo.txt"))
	deleteFiles(t, tmpFiles)

	// double star prefix
	assert.True(t, globMatch("**/foo.txt", v+"/hello/foo.txt"))
	assert.True(t, globMatch("**/foo.txt", v+"/some/dirs/foo.txt"))

	// double star suffix
	assert.True(t, globMatch(tempLoc()+"/hello/**", v+"/hello/foo.txt"))
	assert.False(t, globMatch(tempLoc()+"/hello/**", v+"/some/dirs/foo.txt"))

	// double star in path
	assert.True(t, globMatch(tempLoc()+"/hello/**/world.txt", v+"/hello/world.txt"))
	assert.True(t, globMatch(tempLoc()+"/hello/**/world.txt", v+"/hello/stuff/world.txt"))
	assert.False(t, globMatch(tempLoc()+"/hello/**/world.txt", v+"/some/dirs/foo.txt"))

	// negate doubl start patterns
	assert.False(t, globMatch("!**/foo.txt", v+"/hello/foo.txt"))
	assert.False(t, globMatch("!"+tempLoc()+"/hello/**", v+"/hello/foo.txt"))
	assert.False(t, globMatch("!"+tempLoc()+"/hello/**/world.txt", v+"/hello/world.txt"))
}
