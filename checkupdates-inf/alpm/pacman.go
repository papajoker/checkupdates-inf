package alpm

import (
	"bufio"
	"bytes"
	"checkupdates-inf/theme"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type CheckupdatesOutput struct {
	Name         string
	VersionLocal string
	Version      string
}

func (c CheckupdatesOutput) String() string {
	a, _ := VersionColor(c.Version, c.VersionLocal, theme.ColorWarning)
	b, prefix := VersionColor(c.VersionLocal, c.Version, theme.ColorGreen)
	return fmt.Sprintf("%-32s %24s -> %s\t%s", c.Name, a, b, prefix)
}

func VersionColor(installed, next_version string, color string) (v string, prefix string) {
	prefix = ""
	matchs := strings.SplitN(installed, ":", 2)
	if len(matchs) > 1 {
		prefix = "(" + matchs[0] + ":)"
		installed = matchs[1]
	}
	matchs = strings.SplitN(next_version, ":", 2)
	if len(matchs) > 1 {
		if prefix == "" {
			prefix = "(" + matchs[0] + ":)"
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
		//except IndexError:
		//	break
	}
	v = next_version + theme.ColorNone
	return v, prefix
}

// confn = "/etc/pacman.conf"
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

func Checkupdates() (versions []CheckupdatesOutput) {
	var buffer bytes.Buffer
	cmd := exec.Command("checkupdates")
	cmd.Stdout = &buffer
	_ = cmd.Run()
	// log.Printf("checkupdate output: %s", buffer.String())
	out := buffer.String()
	for _, s := range strings.Split(out, "\n") {
		if len(s) < 2 {
			continue
		}
		data := strings.Split(s, " ")
		if len(data) != 4 {
			continue
		}
		versions = append(versions, CheckupdatesOutput{data[0], data[1], data[3]})
	}
	return versions
}

func UpdatesKeys(versions []CheckupdatesOutput) (keys []string) {
	for _, data := range versions {
		keys = append(keys, data.Name)
	}
	return keys
}
