package main

import (
	"checkupdates-inf/alpm"
	"checkupdates-inf/theme"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	getV := flag.Bool("V", false, "Version")
	flagDownload := flag.Bool("d", false, Lg.T("Download packages"))
	noUseColor := flag.Bool("nc", false, Lg.T("No color"))
	addFakes := flag.String("fakes", "", Lg.T("fake packages deleted")+" (-fakes 'chromium firefox mariadb vi nano pikaur')")
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

	if *flagDownload && os.Getuid() != 0 {
		println(Lg.T("Error"), Lg.T("you cannot perform this operation unless you are root"+"."))
		os.Exit(3)
	}

	fmt.Printf("%vCheckupdates %s...%v\n\n", theme.ColorGray, Lg.T("command"), theme.ColorNone)
	updates := alpm.Checkupdates(*flagDownload)
	println("\n")
	if len(updates) == 0 {
		os.Exit(2)
	}
	maxName := DisplayVersions(updates)

	directory := fmt.Sprintf("/tmp/checkup-db-%d", os.Getuid())
	fmt.Printf("\n%v#%s : %s%v\n\n", theme.ColorGray, Lg.T("Database directory"), directory, theme.ColorNone)
	pkgsSync := alpm.Load(directory+"/sync", repos)
	pkgsLocal := alpm.Load("/var/lib/pacman/sync", repos)

	DisplayUpdates(updates, pkgsSync, strconv.Itoa(maxName))

	for _, k := range strings.Fields(*addFakes) {
		delete(pkgsSync, k)
	}

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
		tpl := "  %v%-" + strconv.Itoa(maxName) + "s%v : %s %s\n"
		for _, pkg := range diff {
			fmt.Printf(
				tpl,
				theme.ColorGreen, pkg.NAME, theme.ColorNone, pkg.DESC, pkg.URL,
			)
		}
	}

	diff = []*alpm.Package{}
	for k, pkg := range pkgsLocal {
		if _, ok := pkgsSync[k]; !ok {
			diff = append(diff, pkg)
		}
	}

	/*
		if *addFake {
			diff = append(diff, &alpm.Package{NAME: "pacman", DESC: "FAKE", ReplacedBy: "pacman-plus"})
			diff = append(diff, &alpm.Package{NAME: "yay-bin", DESC: "FAKE", ReplacedBy: ""})
			diff = append(diff, &alpm.Package{NAME: "yay", DESC: "FAKE", ReplacedBy: ""})
			diff = append(diff, &alpm.Package{NAME: "systemd", DESC: "FAKE", ReplacedBy: ""})
			diff = append(diff, &alpm.Package{NAME: "mariadb", DESC: "FAKE", ReplacedBy: ""})
			pkgsSync["pacman"].ReplacedBy = "trucMuche!"
		}
	*/
	if len(diff) > 0 {
		fmt.Printf("\n%s :\n", Lg.T("Deleted packages"))
		maxName := 12
		for _, p := range diff {
			if len(p.NAME) > maxName {
				maxName = len(p.NAME)
			}
		}
		tpl := "  %v%-" + strconv.Itoa(maxName) + "s%v : %s %s %s\n"
		for _, pkg := range diff {
			replace := ""
			if replaced, ok := alpm.Replaced(pkg.NAME, pkgsSync); ok {
				pkg.ReplacedBy = replaced.NAME
				replace = fmt.Sprintf(" -> %v%s%v", theme.ColorGreen, replaced.NAME, theme.ColorNone)
			}
			fmt.Printf(
				tpl,
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
				if pkg.ReplacedBy != "" || pkg.IsDep {
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
					if !pkg.IsDep && len(matchs) < 40 {
						desc := alpm.AurRequestExists(pkg.NAME)
						desc = Lg.T(desc)
						fmt.Printf(
							"  %-"+strconv.Itoa(maxName)+"s ? %s ... %s\n", "",
							Lg.T("Is in AUR"), desc,
						)
					}
				}

				if pkg.ReplacedBy == "" {
					if provides, ok := alpm.ProvideBy(pkg.NAME, pkgsSync); ok {
						fmt.Printf(
							"  %-"+strconv.Itoa(maxName)+"s   %s: %v\n", "",
							Lg.T("can replace by"), provides,
						)
					}
				}
				fmt.Println("")
			}
		}
	}

}

func init() {
	Lg = NewLang()
}

// print Checkupdates output
func DisplayVersions(updates []alpm.CheckupdatesOutput) int {
	const mini = 18
	if len(updates) < 1 {
		return mini
	}
	headers := make([]int, 3)
	for _, p := range updates {
		sizes := p.Sizes(mini)
		for k := range headers {
			if sizes[k] > headers[k] {
				headers[k] = sizes[k]
			}
		}
	}
	tpl := "%-" + strconv.Itoa(headers[0]) + "s  %" + strconv.Itoa(headers[1]) + "s -> %-" + strconv.Itoa(headers[2]) + "s\t%s\n"
	for _, p := range updates {
		a, _ := alpm.VersionColor(p.Version, p.VersionLocal, theme.ColorWarning)
		b, epoch := alpm.VersionColor(p.VersionLocal, p.Version, theme.ColorGreen)
		fmt.Printf(
			tpl,
			p.Name, a, b, epoch,
		)
	}
	return headers[0]
}

// display packages detail to update
func DisplayUpdates(updates []alpm.CheckupdatesOutput, pkgs alpm.Packages, size string) {
	if len(updates) < 1 {
		return
	}
	fmt.Printf("\n%d %s :\n", len(updates), Lg.T("Updates"))
	tpl := "  %v%-" + size + "s%v : %s %s\n"
	for _, p := range updates {
		if pkg, ok := pkgs[p.Name]; ok {
			fmt.Printf(
				tpl,
				theme.ColorGreen, pkg.NAME, theme.ColorNone, pkg.Desc(56), pkg.URL,
			)
		}
	}
	println()
}
