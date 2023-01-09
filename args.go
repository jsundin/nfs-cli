package main

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/go-nfs/nfsv3/nfs/rpc"
	"github.com/go-nfs/nfsv3/nfs/util"
	"github.com/spf13/cobra"
)

func launch_client() {
	var machine string
	var uidgid string
	var debug bool
	var priv bool

	rootCmd := &cobra.Command{
		Use:   "nfs-cli <rhost> <target> [command]",
		Short: "Simple NFS cli",
		Args:  cobra.MinimumNArgs(2),

		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				util.DefaultLogger.SetDebug(true)
			}

			var uid, gid int
			if uidgid == "" {
				whoami, err := user.Current()
				if err != nil {
					panic(err)
				}
				if uid, err = strconv.Atoi(whoami.Uid); err != nil {
					panic(err)
				}
				if gid, err = strconv.Atoi(whoami.Gid); err != nil {
					panic(err)
				}
				util.Debugf("using uid=%d,gid=%d from current user %s", uid, gid, whoami.Username)
			} else {
				var err error
				ug := strings.Split(uidgid, ":")
				if len(ug) == 1 || len(ug) == 2 {
					if uid, err = strconv.Atoi(ug[0]); err != nil {
						panic(err)
					}
					if len(ug) == 2 {
						if gid, err = strconv.Atoi(ug[1]); err != nil {
							panic(err)
						}
					} else {
						gid = uid
					}
				} else {
					panic("bad uidgid format")
				}
			}

			auth_unix := rpc.NewAuthUnix(machine, uint32(uid), uint32(gid))
			client(auth_unix.Auth(), args[0], args[1], priv, args[2:])
		},
	}

	rootCmd.Flags().StringVarP(&machine, "machine", "m", "localhost", "machine identifier")
	rootCmd.Flags().StringVarP(&uidgid, "user", "u", "", "user, format: uid[:gid] (default current user)")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "enable go-nfs debugging")
	rootCmd.Flags().BoolVarP(&priv, "priv", "p", false, "use a privileged port, usually requires root")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
}
