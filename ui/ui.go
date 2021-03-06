/*
Package ui is the endive subpackage that implements the UserInterface.

It aims at providing functions to present data to the user and get the
necessary user input.

Displaying large amounts of data relies on `less`, editing large texts relies
on `$EDITOR`, falling back to `nano` if the variable is not found.

*/
package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

const (
	editOrKeep    = "[E]dit or [K]eep current value: "
	enterNewValue = "Enter new value: "
	invalidChoice = "Invalid choice."
	emptyValue    = "Empty value detected."
	notConfirmed  = "Manual entry not confirmed, trying again."
	tooManyErrors = "Too many errors, giving up."
	userAborted   = "User aborted."

	// LocalTag an option to show it's the value in current database
	LocalTag = "[current] "
	// OnlineTag an option to show it's from GR
	OnlineTag = "[online/new] "
)

// UI implements endive.UserInterface
type UI struct {
	// Logger provides a logger to both stdout and a log file (for debug).
	logger *logging.Logger
	// LogFile is the pointer to the log file, to be closed by the main function.
	logFile *os.File
}

// RemoveDuplicates in []string
func RemoveDuplicates(options *[]string, otherStringsToClean ...string) {
	found := make(map[string]bool)
	// specifically remove other strings from values
	for _, o := range otherStringsToClean {
		found[o] = true
	}
	j := 0
	for i, x := range *options {
		if !found[x] && x != "" {
			found[x] = true
			(*options)[j] = (*options)[i]
			j++
		}
	}
	*options = (*options)[:j]
}

// SelectOption among several, or input a new one, and return user input.
func (ui UI) SelectOption(title, usage string, options []string, longField bool) (string, error) {
	ui.SubPart(title)
	if usage != "" {
		fmt.Println(ui.Green(usage))
	}

	// remove duplicates from options and display them
	RemoveDuplicates(&options)
	for i, o := range options {
		fmt.Printf("%d. %s\n", i+1, o)
	}

	var choice string
	errs := 0
	validChoice := false
	for !validChoice {
		if len(options) == 0 {
			ui.Choice("Leave [B]lank, [E]dit manually, or [A]bort: ")
		} else if len(options) > 1 {
			ui.Choice("Choose option [1-%d], leave [B]lank, [E]dit manually, or [A]bort: ", len(options))
		} else {
			ui.Choice("Choose [1], leave [B]lank, [E]dit manually, or [A]bort: ")
		}
		choice, scanErr := ui.GetInput()
		if scanErr != nil {
			return "", scanErr
		}

		if strings.ToUpper(choice) == "E" {
			var edited string
			var scanErr error
			if longField {
				allVersions := ""
				for i, o := range options {
					allVersions += fmt.Sprintf("--- %d ---\n%s\n", i+1, ui.unTag(o))
				}
				edited, scanErr = ui.Edit(allVersions)
			} else {
				ui.Choice(enterNewValue)
				edited, scanErr = ui.GetInput()
			}
			if scanErr != nil {
				return "", scanErr
			}
			if edited == "" {
				ui.Warning(emptyValue)
			}
			confirmed := ui.Accept("Confirm: " + edited)
			if confirmed {
				return edited, nil
			}
			ui.Warning(notConfirmed)
		} else if strings.ToUpper(choice) == "A" {
			return "", errors.New(userAborted)
		} else if strings.ToUpper(choice) == "B" {
			return "", nil
		} else if index, err := strconv.Atoi(choice); err == nil && 0 < index && index <= len(options) {
			return ui.unTag(options[index-1]), nil
		}

		if !validChoice {
			ui.Warning(invalidChoice)
			errs++
			if errs > 10 {
				ui.Warning(tooManyErrors)
				return "", errors.New(invalidChoice)
			}
		}
	}
	return choice, nil
}

// UpdateValue with user input
func (ui UI) UpdateValue(field, usage, oldValue string, longField bool) (newValue string, err error) {
	ui.SubPart("Modifying " + field)
	if usage != "" {
		ui.Info(ui.Green(usage)) // TODO ui.Info dans SelectOption aussi!
	}
	fmt.Printf("Current value: %s\n", oldValue)

	validChoice := false
	errs := 0
	for !validChoice {
		ui.Choice(editOrKeep)
		choice, scanErr := ui.GetInput()
		if scanErr != nil {
			return newValue, scanErr
		}
		switch strings.ToLower(choice) {
		case "e":
			var choice string
			var scanErr error
			if longField {
				choice, scanErr = ui.Edit(oldValue)
			} else {
				ui.Choice(enterNewValue)
				choice, scanErr = ui.GetInput()
			}
			if scanErr != nil {
				return newValue, scanErr
			}
			if choice == "" {
				ui.Warning(emptyValue)
			}
			if ui.Accept("Confirm") {
				newValue = choice
				validChoice = true
			} else {
				ui.Warning(notConfirmed)
				ui.Choice(editOrKeep)
			}
		case "k":
			newValue = oldValue
			validChoice = true
		default:
			ui.Warning(invalidChoice)
			errs++
			if errs > 10 {
				return "", errors.New(invalidChoice)
			}
		}
	}
	return strings.TrimSpace(newValue), nil
}

// GetInput from user
func (ui UI) GetInput() (string, error) {
	scanner := bufio.NewReader(os.Stdin)
	choice, scanErr := scanner.ReadString('\n')
	return strings.TrimSpace(choice), scanErr
}

// Accept asks a question and returns the answer
func (ui UI) Accept(question string) bool {
	fmt.Printf(ui.BlueBold("%s Y/N : "), question)
	input, err := ui.GetInput()
	if err == nil {
		switch input {
		case "y", "Y", "yes":
			return true
		}
	}
	return false
}

// Display text through a pager if necessary.
func (ui UI) Display(output string) {
	// -e Causes less to automatically exit the second time it reaches end-of-file.
	// -F or --quit-if-one-screen  Causes less to automatically exit if the entire file can be displayed on the first screen.
	// -Q Causes totally "quiet" operation: the terminal bell is never rung.
	// -X or --no-init Disables sending the termcap initialization and deinitialization strings to the terminal. This is sometimes desirable if the deinitialization string does something unnecessary, like clearing the screen.
	cmd := exec.Command("less", "-e", "-F", "-Q", "-X", "--buffers=-1")
	if err := runCommand(cmd, output); err != nil {
		ui.Error(err.Error())
	}
}

func runCommand(cmd *exec.Cmd, output string) error {
	r, stdin := io.Pipe()
	cmd.Stdin = r
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// create a blocking chan, run the pager and unblock once it is finished
	c := make(chan error)
	go func() {
		defer close(c)
		c <- cmd.Run()
	}()
	// create a channel to write output to less, because for large amounts of
	// lines, it's blocking until less displays what was already sent
	d := make(chan error)
	go func() {
		defer close(d)
		_, err := io.WriteString(stdin, output)
		if err != nil {
			return
		}
		d <- err
	}()
	// either wait for write to finish or for user to quit less
	select {
	case <-d:
		// if write is finished, wait for user to quit
	case <-c:
		// if user quits first, stop writing
		close(d)
	}
	// close stdin
	if err := stdin.Close(); err != nil {
		return err
	}
	// wait for the pager to be finished
	return <-c
}

// Edit long value using external $EDITOR
func (ui *UI) Edit(oldValue string) (string, error) {
	// create temp file
	content := []byte(oldValue)
	tmpfile, err := ioutil.TempFile("", "edit")
	if err != nil {
		return oldValue, err
	}
	// clean up
	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// write input string inside
	if _, err := tmpfile.Write(content); err != nil {
		return oldValue, err
	}
	if err := tmpfile.Close(); err != nil {
		return oldValue, err
	}

	// find $EDITOR
	editor := os.Getenv("EDITOR")
	if editor == "" {
		ui.Warning("$EDITOR not set, falling back to nano")
		editor = "nano"
	}

	// open it with $EDITOR
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return oldValue, err
	}

	// read file back, set output string
	newContent, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		return oldValue, err
	}
	return strings.TrimSpace(string(newContent)), nil
}

// Tag an entry local or online
func (ui *UI) Tag(entry string, isLocal bool) string {
	if isLocal {
		return ui.CyanBold(LocalTag) + entry
	}
	return ui.YellowBold(OnlineTag) + entry
}

// unTag strings tagged with Tag.
func (ui *UI) unTag(option string) string {
	out := option
	out = strings.Replace(out, ui.CyanBold(LocalTag), "", -1)
	out = strings.Replace(out, ui.YellowBold(OnlineTag), "", -1)
	return strings.TrimSpace(out)
}
