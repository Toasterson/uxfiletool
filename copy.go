package uxfiletool

import (
	"os"
	"strings"
	"syscall"
	"io"
	"time"
	"path/filepath"
)

type copyWorker struct {
	src string
	dest string
	list []string

}

func ExactCopy(src, dest string) error {
	srcMode, err := os.Stat(src)
	if err != nil {
		return err
	}
	return copyFileExact(src, srcMode, dest)
}

func ExactCopyPath(src, dest string, entries []string) error {
	wk := copyWorker{src:src, dest: dest, list: entries}
	if len(entries) > 0 {
		for _, entry := range entries {
			path := filepath.Join(src, entry)
			if err := filepath.Walk(path, wk.walkCopy); err != nil {
				return err
			}
		}
		return nil
	} else {
		return filepath.Walk(src, wk.walkCopy)
	}
}

func (wk copyWorker)walkCopy(path string, info os.FileInfo, err error) error {
	dstpath := strings.Replace(path, wk.src, wk.dest, -1)
	lsrcinfo, err := os.Lstat(path)
	if os.IsNotExist(err) {
		//Ignore nonexistent Directories
		return nil
	}
	if err != nil{
		return err
	}
	if info.IsDir() {
		if err := os.Mkdir(dstpath, info.Mode()); err != nil{
			return err
		}
		srcStat := info.Sys().(*syscall.Stat_t)
		if err := syscall.Chmod(dstpath, srcStat.Mode); err != nil{
			return err
		}
		if err := syscall.Chown(dstpath, int(srcStat.Uid), int(srcStat.Gid)); err != nil{
			return err
		}
	} else if lsrcinfo.Mode()&os.ModeSymlink != 0 {
		//We have a Symlink thus Create it on the Target
		dstTarget, _ := os.Readlink(path)
		if err := os.Symlink(dstTarget, dstpath); err != nil{
			return err
		}
	} else {
		//We Have a regular File Copy it
		return copyFileExact(path, info, dstpath)
	}
	return nil
}

func copyFileExact(source string, srcInfo os.FileInfo, dest string) error {
	src, err := os.Open(source)
	defer src.Close()
	if err != nil{
		return err
	}
	dst, err := os.Create(dest)
	defer dst.Close()
	if err != nil{
		return err
	}
	_, err = io.Copy(dst, src)
	if err != nil{
		return err
	}
	srcStat := srcInfo.Sys().(*syscall.Stat_t)
	if err := syscall.Chmod(dest, srcStat.Mode); err != nil{
		return err
	}
	if syscall.Chown(dest, int(srcStat.Uid), int(srcStat.Gid)); err != nil{
		return err
	}
	return nil
}

func copyTimes(target string, srcStat *syscall.Stat_t) error {
	return os.Chtimes(target, time.Unix(int64(srcStat.Atim.Sec),int64(srcStat.Atim.Nsec)), time.Unix(int64(srcStat.Mtim.Sec),int64(srcStat.Mtim.Nsec)))
}
