package file

import (
	"os/user"
	"testing"
)

func TestEnsureDir(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		t.Error("error, CurrentUser")
	}
	uname := user.Name
	if uname == "root" {
		return
	}

	rootDir := "/root/test_ensure_dir/"
	tmpDir := "/tmp/test_ensure_dir/abc"
	err1 := EnsureDir(rootDir)
	err2 := EnsureDir(tmpDir)
	if !(err1 != nil && err2 == nil) {
		t.Error("error, EnsureDir")
	}
}

func TestEnsureDirRW(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		t.Error("error, CurrentUser")
	}
	uname := user.Name
	if uname == "root" {
		return
	}

	tmpDir := "/tmp/test_ensure_dir/abc"
	err1 := EnsureDirRW(tmpDir)
	if !(err1 == nil) {
		t.Error("error, EnsureDirRW", err1)
	}
}
