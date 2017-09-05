package uxfiletool

import (
	"os"
	"path/filepath"
	"strings"
	"github.com/toasterson/opencloud/common"
)

type FindType int

const (
	FindTypeFile = iota
	FindTypeDir = iota
	FindTypeLink = iota
)

var (
	isaexec_sufixes = []string{"amd64", "i86"}
	lib_paths = []string{"/usr/lib", "/lib"}
)

func FindByGlob(pattern string) (files []string){
	paths := lib_paths
	for _, path_var := range []string{"$LD_LIBRARY_PATH", "$PATH"}{
		path_var = os.ExpandEnv(path_var)
		paths = append(paths, filepath.SplitList(path_var)...)
	}
	for _, vari := range paths{
		path := filepath.Join(vari, pattern)
		found, err := filepath.Glob(path)
		if err == nil && found != nil {
			for _, f := range found {
				info, err := os.Stat(f)
				if err != nil {
					continue
				}
				if info.Mode().IsDir() {
					files = append(files, FindAllIn(f, FindTypeFile)...)
					files = append(files, FindAllIn(f, FindTypeLink)...)
				} else {
					files = append(files, f)
				}
			}
		}
	}
	return
}

func FindAllIn(dir string, findType FindType) (files []string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		switch findType {
		case FindTypeDir:
			if info.Mode().IsDir(){
				files = append(files, path)
			}
		case FindTypeFile:
			if info.Mode().IsRegular(){
				files = append(files, path)
			}
		case FindTypeLink:
			if info.Mode()&os.ModeSymlink != 0 {
				files = append(files, path)
			}
		}
		return nil
	})
	return
}

func FindLib(lib string) (libs []string) {
	paths := lib_paths
	path_var := os.ExpandEnv("$LD_LIBRARY_PATH")
	if strings.Contains(path_var, string(os.PathListSeparator)){
		paths = append(paths, filepath.SplitList(path_var)...)
	}
	for _, libp := range paths {
		path := filepath.Join(libp, lib)
		if _, err := os.Stat(path); err == nil {
			libs = append(libs, path)
		}
	}
	return
}

func FindInPath(bin string) (bins []string) {
	var paths = []string{}
	path_var := os.ExpandEnv("$PATH")
	if strings.Contains(path_var, string(os.PathListSeparator)){
		paths = filepath.SplitList(path_var)
	}
	for _, binp := range paths{
		path := filepath.Join(binp, bin)
		if common.FileExists(path){
			bins = append(bins, path)
		}
		for _, arch_part := range isaexec_sufixes{
			path := filepath.Join(binp, arch_part, bin)
			if common.FileExists(path){
				bins = append(bins, path)
			}
		}
	}
	return
}
