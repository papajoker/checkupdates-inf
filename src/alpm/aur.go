package alpm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const url = "https://aur.archlinux.org/rpc/?v=5&type=info&by=name&arg="

type responce struct {
	Resultcount int `json:"resultcount"`
	Results     []struct {
		Description  string      `json:"Description"`
		LastModified int         `json:"LastModified"`
		Name         string      `json:"Name"`
		OutOfDate    interface{} `json:"OutOfDate"`
		PackageBase  string      `json:"PackageBase"`
		URL          string      `json:"URL"`
		URLPath      string      `json:"URLPath"`
		Version      string      `json:"Version"`
	} `json:"results"`
	Type    string `json:"type"`
	Version int    `json:"version"`
}

func AurRequestExists(pkgname string) (rurl string) {
	for _, e := range [3]string{"", "-bin", "-git"} {
		rurl = aurRequest(pkgname + e)
		if strings.HasPrefix(rurl, "http") {
			return rurl
		}
	}
	return rurl
}

func aurRequest(pkgname string) string {
	no := "NO"

	resp, err := http.Get(url + pkgname)
	if err != nil {
		fmt.Println(err)
		return no
	}
	defer resp.Body.Close()

	r := responce{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		fmt.Println(err)
		return no
	}

	if r.Resultcount > 0 {
		return "https://aur.archlinux.org/packages/" + pkgname
	}

	return no
}
