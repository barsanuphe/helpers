package helpers

import (
	"fmt"
	"time"

	"github.com/tj/go-spin"

	i "github.com/barsanuphe/helpers/ui"
)

// TimeTrack helps track the time taken by a function.
func TimeTrack(ui i.UserInterface, start time.Time, name string) {
	elapsed := time.Since(start)
	ui.Debugf("-- %s in %s\n", name, elapsed)
}

//SpinWhileThingsHappen is a way to launch a function and display a spinner while it is being executed.
func SpinWhileThingsHappen(title string, f func() error) (err error) {
	c1 := make(chan bool)
	c2 := make(chan error)

	// first routine for the spinner
	ticker := time.NewTicker(time.Millisecond * 100)
	go func() {
		for range ticker.C {
			c1 <- true
		}
	}()
	// second routine deals with the function
	go func() {
		// run function
		c2 <- f()
	}()

	// await both of these values simultaneously,
	// dealing with each one as it arrives.
	functionDone := false
	s := spin.New()
	for !functionDone {
		select {
		case <-c1:
			fmt.Printf("\r%s... %s ", title, s.Next())
		case err := <-c2:
			if err != nil {
				fmt.Printf("\r%s... KO.\n", title)
				return err
			}
			fmt.Printf("\r%s... Done.\n", title)
			functionDone = true
		}
	}
	return
}

// CheckErrors and return the first non-nil one.
func CheckErrors(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
