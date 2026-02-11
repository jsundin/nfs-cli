package cli

import (
	"io"
	"strings"

	"github.com/peterh/liner"
)

type CommandLineReader_t interface {
	ReadCommand(prompt string) (string, error)
	io.Closer
}

type LinerBasedCommandLineReader struct {
	state *liner.State
}

type ArrayBasedCommandLineReader struct {
	args []string
}

func NewLinerBasedCommandLineReader(commands []string) *LinerBasedCommandLineReader {
	state := liner.NewLiner()
	state.SetCtrlCAborts(true)
	state.SetCompleter(func(line string) (c []string) {
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, line) {
				c = append(c, cmd)
			}
		}
		return c
	})

	return &LinerBasedCommandLineReader{state: state}
}

func (linerCli *LinerBasedCommandLineReader) ReadCommand(prompt string) (string, error) {
	line, err := linerCli.state.Prompt(prompt)
	if err == nil {
		linerCli.state.AppendHistory(line)
	}
	return line, err

}

func (linerCli *LinerBasedCommandLineReader) Close() error {
	return linerCli.state.Close()
}

func NewArrayBasedCommandLineReader(src []string) *ArrayBasedCommandLineReader {
	return &ArrayBasedCommandLineReader{args: src}
}

func (arrCli *ArrayBasedCommandLineReader) ReadCommand(string) (string, error) {
	if len(arrCli.args) == 0 {
		return "", io.EOF
	}
	out := arrCli.args[0]
	arrCli.args = arrCli.args[1:]
	return out, nil
}

func (arrCli *ArrayBasedCommandLineReader) Close() error {
	return nil
}
