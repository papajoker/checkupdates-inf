package alpm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/leonelquinteros/gotext"
)

type (
	tdesc map[string][]string

	Package struct {
		NAME     string
		VERSION  string
		DESC     string
		REPO     string
		URL      string
		REPLACES []string
		PROVIDES []string

		IsDep      bool
		ReplacedBy string
	}

	Packages map[string]*Package
)

// parse desc file content
func (p *Package) set(desc string) bool {
	tmpdesc := strings.Split(desc, "\n\n")
	adesc := make(tdesc)
	for i := range tmpdesc {
		tmp := strings.Split(tmpdesc[i], "\n")
		idx := strings.Replace(tmp[0], "%", "", -1)
		if len(tmp) > 1 {
			adesc[idx] = tmp[1:]
		} else {
			adesc[idx] = make([]string, 0)
		}
	}

	p.VERSION = getFieldString(adesc, "VERSION")
	p.NAME = getFieldString(adesc, "NAME")
	p.DESC = getFieldString(adesc, "DESC")
	p.URL = getFieldString(adesc, "URL")
	p.REPLACES = getFieldArray(adesc, "REPLACES")
	p.PROVIDES = getFieldArray(adesc, "PROVIDES")
	p.IsDep = getFieldString(adesc, "REASON") == "1"
	return true
}

func (p Package) Desc(maxi int) string {
	if maxi > len(p.DESC) {
		return p.DESC
	}
	return p.DESC[:maxi-1] + "…"

}

func getFieldString(adesc tdesc, key string) string {
	if len(adesc[key]) < 1 {
		return ""
	}
	return strings.TrimSpace(adesc[key][0])
}

func getFieldArray(adesc tdesc, key string) []string {
	if len(adesc[key]) < 1 {
		return make([]string, 0)
	}
	for k, v := range adesc[key] {
		adesc[key][k] = strings.TrimSpace(strings.SplitN(v, ":", 2)[0])
	}
	return adesc[key][0:]
}

func ExtractTarGz(gzipStream io.Reader, pkgs Packages, repo string) Packages {

	errMsg := gotext.Get("Error")

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		fmt.Println(errMsg, err.Error())
		return pkgs
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(errMsg, err.Error())
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			/*fmt.Println("::dir:",header.Name)*/
		case tar.TypeReg:

			buf := new(bytes.Buffer)

			if _, err := buf.ReadFrom(tarReader); err != nil {
				if err != io.EOF {
					//nb = nb + 1
					fmt.Println(errMsg, err.Error())
					log.Fatalf("ExtractTarGz:  failed: %s", err.Error())
				}
			}

			pkg := Package{REPO: repo}
			if pkg.set(buf.String()) {
				if _, ok := pkgs["pkg.NAME"]; ok {
					fmt.Printf("\t # %s : %s\n", gotext.Get("ignore duplicate"), pkg.NAME)
				} else {
					pkgs[pkg.NAME] = &pkg
				}
			}
		default:
			fmt.Println(errMsg, "def", header.Typeflag, header.Name)
			fmt.Printf("ExtractTarGz: uknown type: %v in %s\n", header.Typeflag, header.Name)
			os.Exit(8)
		}
	}
	return pkgs
}

func Load(dir string, repos []string) (pkgs Packages) {
	pkgs = make(Packages, 5000)
	for _, repo := range repos {
		nb := len(pkgs)
		//fmt.Printf("%v# %s ...%v\t", theme.ColorGray, repo, theme.ColorNone)
		f, err := os.Open(dir + "/" + repo + ".db")
		if err != nil {
			fmt.Printf("Error: can't read file %s\n", dir+"/"+repo+".db")
			//os.Exit(1)
			continue
		}
		defer f.Close()
		pkgs = ExtractTarGz(f, pkgs, repo)
		if len(pkgs)-nb == 0 {
			sync := "sync"
			if strings.Contains(dir, "/var/lib/") {
				sync = "local"
			}
			fmt.Printf(
				"%s: '%s' %s, %s\n",
				gotext.Get("warning"),
				repo, sync,
				gotext.Get("repo empty ? or all packages are ignored"),
			)
		}
		//fmt.Println(repo, len(pkgs)-nb, "packages")
	}
	return pkgs
}

func Replaced(pkgName string, pkgs Packages) (pkg *Package, err bool) {
	for _, pkg := range pkgs {
		for _, alias := range pkg.REPLACES {
			if pkgName == alias {
				return pkg, true
			}
		}
	}
	return nil, false
}

func ProvideBy(pkgName string, pkgs Packages) (provides []string, ok bool) {
	for _, pkg := range pkgs {
		for _, alias := range pkg.PROVIDES {
			if strings.Contains(alias, ".so") {
				continue
			}
			alias = func(s string) string {
				if matchs := strings.SplitN(s, "=", 2); len(matchs) > 1 {
					s = matchs[0]
				}
				return s
			}(alias)
			if pkgName == alias {
				provides = append(provides, pkg.NAME)
			}
		}
	}
	return provides, len(provides) > 0
}

func LocalParse(directory string, searchs []*Package) (results []*Package, ok bool) {

	errMsg := gotext.Get("Error")

	matchs, err := filepath.Glob(directory + "/*/desc")
	if err != nil {
		fmt.Printf("%s ! %v", errMsg, err)
	}
	if len(matchs) < 1 {
		fmt.Printf("%s ! %s empty ?", errMsg, directory)
	}
	for _, f := range matchs {
		for _, n := range searchs {
			if strings.Contains(f, "/"+n.NAME+"-") {
				pkg := Package{REPO: "local"}
				b, _ := os.ReadFile(f)
				pkg.set(string(b))
				if pkg.NAME == n.NAME {
					//fmt.Printf("  %v\n", pkg)
					pkg.ReplacedBy = n.ReplacedBy
					results = append(results, &pkg)
				}
			}
		}
	}
	return results, len(results) > 0
}
