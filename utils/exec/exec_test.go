package exec

import (
	"os/exec"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestRunCommand(t *testing.T) {
	hook := test.NewGlobal()
	log.SetLevel(log.DebugLevel)
	defer log.SetLevel(log.InfoLevel)

	message, err := RunCommand("echo", CmdOpts{}, "hello world")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", message)

	assert.Len(t, hook.Entries, 2)
}

func TestRunInDir(t *testing.T) {
	cmd := exec.Command("pwd")
	// Mock out the directory to something
	cmd.Dir = "/"

	message, err := RunCommandExt(cmd, CmdOpts{})
	assert.Nil(t, err)
	assert.Equal(t, "/", message)
}
