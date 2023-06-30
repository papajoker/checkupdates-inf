#!/usr/bin/env bash

# xgotext -in ./src/ -out ./src/locale/
echo
echo "[-c] : compress binary (55%)"
echo

gomod=$(find . -name "go.mod")
[ -z "$gomod" ] && {
    echo "ERROR: No file go.mod found"
    exit 1
}
project=$(awk '/^module/{print $2}' "$gomod")

dir=$(dirname "$gomod")
echo "project directory: ${dir}"
cd "$dir"
pwd


version="$(git tag -l | tail -n1)"
if [[ -z "$version" ]]; then
    echo "no git version :("
    version="0.0.1"
else
    [[ ${version::1} == 'v' ]] && version=${version:1}
fi

commit="$(git rev-parse --short HEAD 2>/dev/null)"

echo ""
go vet || {
    echo -e "\n🔴 ERROR ⚡️ no build 🔴\n"
    exit 1
}

go build  \
    -ldflags \
    "-s -w
    -X main.GitID=${commit} \
    -X main.GitBranch=$(git branch --show-current 2>/dev/null) \
    -X main.Version=${version} \
    -X main.BuildDate=$(date +%F)" \
    -o "../${project}"


cd ..
[ -f "${project}_${version}_linux-64bit.tar.gz" ] && \
    rm "${project}_${version}_linux-64bit.tar.gz"
if [[ "$1" == "-c" ]]; then
    upx -9 "${project}"
    tar -czvf "${project}_${version}_linux-64bit.tar.gz" "./${project}" "./${project}"
fi

echo
pwd
ls -lh "${project}" --color=always --file-type
[ -f "${project}_${version}_linux-64bit.tar.gz" ] && \
    ls -lh "${project}_${version}_linux-64bit.tar.gz" --color=always --file-type
