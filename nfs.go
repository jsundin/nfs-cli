package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/willscott/go-nfs-client/nfs"
	"github.com/willscott/go-nfs-client/nfs/rpc"
	"github.com/willscott/go-nfs-client/nfs/util"
	"github.com/willscott/go-nfs-client/nfs/xdr"
)

func newMountdClient(mountdAddr string) (*nfs.Mount, error) {
	mountdClient, err := rpc.DialTCP("tcp", mountdAddr, false)
	if err != nil {
		return nil, err
	}
	mountd := &nfs.Mount{
		Client: mountdClient,
	}
	return mountd, nil
}

func getFileHandle(mountdAddr string, path string, auth rpc.Auth) ([]byte, error) {
	mountd, err := newMountdClient(mountdAddr)
	if err != nil {
		return nil, err
	}
	defer mountd.Close()

	return internalGetFileHandle(mountd, path, auth)
}

func internalGetFileHandle(m *nfs.Mount, path string, auth rpc.Auth) ([]byte, error) {
	type mount struct {
		rpc.Header
		Dirpath string
	}
	res, err := m.Call(&mount{
		rpc.Header{
			Rpcvers: 2,
			Prog:    nfs.MountProg,
			Vers:    nfs.MountVers,
			Proc:    nfs.MountProc3MNT,
			Cred:    auth,
			Verf:    rpc.AuthNull,
		},
		path,
	})
	if err != nil {
		return nil, err
	}
	mountstat3, err := xdr.ReadUint32(res)
	if err != nil {
		return nil, err
	}

	switch mountstat3 {
	case nfs.MNT3Ok:
		fh, err := xdr.ReadOpaque(res)
		if err != nil {
			return nil, err
		}
		_, _ = xdr.ReadUint32List(res)
		return fh, nil

	case nfs.MNT3ErrPerm:
		return nil, errors.New("MNT3ERR_PERM")
	case nfs.MNT3ErrNoEnt:
		return nil, errors.New("MNT3ERR_NOENT")
	case nfs.MNT3ErrIO:
		return nil, errors.New("MNT3ERR_IO")
	case nfs.MNT3ErrAcces:
		return nil, errors.New("MNT3ERR_ACCES")
	case nfs.MNT3ErrNotDir:
		return nil, errors.New("MNT3ERR_NOTDIR")
	case nfs.MNT3ErrNameTooLong:
		return nil, errors.New("MNT3ERR_NAMETOOLONG")
	}
	return nil, fmt.Errorf("unknown mount stat: %d", mountstat3)
}

func resolvePorts(portmapperAddr string, mountdPort int, nfsPort int, privileged bool) (int, int, error) {
	unresolvedPortsQueries := []*rpc.Mapping{}
	if mountdPort == 0 {
		unresolvedPortsQueries = append(unresolvedPortsQueries, &rpc.Mapping{Prog: nfs.MountProg, Vers: nfs.MountVers, Prot: rpc.IPProtoTCP, Port: 0})
	}
	if nfsPort == 0 {
		unresolvedPortsQueries = append(unresolvedPortsQueries, &rpc.Mapping{Prog: nfs.Nfs3Prog, Vers: nfs.Nfs3Vers, Prot: rpc.IPProtoTCP, Port: 0})
	}
	if len(unresolvedPortsQueries) > 0 {
		if err := resolvePortsUsingPortmapper(portmapperAddr, privileged, unresolvedPortsQueries...); err != nil {
			return 0, 0, err
		}
		for _, res := range unresolvedPortsQueries {
			switch res.Prog {
			case nfs.MountProg:
				mountdPort = int(res.Port)
				util.DefaultLogger.Debugf("resolved mountd port using portmapper: %d", mountdPort)

			case nfs.Nfs3Prog:
				nfsPort = int(res.Port)
				util.DefaultLogger.Debugf("resolved nfs port using portmapper: %d", nfsPort)

			default:
				return 0, 0, errors.New("unknown portmapper response")
			}
		}
	}
	return mountdPort, nfsPort, nil
}

func resolvePortsUsingPortmapper(addr string, privileged bool, mappings ...*rpc.Mapping) error {
	client, err := rpc.DialTCP("tcp", addr, privileged)
	if err != nil {
		return err
	}
	pm := &rpc.Portmapper{Client: client}
	defer pm.Close()
	for _, mapping := range mappings {
		port, err := pm.Getport(*mapping)
		if err != nil {
			return err
		}
		mapping.Port = uint32(port)
	}
	return nil
}

func readPaddedOpaque(r io.Reader) ([]byte, error) {
	v, err := xdr.ReadOpaque(r)
	if err != nil {
		return v, err
	}
	pad := (4 - (len(v) % 4)) % 4
	_, err = r.Read(make([]byte, pad))
	return v, err
}

type ExportedFilesystem struct {
	Directory string
	Options   []string
}

func showMount(mountdAddr string, auth rpc.Auth) ([]*ExportedFilesystem, error) {
	mountd, err := newMountdClient(mountdAddr)
	if err != nil {
		return nil, err
	}
	defer mountd.Close()

	type ExportArgs struct {
		rpc.Header
	}

	res, err := mountd.Call(&ExportArgs{
		rpc.Header{
			Rpcvers: 2,
			Prog:    nfs.MountProg,
			Vers:    nfs.MountVers,
			Proc:    nfs.MountProc3Export,
			Cred:    auth,
			Verf:    rpc.AuthNull,
		},
	})
	if err != nil {
		return nil, err
	}

	exports := []*ExportedFilesystem{}
	for {
		export := ExportedFilesystem{}

		exportListEntryFollows, err := xdr.ReadUint32(res)
		if err != nil {
			return nil, err
		}
		if exportListEntryFollows == 0 {
			util.DefaultLogger.Debugf("no more export list entries")
			break
		}

		directory, err := readPaddedOpaque(res)
		if err != nil {
			return nil, err
		}
		util.DefaultLogger.Debugf("got directory '%s'", string(directory))
		export.Directory = string(directory)

		for {
			groupValueFollows, err := xdr.ReadUint32(res)
			if err != nil {
				return nil, err
			}
			if groupValueFollows == 0 {
				util.DefaultLogger.Debugf("no more group values")
				break
			}

			groupValue, err := readPaddedOpaque(res)
			if err != nil {
				return nil, err
			}
			util.DefaultLogger.Debugf("got group value '%s'", string(groupValue))

			export.Options = append(export.Options, string(groupValue))
		}

		exports = append(exports, &export)
	}

	return exports, err
}
