package main

import (
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig_CliCmds(t *testing.T) {
	args := []string{"foo", "bar"}
	cliCmds = argSlice(args)
	defer func() { cliCmds = nil }()

	c, err := parseConfig()
	require.NoError(t, err)
	assert.Equal(t, args, c.Build)
}

func TestParseConfig_NoSnagFile(t *testing.T) {
	wd, tmpDir := tmpDirectory(t)
	defer os.RemoveAll(tmpDir)

	chdir(t, tmpDir)
	defer os.Chdir(wd)

	_, err := parseConfig()
	require.Error(t, err)
	assert.Equal(t, "Could not find '.snag.yml' in your current directory", err.Error())
}

func TestParseConfig_FunkyYml(t *testing.T) {
	wd, tmpDir := tmpDirectory(t)
	defer os.RemoveAll(tmpDir)

	chdir(t, tmpDir)
	defer os.Chdir(wd)

	writeSnagFile(t, "I like to thing I'm yaml")
	_, err := parseConfig()
	require.Error(t, err)
	rx := regexp.MustCompile(`^Could not parse yml file\. .+`)
	assert.Regexp(t, rx, err.Error())
}

func TestParseConfig_ScriptAndBuild(t *testing.T) {
	wd, tmpDir := tmpDirectory(t)
	defer os.RemoveAll(tmpDir)

	chdir(t, tmpDir)
	defer os.Chdir(wd)

	writeSnagFile(t, "verbose: true\nbuild:\n  - echo 'hello'\nscript:\n  - echo 'hello'")
	_, err := parseConfig()
	require.Error(t, err)
	assert.Equal(t, "Cannot use 'script' and 'build' together. The 'script' tag is deprecated, please use 'build' instead.", err.Error())
}

func TestParseConfig_Script(t *testing.T) {
	wd, tmpDir := tmpDirectory(t)
	defer os.RemoveAll(tmpDir)

	chdir(t, tmpDir)
	defer os.Chdir(wd)

	writeSnagFile(t, "verbose: true\nscript:\n  - echo 'hello'")
	c, err := parseConfig()
	require.NoError(t, err)
	assert.Equal(t, []string{"echo 'hello'"}, c.Build)
}

func TestParseConfig_EmptyBuild(t *testing.T) {
	wd, tmpDir := tmpDirectory(t)
	defer os.RemoveAll(tmpDir)

	chdir(t, tmpDir)
	defer os.Chdir(wd)

	writeSnagFile(t, "verbose: true")
	_, err := parseConfig()
	require.Error(t, err)
	assert.Equal(t, "You must specify at least 1 command.", err.Error())
}

func TestParseConfig_Verbose(t *testing.T) {
	verbose = true
	defer func() { verbose = false }()

	wd, tmpDir := tmpDirectory(t)
	defer os.RemoveAll(tmpDir)

	chdir(t, tmpDir)
	defer os.Chdir(wd)

	writeSnagFile(t, "build:\n  - echo 'hello'")
	c, err := parseConfig()
	require.NoError(t, err)
	assert.True(t, c.Verbose, "verbosity was not set correctly")
}

func chdir(t *testing.T, path string) {
	err := os.Chdir(path)
	require.NoError(t, err, "could not change directories")
}

func tmpDirectory(t *testing.T) (string, string) {
	wd, err := os.Getwd()
	require.NoError(t, err, "could not get working directory")

	tmpDir, err := ioutil.TempDir("", strconv.FormatInt(time.Now().UnixNano(), 10))
	require.NoError(t, err, "could not create tmp directory")
	return wd, tmpDir
}

func writeSnagFile(t *testing.T, content string) {
	f, err := os.Create(".snag.yml")
	require.NoError(t, err, "could not create snag.yml")
	defer f.Close()

	_, err = f.WriteString(content)
	require.NoError(t, err, "could not write content to snag.yml")
}
