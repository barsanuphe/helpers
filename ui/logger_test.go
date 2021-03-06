package ui

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUILogger(t *testing.T) {
	fmt.Println("+ Testing UI/GetLogger()...")
	logFilename := "../test/testing"
	ui := UI{}
	err := ui.getLogger(logFilename)
	assert.Nil(t, err)
	defer ui.CloseLog()

	ui.Error("Error")
	ui.Infof("Test %d%% complete", 50)
	// should not be displayed
	ui.Debug("Debug")

	// checking log file
	output, err := ioutil.ReadFile(logFilename)
	if os.IsNotExist(err) {
		t.Error("Error: cannot find log file")
		t.FailNow()
	} else if err != nil {
		t.Errorf("Error reading log file: %s", err.Error())
	}
	// checking 3 lines were written + 1 for setup + 1 return at end of file
	lines := strings.Split(string(output), "\n")
	assert.Equal(t, 5, len(lines), "Error checking log file: wrong number of lines")

	// remove log file
	err = os.Remove(logFilename)
	require.Nil(t, err, "Error removing test log file")
}
