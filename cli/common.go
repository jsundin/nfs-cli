package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/peterh/liner"
	"github.com/willscott/go-nfs-client/nfs"
)

const (
	displayTimeLayout = time.RFC3339
	inputTimeLayout   = "2006-01-02T15:04:05"
)

func formatTime(nfst nfs.NFS3Time) string {
	t := time.Unix(int64(nfst.Seconds), int64(nfst.Nseconds))
	return t.Format(displayTimeLayout)
}

func parseTimeOrDuration(v string) (time.Time, error) {
	if t, err := time.Parse(inputTimeLayout, v); err == nil {
		return t, nil
	}
	if d, err := time.ParseDuration(v); err == nil {
		return time.Now().Add(d), nil
	}
	return time.Unix(0, 0), fmt.Errorf("failed to parse input")
}

func parseMode(v string) (os.FileMode, error) {
	i, err := strconv.ParseInt(v, 0, 32)
	return os.FileMode(i), err
}

func stdinReader() io.Reader {
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		fmt.Println("Enter your file contents here. End with a '.' on an empty line.")
		l := liner.NewLiner()
		defer l.Close()
		defer pipeWriter.Close()
		for {
			input, err := l.Prompt(">> ")
			if input == "." {
				break
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "stdin: %v\n", err)
				break
			}
			if _, err := pipeWriter.Write([]byte(input + "\n")); err != nil {
				fmt.Fprintf(os.Stderr, "stdin: failed to write to pipe: %v\n", err)
				break
			}
		}
	}()
	return pipeReader
}

func runInShell(cmdline string) {
	cmd := exec.Command(os.Getenv("SHELL"), "-c", cmdline)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}
