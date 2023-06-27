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

var GitBranch string
var Version string
var BuildDate string
var GitID string

func main() {

	getV := flag.Bool("V", false, "version")
	noUseColor := flag.Bool("nc", false, "No color")
	addFake := flag.Bool("fake", false, "use fake for tests")
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
	fmt.Printf("Repos : %v\n\n", repos)

	fmt.Printf("%vCheckupdates command...%v\n\n", theme.ColorGray, theme.ColorNone)
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
	fmt.Printf("\n%v#Database directory : %s%v\n\n", theme.ColorGray, directory, theme.ColorNone)
	pkgsSync := alpm.Load(directory+"/sync", repos)
	pkgsLocal := alpm.Load("/var/lib/pacman/sync", repos)

	println("\nUpdates :")
	for _, data := range updates {
		if pkg, ok := pkgsLocal[data.Name]; ok {
			fmt.Printf("  %v%-"+strconv.Itoa(maxName)+"s%v : %s %s\n", theme.ColorGreen, pkg.NAME, theme.ColorNone, pkg.Desc(56), pkg.URL)
		}
	}
	println()

	fmt.Printf("Database Local : %v%d%v\n", theme.ColorGreen, len(pkgsLocal), theme.ColorNone)
	fmt.Printf("Database Next :  %v%d%v\n", theme.ColorGreen, len(pkgsSync), theme.ColorNone)

	diff := []*alpm.Package{}
	for k, pkg := range pkgsSync {
		if _, ok := pkgsLocal[k]; !ok {
			diff = append(diff, pkg)
		}
	}
	if len(diff) > 0 {
		println("\nNew packages :")
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
		println("\nDeleted packages :")
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
			println("\nDeleted packages but INSTALLED:\n")
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
						"  %-"+strconv.Itoa(maxName)+"s -> %v%s%v (replaced by)\n",
						" ", theme.ColorGreen, pkg.ReplacedBy, theme.ColorNone,
					)
				} else {
					desc := alpm.AurRequestExists(pkg.NAME)
					fmt.Printf("  %-"+strconv.Itoa(maxName)+"s ? Is in AUR ... %s\n", "", desc)
				}
				fmt.Println("")
			}
		}
	}

}
