package cli

import (
	"fmt"
	"os"

	"github.com/google/shlex"
	"github.com/spf13/cobra"
	"github.com/willscott/go-nfs-client/nfs"
)

var (
	target           *nfs.Target
	cwd              CWD
	commandProviders = []func() *cobra.Command{}
)

func Main(t *nfs.Target, mountPath string, args []string) error {
	target = t
	cwd = CWD{mountPath: mountPath, relativePath: "/"}

	var cliReader CommandLineReader_t
	if len(args) > 0 {
		cliReader = NewArrayBasedCommandLineReader(args)
	} else {
		tmpRootCommand := newRootCommand()
		commands := []string{}
		for _, cmd := range tmpRootCommand.Commands() {
			commands = append(commands, cmd.DisplayName())
		}
		cliReader = NewLinerBasedCommandLineReader(commands)
	}
	defer cliReader.Close()

	for {
		in, err := cliReader.ReadCommand(fmt.Sprintf("%s >> ", cwd.String()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return err
		}

		if len(in) > 0 && in[0] == '!' {
			runInShell(in[1:])
			continue
		}

		parsed, err := shlex.Split(in)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse command line: %v\n", err)
			return err
		}

		if len(parsed) == 0 {
			continue
		}

		if parsed[0] == "exit" {
			break
		}

		rootCommand := newRootCommand()
		rootCommand.SetArgs(parsed)
		if err = rootCommand.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to execute command: %v\n", err)
		}
	}

	return nil
}

func newRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
		Use:           "cli",
	}
	for _, provider := range commandProviders {
		rootCommand.AddCommand(provider())
	}
	return rootCommand
}
