package alpm

import (
	"bufio"
	"bytes"
	"checkupdates-inf/theme"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
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

func Checkupdates() (versions []CheckupdatesOutput, keys []*string) {
	var buffer bytes.Buffer
	cmd := exec.Command("checkupdates")
	cmd.Stdout = &buffer
	_ = cmd.Run()
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
	sort.SliceStable(versions, func(i, j int) bool { return versions[i].Name < versions[j].Name })
	keys = make([]*string, len(versions))
	for _, v := range versions {
		keys = append(keys, &v.Name)
	}
	return versions, keys
}

func UpdatesKeys(versions []CheckupdatesOutput) (keys []string) {
	for _, data := range versions {
		keys = append(keys, data.Name)
	}
	return keys
}
