#!/usr/bin/env python
"""
V0.4    06-18
"""
from enum import Enum
import json
import os
import gzip
from pathlib import Path
import subprocess
from urllib.request import urlopen


class Colors(Enum):
    BLUE = '\033[94m'
    GREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    GRAY = '\033[90m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

    def version(self, installed: str, next_version: str) -> str:
        for i, char in enumerate(installed):
            try:
                if char != next_version[i]:
                    next_version = next_version[0:i] + self.value + next_version[i:]
                    break
            except IndexError:
                break
        return f"{next_version}{self.ENDC.value}"

    def format(self, text) -> str:
        return f"{self.value}{text}{self.ENDC.value}"

class Package:
    """ pacman package """
    __slots__ = ("name", "version", "desc", "url", "repo", "update", "alias")
    TAB = 18
    COLOR = Colors.GREEN

    def __init__(self, name: str, repo: str = "local") -> None:
        self.name = name
        self.version = "?"
        self.desc = ""
        self.url = ""
        self.update = False
        self.repo = repo
        self.alias = []

    def __str__(self) -> str:
        repo = f"({self.repo})" if self.repo != "local" else ""
        desc = self.desc if len(self.desc) <= 54 else f"{self.desc[0:53]}…"
        return f'{self.COLOR.value}{self.name:{self.TAB}}{Colors.ENDC.value}   "{desc}"  {self.url}  {repo}'


def get_repos_name():
    proc = subprocess.run("pacman-conf -l", capture_output=True, text=True, shell=True)
    return proc.stdout.split()


def checkupdates(display=True):
    """ shell commande checkupdates """
    proc = subprocess.run("checkupdates", shell=False, capture_output=True, text=True)
    datas = proc.stdout.split()
    if not datas:
        return []
    if display:

        def convert_version(version: str) -> tuple[str, str]:
            versions = version.split(":", 2)
            if len(versions) < 2:
                return version, ""
            return versions[1], f" ({versions[0]}:)"

        maxi = max(len(p) for p in datas[::4])
        maxiv = max(len(p) for p in datas[1::4]) + len(Colors.GREEN.value) + len(Colors.ENDC.value)
        for i in range(0, len(datas), 4):
            line = datas[i:i+4]
            version1, _ = convert_version(line[1])
            version2, prev = convert_version(line[3])
            version = Colors.WARNING.version(version2, version1)
            print(f'{line[0]:{maxi}}\t{version:>{maxiv}} -> {Colors.GREEN.version(version1, version2)} {prev}')
    return datas[::4]


def load_repo(directory_check, repo, updates, use_replaces=False):
    """ parse pacman .db file and return generator """
    pkg = None
    try:
        with gzip.open(directory_check / f"{repo}.db", 'rt') as f_files:
            for line in f_files:
                line = line.strip()
                if not line:
                    continue
                match line:
                    case '%NAME%':
                        if pkg:
                            # if pkg.name == "python":
                            #     pkg.alias.append("python-cacheyou")
                            yield pkg
                        pkg = Package(name=next(f_files).rstrip(), repo=repo)
                        pkg.update = pkg.name in updates
                    case '%DESC%':
                        pkg.desc = next(f_files).rstrip()
                    case '%URL%':
                        pkg.url = next(f_files).rstrip()
                    case '%VERSION%':
                        pkg.version = next(f_files).rstrip()
                    case '%REPLACES%' if use_replaces:
                        while key := next(f_files).rstrip():
                            pkg.alias.append(key.split("<", 2)[0].split("=", 2)[0])
            yield pkg
    except gzip.BadGzipFile:
        print("! Bad Gz File :", directory_check / f"{repo}.db")
        raise


def get_installed(rm):
    """ Installed but not in (next) repos"""
    removeds = [p.name for p in rm]
    directory = Path(f"/tmp/checkup-db-{os.getuid()}/local/")

    for search in [f"{p}-*/desc" for p in removeds]:
        for pkg_desc in directory.glob(search):
            with pkg_desc.open() as f_files:
                for line in f_files:
                    line = line.strip()
                    if not line:
                        continue
                    match line:
                        case '%PACKAGER%':
                            if pkg.name in removeds:
                                yield pkg
                        case '%NAME%':
                            pkg = Package(name=next(f_files).rstrip())
                        case '%DESC%':
                            pkg.desc = next(f_files).rstrip()
                        case '%URL%':
                            pkg.url = next(f_files).rstrip()
                        case '%VERSION%':
                            pkg.version = next(f_files).rstrip()


def get_diffs(pkgs_local, pkgs_news) -> tuple[list[Package], list[Package]]:
    """ list new packages and removed """
    rm = []
    new = []
    keys = [p.name for p in pkgs_news]
    for pkg in pkgs_local:
        if pkg.name not in keys:
            rm.append(pkg)
    if rm:
        rm.sort(key=lambda x: x.name)
    keys = [p.name for p in pkgs_local]
    for pkg in pkgs_news:
        if pkg.name not in keys:
            new.append(pkg)
    if new:
        new.sort(key=lambda x: x.name)
    return rm, new


if __name__ == '__main__':

    repos = get_repos_name()
    print("# Dépôts:", repos)
    print("# checkupdates ...")
    updates = checkupdates()

    directory_check = Path(f"/tmp/checkup-db-{os.getuid()}/sync/")
    pkgs_news = []
    for repo in repos:
        print(Colors.GRAY.format(f"# read repo {repo} ..."))
        pkgs_news.extend(load_repo(directory_check, repo, updates, use_replaces=True))
    pkgs_news.sort(key=lambda x: x.name)

    print("\nUpdates:")
    for pkg in (p for p in pkgs_news if p.update):
        print(" ", pkg)
    print()

    directory_check = Path("/var/lib/pacman/sync/")
    pkgs_local = []
    for repo in repos:
        print(Colors.GRAY.format(f"# read local repo {repo}..."))
        pkgs_local.extend(load_repo(directory_check, repo, []))

    print("\n ---- Dépôts Différences ----\n")
    rm, new = get_diffs(pkgs_local, pkgs_news)
    print(len(pkgs_local), "paquets avant mise à jour")
    print(len(pkgs_news), "paquets après mise à jours")

    if new:
        print("\nNouveaux paquets disponibles:")
        for p in new:
            print(" ", p)

    if rm:
        print("\nPaquets prochainement supprimés:")
        Package.COLOR = Colors.WARNING
        for p in rm:
            replace = ""
            if replaces := [a for a in pkgs_news if p.name in a.alias]:
                replace = f" (Replace by:{ Colors.GREEN.format(replaces[0].name)})"
            print(" ", p, replace)

        print()
        # print("# read pacman database ...")
        print("\n ---- Installé mais supprimé de nos dépôts après l'update ----\n")
        pkgs_local = list(get_installed(rm))

        for pkg in pkgs_local:
            Package.COLOR = Colors.FAIL
            print(" ", pkg)
            if replace := [p for p in pkgs_news if pkg.name in p.alias]:
                print("      Remplacé par:", *replace)
            else:
                if len(pkgs_local) > 20:
                    continue
                url_aur = f"https://aur.archlinux.org/rpc/?v=5&type=info&arg[]={pkg.name}"
                with urlopen(url_aur) as response:
                    if response.getcode() != 200:
                        continue
                    data = json.loads(response.read().decode("utf-8"))
                    if data["resultcount"] != 1:
                        continue
                    pkg_aur = Package(pkg.name, "AUR")
                    pkg_aur.url = f"https://aur.archlinux.org/packages/{pkg.name}"
                    pkg_aur.version = data["results"][0]["Version"]
                    Package.COLOR = Colors.ENDC
                    print("     ", pkg_aur)
        if len(pkgs_local) > 0:
            print("\nTODO: Voir si est maintenant dans AUR ou remplacé ou, à supprimer...")
