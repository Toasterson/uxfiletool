package uxfiletool

import (
	"os"
	"strings"
	"syscall"
	"io"
	"time"
	"path/filepath"
	"fmt"
	"github.com/toasterson/glog"
)

type copyWorker struct {
	src string
	dest string
	list []string

}

func ExactCopy(srcFile, destDir string) error {
	destFile := filepath.Join(destDir, srcFile)
	srcMode, err := os.Lstat(srcFile)
	if err != nil {
		return err
	}
	if srcMode.IsDir() {
		return fmt.Errorf("ignoring directory %s", srcFile)
	}
	//Ignore Cases where we have a file directly in /bin being a link to /usr/bin
	if strings.Count(srcFile, string(os.PathSeparator)) == 1{
		glog.Tracef("%s is Directly in / not making directories", srcFile)
	} else {
		if err := mirrorDirPath(srcFile, destDir); err != nil {
			return err
		}
	}
	srcStat := srcMode.Sys().(*syscall.Stat_t)
	if srcMode.Mode()&os.ModeSymlink != 0 {
		//We have a Symlink thus Create it on the Target
		dstTarget, _ := os.Readlink(srcFile)
		if err := os.Symlink(dstTarget, destFile); err != nil{
			return err
		}
		return nil
	}
	return copyFileExact(srcFile, destFile, srcStat)
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
		srcStat := info.Sys().(*syscall.Stat_t)
		return copyFileExact(path, dstpath, srcStat)
	}
	return nil
}

func copyTimes(target string, srcStat *syscall.Stat_t) error {
	return os.Chtimes(target, time.Unix(int64(srcStat.Atim.Sec),int64(srcStat.Atim.Nsec)), time.Unix(int64(srcStat.Mtim.Sec),int64(srcStat.Mtim.Nsec)))
}

func mirrorDirPath(srcFile, destDir string) error {
	pathList := strings.Split(filepath.Dir(srcFile), string(os.PathSeparator))
	path := "/"
	if strings.HasPrefix(srcFile, "./"){
		path = "./"
	}
	for _, p := range pathList {
		path = filepath.Join(path, p)
		dirMode, err := os.Lstat(path)
		if err != nil {
			return err
		}
		dirStat := dirMode.Sys().(*syscall.Stat_t)
		destPath := filepath.Join(destDir, path)
		if err := os.MkdirAll(destPath, dirMode.Mode()); err != nil {
			if os.IsExist(err) {
				continue
			}
			return err
		}
		if err := os.Chown(destPath, int(dirStat.Uid), int(dirStat.Gid)); err != nil {
			return err
		}
	}
	return nil
}

func copyFileExact(source, dest string, srcStat *syscall.Stat_t) error {
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
	if _, err = io.Copy(dst, src); err != nil{
		return err
	}

	if err := syscall.Chmod(dest, srcStat.Mode); err != nil{
		return err
	}
	if err := syscall.Chown(dest, int(srcStat.Uid), int(srcStat.Gid)); err != nil{
		return err
	}
	return nil
}
