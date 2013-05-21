package main

import (
	"code.google.com/p/go9p/p"
	"code.google.com/p/go9p/p/srv"
	"github.com/conformal/btcwire"
	"os"
)

type BlockFile struct {
	srv.File
	MerkleRoot *btcwire.ShaHash
}

func (bf *BlockFile) Read(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	str := bf.MerkleRoot.String()
	b := []byte(str)

	n := len(b)

	if offset >= uint64(n) {
		return 0, nil
	}

	b = b[int(offset):n]
	n -= int(offset)
	if len(buf) < n {
		n = len(buf)
	}

	copy(buf[offset:int(offset)+n], b[offset:])
	return n, nil
}

func (bf *BlockFile) Wstat(fid *srv.FFid, dir *p.Dir) error {
	return nil
}

func (bf *BlockFile) Remove(fid *srv.FFid) error {
	return nil
}

func SrvHeaders(headers chan *btcwire.BlockHeader) error {
	var err error
	var s *srv.Fsrv

	user := p.OsUsers.Uid2User(os.Geteuid())
	root := new(srv.File)

	err = root.Add(nil, "/", user, nil, p.DMDIR|0777, nil)
	if err != nil {
		goto error
	}

	go func() {
		for hdr := range headers {
			sha, _ := hdr.BlockSha(btcwire.ProtocolVersion)
			blkf := &BlockFile{MerkleRoot: &hdr.MerkleRoot}
			_ = blkf.Add(root, sha.String(), p.OsUsers.Uid2User(os.Geteuid()), nil, 0444, blkf)
		}
	}()

	s = srv.NewFileSrv(root)
	s.Dotu = true
	//s.Debuglevel = 1
	s.Start(s)
	err = s.StartNetListener("tcp", ":5640")
	if err != nil {
		goto error
	}

	return nil
error:
	return err
}
