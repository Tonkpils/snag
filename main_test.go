package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnagInit(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	defer log.SetOutput(os.Stdout)

	wd, tmpDir := tmpDirectory(t)
	defer os.RemoveAll(tmpDir)

	chdir(t, tmpDir)
	defer os.Chdir(wd)

	_, err := os.Stat(SnagFile)
	assert.True(t, os.IsNotExist(err))

	err = initSnag()
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "Success")

	_, err = os.Stat(SnagFile)
	require.NoError(t, err)

	err = initSnag()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "file already exists")
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
