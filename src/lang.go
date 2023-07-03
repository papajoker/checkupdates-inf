package main

import (
	"embed"
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

	return ret
}
