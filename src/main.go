package main

import (
	"checkupdates-inf/alpm"
	"checkupdates-inf/theme"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

var (
	GitBranch string
	Version   string
	BuildDate string
	GitID     string
)
var (
	Lg Lang
)

func main() {

	getV := flag.Bool("V", false, "version")
	noUseColor := flag.Bool("nc", false, Lg.T("No color"))
	addFake := flag.Bool("fake", false, Lg.T("use fake for tests"))
	flag.Parse()
	if *getV {
		fmt.Println("checkupdates-inf")
		fmt.Printf("\n%s Version: %v %v %v %v\n", filepath.Base(os.Args[0]), Version, GitID, GitBranch, BuildDate)
		os.Exit(0)
	}
	if *noUseColor {
		theme.Reset()
	}

	repos := alpm.ListRepos("/etc/pacman.conf")
	fmt.Printf("%s : %v\n\n", Lg.T("Repos"), repos)

	fmt.Printf("%vCheckupdates %s...%v\n\n", theme.ColorGray, Lg.T("command"), theme.ColorNone)
	updates := alpm.Checkupdates()
	maxName := 12
	for _, pkg := range updates {
		if len(pkg.Name) > maxName {
			maxName = len(pkg.Name)
		}
	}
	for _, p := range updates {
		a, _ := alpm.VersionColor(p.Version, p.VersionLocal, theme.ColorWarning)
		b, prefix := alpm.VersionColor(p.VersionLocal, p.Version, theme.ColorGreen)
		fmt.Printf("%-"+strconv.Itoa(maxName)+"s  %32s -> %s\t%s\n", p.Name, a, b, prefix)
	}
	//updateskeys := alpm.UpdatesKeys(updates)

	directory := fmt.Sprintf("/tmp/checkup-db-%d", os.Getuid())
	fmt.Printf("\n%v#%s : %s%v\n\n", theme.ColorGray, Lg.T("Database directory"), directory, theme.ColorNone)
	pkgsSync := alpm.Load(directory+"/sync", repos)
	pkgsLocal := alpm.Load("/var/lib/pacman/sync", repos)

	fmt.Printf("\n%s :\n", Lg.T("Updates"))
	for _, data := range updates {
		if pkg, ok := pkgsLocal[data.Name]; ok {
			fmt.Printf("  %v%-"+strconv.Itoa(maxName)+"s%v : %s %s\n", theme.ColorGreen, pkg.NAME, theme.ColorNone, pkg.Desc(56), pkg.URL)
		}
	}
	println()

	l := strconv.Itoa(len(Lg.T("Database Next")))
	fmt.Printf("%-"+l+"s : %v%d%v\n", Lg.T("Database Local"), theme.ColorGreen, len(pkgsLocal), theme.ColorNone)
	fmt.Printf("%-"+l+"s : %v%d%v\n", Lg.T("Database Next"), theme.ColorGreen, len(pkgsSync), theme.ColorNone)

	diff := []*alpm.Package{}
	for k, pkg := range pkgsSync {
		if _, ok := pkgsLocal[k]; !ok {
			diff = append(diff, pkg)
		}
	}
	if len(diff) > 0 {
		fmt.Printf("\n%s :\n", Lg.T("New packages"))
		maxName := 12
		for _, p := range diff {
			if len(p.NAME) > maxName {
				maxName = len(p.NAME)
			}
		}
		for _, pkg := range diff {
			fmt.Printf("  %v%-"+strconv.Itoa(maxName)+"s%v : %s %s\n", theme.ColorGreen, pkg.NAME, theme.ColorNone, pkg.DESC, pkg.URL)
		}
	}

	diff = []*alpm.Package{}
	for k, pkg := range pkgsLocal {
		if _, ok := pkgsSync[k]; !ok {
			diff = append(diff, pkg)
		}
	}

	if *addFake {
		diff = append(diff, &alpm.Package{NAME: "pacman", DESC: "FAKE", ReplacedBy: "pacman-plus"})
		diff = append(diff, &alpm.Package{NAME: "yay-bin", DESC: "FAKE", ReplacedBy: ""})
		diff = append(diff, &alpm.Package{NAME: "yay", DESC: "FAKE", ReplacedBy: ""})
		diff = append(diff, &alpm.Package{NAME: "systemd", DESC: "FAKE", ReplacedBy: ""})
		pkgsSync["pacman"].ReplacedBy = "trucMuche!"
	}
	if len(diff) > 0 {
		fmt.Printf("\n%s :\n", Lg.T("Deleted packages"))
		maxName := 12
		for _, p := range diff {
			if len(p.NAME) > maxName {
				maxName = len(p.NAME)
			}
		}
		for _, pkg := range diff {
			replace := ""
			if replaced, ok := alpm.Replaced(pkg.NAME, pkgsSync); ok {
				pkg.ReplacedBy = replaced.NAME
				replace = fmt.Sprintf(" -> %v%s%v", theme.ColorGreen, replaced.NAME, theme.ColorNone)
			}
			fmt.Printf(
				"  %v%-"+strconv.Itoa(maxName)+"s%v : %s %s %s\n",
				theme.ColorWarning, pkg.NAME, theme.ColorNone, pkg.Desc(56), pkg.URL,
				replace,
			)
		}
		println()

		if matchs, ok := alpm.LocalParse(directory+"/local", diff); ok {
			fmt.Printf("\n%s :\n", Lg.T("Deleted packages but INSTALLED"))
			maxName := 12
			for _, p := range matchs {
				if len(p.NAME) > maxName {
					maxName = len(p.NAME)
				}
			}
			for _, pkg := range matchs {
				color := theme.ColorRed
				if pkg.ReplacedBy != "" {
					color = theme.ColorBold
				}
				fmt.Printf(
					"  %v%-"+strconv.Itoa(maxName)+"s%v : %s %s\n",
					color, pkg.NAME, theme.ColorNone, pkg.DESC, pkg.URL,
				)
				if pkg.ReplacedBy != "" {
					fmt.Printf(
						"  %-"+strconv.Itoa(maxName)+"s -> %v%s%v (%s)\n",
						" ", theme.ColorGreen, pkg.ReplacedBy, theme.ColorNone,
						Lg.T("replaced by"),
					)
				} else {
					desc := alpm.AurRequestExists(pkg.NAME)
					fmt.Printf(
						"  %-"+strconv.Itoa(maxName)+"s ? %s ... %s\n", "",
						Lg.T("Is in AUR"), desc,
					)
				}
				fmt.Println("")
			}
		}
	}

}

func init() {
	Lg = NewLang()
}