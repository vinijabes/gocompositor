package runnable

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

//XVFB ...
type XVFB interface {
	Start() error
	Stop()
}

//XVFBImpl ...
type XVFBImpl struct {
	displayID string
	process   *exec.Cmd
	commands  []string
}

//XVFBAlreadyRunningError ...
type XVFBAlreadyRunningError struct {
	displayID string
}

func (e XVFBAlreadyRunningError) Error() string {
	return e.displayID + ": " + "XVFB already running"
}

//Start ...
func (xvfb XVFBImpl) Start() error {
	if xvfb.checkIsUp() {
		return &XVFBAlreadyRunningError{xvfb.displayID}
	}

	log.Println("Starting xvfb...")
	xvfb.process = exec.Command("Xvfb", xvfb.commands...)
	go func() {
		var out bytes.Buffer
		var stderr bytes.Buffer

		xvfb.process.Stdout = &out
		xvfb.process.Stderr = &stderr

		err := xvfb.process.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		}

		fmt.Println("Result: " + out.String())
	}()

	time.Sleep(2 * time.Second)

	if !xvfb.checkIsUp() {
		return errors.New("Can't start xvfb")
	}

	log.Println("xvfb started")

	return nil
}

//Stop ...
func (xvfb XVFBImpl) Stop() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	xvfb.process.Process.Kill()
	xvfb.process = nil
	log.Println("xvfb terminated")
}

func (xvfb XVFBImpl) checkIsUp() bool {
	start := time.Now()
	for {
		os.Setenv("DISPLAY", xvfb.displayID)
		cmd := exec.Command("xset", "q")
		_, err := cmd.Output()

		if err == nil {
			return true
		}

		if time.Now().Sub(start).Seconds() > 3 {
			return false
		}

		time.Sleep(1 * time.Second)
	}
}

//NewXVFB ...
func NewXVFB(id string) XVFB {
	return &XVFBImpl{
		displayID: id,
		commands:  []string{id, "-screen", "0", "1920x1080x24"},
	}
}
