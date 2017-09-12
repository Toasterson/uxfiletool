package uxfiletool

import (
	"testing"
	"os"
)

func TestExactCopySymlink(t *testing.T){
	//Make a Symlink that is in a faked /
	//home := os.ExpandEnv("$HOME")
	os.Symlink("/tmp", "./test")
	os.Mkdir("target", 0755)
	if err := ExactCopy("./test", "target"); err != nil{
		t.Errorf("Cannot Copy symlink because %s", err)
		t.Fail()
	}
	os.Remove("./test")
	os.Remove("./target/test")
	os.Remove("./target")
}

func TestExactCopyDirectory(t *testing.T) {
	home := os.ExpandEnv("$HOME")
	os.Mkdir("./target", 0755)
	if err := ExactCopy(home+"/Desktop", "./target"); err == nil {
		t.Error("This should fail")
	}
	os.RemoveAll("./target")
}

func TestExactCopyFile(t *testing.T) {
	os.Mkdir("./target", 0755)
	os.MkdirAll("./source/to/file", 0755)
	f, _ := os.Create("./source/to/file/file.txt")
	f.Close()
	if err := ExactCopy("./source/to/file/file.txt", "./target"); err != nil {
		t.Errorf("Can not copy sample file: %s", err)
	}
	os.RemoveAll("./target")
	os.RemoveAll("./source")
}