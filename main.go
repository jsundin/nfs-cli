package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/jsundin/nfs-cli/cli"
	"github.com/spf13/cobra"
	"github.com/willscott/go-nfs-client/nfs"
	"github.com/willscott/go-nfs-client/nfs/rpc"
	"github.com/willscott/go-nfs-client/nfs/util"
)

var flags struct {
	debug          bool
	uid            int
	gid            int
	machine        string
	privileged     bool
	portmapperPort int
	mountdPort     int
	nfsPort        int
	timeout        time.Duration
	fhInHex        string
	showmount      bool
}

var rootCmd = &cobra.Command{
	Use:   "nfs-cli",
	Short: "Simple NFS cli",
	Example: fmt.Sprintf(`
%[1]s nfs-server.corp.local /exports/homes -u 1000 -g 0
%[1]s -m admin-server.corp.local --showmount nfs-server.corp.local
`, os.Args[0]),
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(args); err != nil {
			util.DefaultLogger.Errorf("error: %s", err)
			os.Exit(1)
		}
	},
}

func init() {
	whoami, err := user.Current()
	if err != nil {
		panic(fmt.Errorf("failed to get current user: %s", err))
	}

	defaultUid, err := strconv.Atoi(whoami.Uid)
	if err != nil {
		panic(fmt.Errorf("failed to parse uid: %s", err))
	}

	defaultGid, err := strconv.Atoi(whoami.Gid)
	if err != nil {
		panic(fmt.Errorf("failed to parse gid: %s", err))
	}
	defaultPrivileged := false

	if defaultUid == 0 {
		defaultPrivileged = true
	}

	rootCmd.Flags().BoolVarP(&flags.debug, "debug", "d", false, "enable debugging")
	rootCmd.Flags().IntVarP(&flags.uid, "uid", "u", defaultUid, "user id")
	rootCmd.Flags().IntVarP(&flags.gid, "gid", "g", defaultGid, "group id")
	rootCmd.Flags().StringVarP(&flags.machine, "machine", "m", "localhost", "machine name")
	rootCmd.Flags().BoolVarP(&flags.privileged, "privileged", "p", defaultPrivileged, "use privileged port (usually requires root)")
	rootCmd.Flags().IntVar(&flags.portmapperPort, "portmapper-port", rpc.PmapPort, "portmapper port")
	rootCmd.Flags().IntVar(&flags.mountdPort, "mountd-port", 0, "mountd port")
	rootCmd.Flags().IntVar(&flags.nfsPort, "nfs-port", 0, "nfs port")
	rootCmd.Flags().DurationVar(&flags.timeout, "timeout", 10*time.Second, "timeout for nfs operations")
	rootCmd.Flags().StringVar(&flags.fhInHex, "fh", "", "specify file handle in binary hex notation (will skip mountd interaction)")
	rootCmd.Flags().BoolVar(&flags.showmount, "showmount", false, "list exported filesystems and exit")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func run(args []string) error {
	var err error

	if flags.debug {
		util.DefaultLogger.SetDebug(true)
	}

	rhost := args[0]

	var mountdPort int = flags.mountdPort
	var nfsPort int = flags.nfsPort

	if flags.fhInHex != "" && mountdPort == 0 {
		mountdPort = 42 // we won't be using mountd, so no need to look it up
	}

	if flags.showmount && nfsPort == 0 {
		nfsPort = 42 // we won't be using nfsd, so no need to look it up
	}

	portmapperAddr := fmt.Sprintf("%s:%d", rhost, flags.portmapperPort)
	mountdPort, nfsPort, err = resolvePorts(portmapperAddr, mountdPort, nfsPort, flags.privileged)
	if err != nil {
		util.DefaultLogger.Errorf("failed to resolve ports using portmapper: %s", err)
		return err
	}
	mountdAddr := fmt.Sprintf("%s:%d", rhost, mountdPort)

	auth := rpc.NewAuthUnix(flags.machine, uint32(flags.uid), uint32(flags.gid)).Auth()

	if flags.showmount {
		return runShowmount(mountdAddr, auth)
	}

	if len(args) < 2 {
		return fmt.Errorf("missing path")
	}

	path := args[1]
	args = args[2:]

	nfsAddr := fmt.Sprintf("%s:%d", rhost, nfsPort)
	return runCli(mountdAddr, nfsAddr, path, auth, args)
}

func runShowmount(mountdAddr string, auth rpc.Auth) error {
	exports, err := showMount(mountdAddr, auth)
	if err != nil {
		return err
	}

	for _, export := range exports {
		fmt.Printf("- %s (%s)\n", export.Directory, strings.Join(export.Options, ", "))
	}
	return nil
}

func runCli(mountdAddr string, nfsAddr string, path string, auth rpc.Auth, args []string) error {
	var err error

	var fh []byte
	if flags.fhInHex != "" {
		if fh, err = hex.DecodeString(flags.fhInHex); err != nil {
			util.DefaultLogger.Errorf("failed to unhex fh param: %s", err)
			return err
		}
	} else {
		if fh, err = getFileHandle(mountdAddr, path, auth); err != nil {
			util.DefaultLogger.Errorf("failed to obtain file handle for '%s': %s", path, err)
			return err
		}
		util.DefaultLogger.Debugf("obtained file handle for '%s': %s", path, hex.EncodeToString(fh))
	}

	nfsClient, err := rpc.DialTCP("tcp", nfsAddr, flags.privileged)
	if err != nil {
		panic(err)
	}
	target, err := nfs.NewTargetWithClient(nfsClient, auth, fh, path, flags.timeout)
	if err != nil {
		panic(err)
	}
	defer target.Close()

	return cli.Main(target, path, args)
}
