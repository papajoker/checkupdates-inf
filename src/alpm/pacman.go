package alpm

import (
	"bufio"
	"checkupdates-inf/theme"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"github.com/leonelquinteros/gotext"
)

type CheckupdatesOutput struct {
	Name         string
	VersionLocal string
	Version      string
}

func (c CheckupdatesOutput) Sizes(mini int) (sizes [3]int) {
	sizes[0] = len(c.Name)
	sizes[1] = len(c.Version + theme.ColorGreen + theme.ColorNone)
	sizes[2] = len(c.VersionLocal + theme.ColorGreen + theme.ColorNone)
	for k := range sizes {
		if sizes[k] < mini {
			sizes[k] = mini
		}
	}
	return sizes
}

func realVersion(version string) (string, string) {
	if matchs := strings.SplitN(version, ":", 2); len(matchs) > 1 {
		return matchs[1], "(" + matchs[0] + ":)"
	}
	return version, ""
}

func VersionColor(installed, next_version string, color string) (v string, epoch string) {

	installed, epoch = realVersion(installed)
	if matchs := strings.SplitN(next_version, ":", 2); len(matchs) > 1 {
		if epoch == "" {
			epoch = "(" + matchs[0] + ":)"
		}
		next_version = matchs[1]
	}

	for i := range installed {
		if i >= len(next_version) {
			break
		}
		if installed[i] != next_version[i] {
			next_version = next_version[0:i] + color + next_version[i:]
			break
		}
	}
	v = next_version + theme.ColorNone
	return v, epoch
}

func ListRepos(confName string) (repos []string) {
	file, err := os.Open(confName)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") {
			if strings.HasPrefix(line, "[options]") {
				continue
			}
			line := strings.Replace(line[1:], "]", "", -1)
			repos = append(repos, line)
		}
	}
	return repos
}

func runCmd(cmdstr string, tee bool) (versions []CheckupdatesOutput, err error) {

	args := strings.Fields(cmdstr)
	cmd := exec.Command(args[0], args[1:]...)

	var stdout io.ReadCloser
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return versions, err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	scanner := bufio.NewScanner(stdout)
	go func() {
		capture := true
		for scanner.Scan() {

			line := scanner.Text()
			//if tee && !strings.HasPrefix(line, ":") {
			if tee {
				// next block char/char is best ???
				fmt.Println(scanner.Text())
			}

			if capture {
				args = strings.Fields(line)
				if len(args) == 4 && args[2] == "->" {
					versions = append(versions, CheckupdatesOutput{args[0], args[1], args[3]})
				} else {
					capture = false
					//break //BUG etéit pour tester block suivant ...
					// continue for "tee"
				}
			}
		}
		if tee {
			//fmt.Println("\ndownload state ...")
			//var c rune
			c := make([]byte, 1)
			//reader := bufio.NewReader(stdout)
			for err == nil {
				_, err = stdout.Read(c)
				//c, _, err = reader.ReadRune()
				fmt.Printf("%s", string(c))
			}
		}
		wg.Done()
	}()

	if err = cmd.Start(); err != nil {
		return versions, err
	}

	wg.Wait()

	return versions, cmd.Wait()
}

func Checkupdates(download bool) (versions []CheckupdatesOutput) {
	cmd := "checkupdates"
	if download {
		cmd += " -d"
	}
	versions, err := runCmd(cmd, download)
	if err != nil {
		if err.Error() == "exit status 2" {
			return versions
		}
		fmt.Println(gotext.Get("Shell error"), ":", err.Error())
	}
	sort.SliceStable(versions, func(i, j int) bool { return versions[i].Name < versions[j].Name })

	return versions
}

func UpdatesKeys(versions []CheckupdatesOutput) (keys []string) {
	for _, data := range versions {
		keys = append(keys, data.Name)
	}
	return keys
}
