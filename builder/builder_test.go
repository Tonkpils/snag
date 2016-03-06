package builder

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterfaceCompatibility(t *testing.T) {
	var _ Builder = &CmdBuilder{}
}

func TestNewBuilder(t *testing.T) {
	testEnv := "foobar"
	os.Setenv("TEST_ENV", testEnv)
	c := Config{
		Build:   []string{"echo Hello World", "echo $$TEST_ENV"},
		Run:     []string{"echo async here"},
		Verbose: true,
	}
	b := New(nil, c)
	require.NotNil(t, b)

	cb, ok := b.(*CmdBuilder)
	require.True(t, ok)

	require.Len(t, cb.buildCmds, 2)
	assert.Equal(t, c.Build[0], strings.Join(cb.buildCmds[0], " "))
	assert.Equal(t, testEnv, cb.buildCmds[1][1])

	require.Len(t, cb.runCmds, 1)
	assert.Equal(t, c.Run[0], strings.Join(cb.runCmds[0], " "))

	assert.Equal(t, c.Verbose, cb.verbose)
}

func TestNewBuilder_CmdWithQuotes(t *testing.T) {
	tests := []struct {
		Command string
		Chunks  []string
	}{
		{ // one single quote pair
			Command: `echo 'hello world' foo`,
			Chunks:  []string{`echo`, `'hello world'`, `foo`},
		},
		{ // one double quote pair
			Command: `echo "hello world" foo`,
			Chunks:  []string{`echo`, `"hello world"`, `foo`},
		},
		{ // no ending double quote
			Command: `echo "ga ga oh la la`,
			Chunks:  []string{`echo`, `"ga ga oh la la`},
		},
		{ // no ending single quote
			Command: `echo 'ga ga oh la la`,
			Chunks:  []string{`echo`, `'ga ga oh la la`},
		},
		{ // multiple double quotes
			Command: `echo "ga" "foo"`,
			Chunks:  []string{`echo`, `"ga"`, `"foo"`},
		},
		{ // double quotes inside single quotes
			Command: `echo -c 'foo "bar"'`,
			Chunks:  []string{`echo`, `-c`, `'foo "bar"'`},
		},
		{ // single quotes inside double quotes
			Command: `echo -c "foo 'bar'"`,
			Chunks:  []string{`echo`, `-c`, `"foo 'bar'"`},
		},
	}

	for _, test := range tests {
		c := Config{
			Build: []string{test.Command},
			Run:   []string{test.Command},
		}

		b := New(nil, c)
		require.NotNil(t, b)

		cb, ok := b.(*CmdBuilder)
		require.True(t, ok)

		assert.Equal(t, test.Chunks, cb.buildCmds[0])
		assert.Equal(t, test.Chunks, cb.runCmds[0])
	}
}
