package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	_ "github.com/mattn/getwild"
	"github.com/mattn/go-tty"

	"github.com/itchyny/mmv"
)

const name = "mmv"

const version = "0.1.0"

var revision = "HEAD"

func main() {
	os.Exit(run(os.Args[1:]))
}

const (
	exitCodeOK = iota
	exitCodeErr
)

func run(args []string) int {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fs.SetOutput(os.Stdout)
		fmt.Printf(`%[1]s - rename multiple files with editor

Version: %s (rev: %s/%s)

Synopsis:
  %% %[1]s file ...

Options:
`, name, version, revision, runtime.Version())
		fs.PrintDefaults()
	}
	var showVersion bool
	var dryRun bool
	fs.BoolVar(&showVersion, "version", false, "print version")
	fs.BoolVar(&dryRun, "dry-run", false, "only show the operation that would have been performed")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return exitCodeOK
		}
		return exitCodeErr
	}
	if showVersion {
		fmt.Printf("%s %s (rev: %s/%s)\n", name, version, revision, runtime.Version())
		return exitCodeOK
	}
	args = fs.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: %s file ...\n", name)
		return exitCodeErr
	}
	if err := rename(args, dryRun); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", name, err)
		return exitCodeErr
	}
	return exitCodeOK
}

func rename(args []string, dryRun bool) error {
	xs := make(map[string]bool, len(args))
	for _, src := range args {
		if xs[src] {
			return fmt.Errorf("duplicate source: %s", src)
		}
		xs[src] = true
	}
	f, err := ioutil.TempFile("", name+"-")
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	for _, arg := range args {
		f.WriteString(arg)
		f.WriteString("\n")
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	tty, err := tty.Open()
	if err != nil {
		return err
	}
	defer tty.Close()
	cmd := exec.Command(editor, f.Name())
	cmd.Stdin = tty.Input()
	cmd.Stdout = tty.Output()
	cmd.Stderr = tty.Output()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("abort renames: %s", err)
	}
	if err := f.Close(); err != nil {
		return err
	}
	cnt, err := ioutil.ReadFile(f.Name())
	if err != nil {
		return err
	}
	got := strings.Split(strings.TrimRight(string(cnt), "\n"), "\n")
	if len(args) != len(got) {
		return errors.New("do not delete or add lines")
	}
	files := make(map[string]string, len(args))
	for i, src := range args {
		files[src] = got[i]
	}
	return mmv.Rename(files, dryRun)
}
