package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/leonelquinteros/gotext"
)

var (
	//go:embed locale/*.po
	poFs embed.FS
)

type Lang struct {
	lang string
	Po   *gotext.Po
}

func (l Lang) T(str string, vars ...interface{}) string {
	return l.Po.Get(str, vars...)
}

func NewLang() Lang {
	lang := os.Getenv("LANG")
	if lc := os.Getenv("LANGUAGE"); lc != "" {
		lang = lc
	} else if lc := os.Getenv("LC_ALL"); lc != "" {
		lang = lc
	} else if lc := os.Getenv("LC_MESSAGES"); lc != "" {
		lang = lc
	}
	lang = lang[:2]

	ret := Lang{
		lang: lang,
		Po:   gotext.NewPoFS(poFs),
	}
	ret.Po.ParseFile("locale/" + lang + ".po")

	if os.Getenv("LOG") == "1" {
		xs, _ := getAllFilenames(&poFs)
		for _, x := range xs {
			fmt.Println(" -> ", x)
		}
		fmt.Println("lang  -> ", lang)
		fmt.Printf("domain  -> %v \n\n", ret.Po.GetDomain())
	}

	return ret
}

func getAllFilenames(efs *embed.FS) (files []string, err error) {
	if err := fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		files = append(files, path)

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}
