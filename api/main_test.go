package api

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
)

type TestAgent struct{}

func (ta TestAgent) GetName() string {
	return "test agent"
}

// Test registering commands
func (ta TestAgent) RegisterCommands() map[string]Command {
	var cmds = make(map[string]Command)

	cmds["test"] = Command{
		Closure: func(r io.Writer, args ...string) ([][]string, error) {
			var resp = [][]string{
				{"TEST"},
			}

			return resp, nil
		},
		Definition: Definition{},
	}

	cmds["test_with_error"] = Command{
		Closure: func(r io.Writer, args ...string) ([][]string, error) {
			return nil, errors.New("Test error")
		},
		Definition: Definition{},
	}

	fmt.Print(cmds["test_with_error"].String())

	return cmds
}

var (
	API   = New("../tests/var/run/leprechaun/.sock")
	Agent = &TestAgent{}
)

func TestRegister(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)

	go API.Register(Agent)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-API.readyChan:
				API.Command("agent test")
				API.Command("agent test_with_error")
				API.Command("test")
				API.Command("agent not_exist")

				return
			}

		}
	}()

	wg.Wait()
}

func TestCall(t *testing.T) {
	API.commands = Agent.RegisterCommands()
	r, err := API.Call(os.Stdout, "test")
	if err != nil {
		t.Fail()
	}

	if r[0][0] != "TEST" {
		t.Fail()
	}

	_, err = API.Call(os.Stdout, "test_with_error")
	if err == nil {
		t.Fail()
	}
}

func TestGetCommands(t *testing.T) {
	cmds := API.GetCommands()
	if len(cmds) < 2 {
		t.Fail()
	}
}
