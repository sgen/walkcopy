package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

type Data struct {
	Name string
}

func copyTemplate(from, to string, data Data, perm os.FileMode) error {
	to = to[:len(to)-5]
	fmt.Printf("copying template  (%s) to (%s)\n", from, to)
	tmpl, err := template.ParseFiles(from)
	if err != nil {
		return err
	}
	t, err := os.OpenFile(to, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer t.Close()
	if err := tmpl.Execute(t, data); err != nil {
		return err
	}
	return t.Close()
}

func copyFile(from, to string, perm os.FileMode) error {
	fmt.Printf("copying file      (%s) to (%s)\n", from, to)
	f, err := os.Open(from)
	if err != nil {
		return err
	}
	defer f.Close()
	t, err := os.OpenFile(to, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer t.Close()

	if _, err := io.Copy(t, f); err != nil {
		return err
	}
	return t.Close()
}

func copyDir(from, to string, perm os.FileMode) error {
	fmt.Printf("copying directory (%s) to (%s)\n", from, to)
	return os.MkdirAll(to, perm)
}

func walkFunc(root, to string, data Data) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relToRoot := path[len(root):]
		specTo := filepath.Join(to, relToRoot)
		if info.IsDir() {
			return copyDir(path, specTo, info.Mode().Perm())
		} else if filepath.Ext(path) == ".tmpl" {
			return copyTemplate(path, specTo, data, info.Mode().Perm())
		}
		return copyFile(path, specTo, info.Mode().Perm())
	}
}

func handle(err error, msg string) {
	if err == nil {
		return
	}
	if msg == "" {
		msg = "error"
	}
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err)
	os.Exit(1)
}

func main() {
	data := Data{
		Name: "World",
	}
	from, err := filepath.Abs("./root/from")
	handle(err, "could not find absolute path for from")

	to, err := filepath.Abs("./root/to")
	handle(err, "could not find absolute path for to")

	err = filepath.Walk(from, walkFunc(from, to, data))
	handle(err, "could not copy")
}
