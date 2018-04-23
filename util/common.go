package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"

	"os/user"

)


func CheckErr(err error) bool {
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func CheckAndExit(err error) {
	if !CheckErr(err) {
		os.Exit(1)
	}
}

func MustExec(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(b))
		os.Exit(1)
	}
}

func MustExecRtOut(name string, arg ...string) string {
	cmd := exec.Command(name, arg...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(b))
		os.Exit(1)
	}
	return string(b)
}

func MustExecNoOut(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	CheckAndExit(cmd.Run())
}

func TryExec(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CheckRoot() {
	u, err := user.Current()
	CheckAndExit(err)

	if u.Uid != "0" || u.Gid != "0" {
		fmt.Println("This command must be run as root! (sudo)")
		os.Exit(1)
	}
}

func OSEditInput() string {

	f, err := ioutil.TempFile("", "gitflow-toolkit")
	CheckAndExit(err)
	defer os.Remove(f.Name())

	// write utf8 bom
	bom := []byte{0xef, 0xbb, 0xbf}
	_, err = f.Write(bom)
	CheckAndExit(err)

	// use system edit
	editor := "vim"
	if runtime.GOOS == "windows" {
		editor = "notepad"
	}
	if v := os.Getenv("VISUAL"); v != "" {
		editor = v
	} else if e := os.Getenv("EDITOR"); e != "" {
		editor = e
	}

	// edit tmp file
	cmd := exec.Command(editor, f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	CheckAndExit(cmd.Run())
	raw, err := ioutil.ReadFile(f.Name())
	CheckAndExit(err)
	input := string(bytes.TrimPrefix(raw, bom))

	return input
}
