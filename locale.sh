#!/usr/bin/env bash

LANGS=(de es fr it)
POTFILE="default.pot"

cd "./po/" || exit 2


xgotext -in ../checkupdates-inf/ -out .

for lang in ${LANGS[@]}; do
    echo " - $lang"
	test -f "$lang.po" || msginit -l "$lang.po" -i "${POTFILE}" -o "$lang.po"
	msgmerge -U "$lang.po" "${POTFILE}"
	touch "$lang.po"
done


for lang in ${LANGS[@]}; do
   mkdir -p "../src/locale/${lang}/LC_MESSAGES/"
   msgfmt -o "../src/locale/${lang}/LC_MESSAGES/checkupdates-inf.mo" "$lang.po"
done


echo "Usage dev:"
echo 'LANGUAGE=es go run *.go -h'