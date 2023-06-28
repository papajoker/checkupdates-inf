#!/usr/bin/env bash

# voir: https://github.com/Jguer/yay/blob/next/Makefile#L127

LANGS=(de es fr it)
POTFILE="default.pot"

cd "./po/" || exit 2
pwd

xgotext -in ../checkupdates-inf/ -out .

for lang in ${LANGS[@]}; do
    echo " - $lang"
	test -f "$lang.po" || msginit -l "$lang.po" -i "${POTFILE}" -o "$lang.po"
	msgmerge -U "$lang.po" "${POTFILE}"
	touch "$lang.po"
done

for lang in ${LANGS[@]}; do
   mkdir -p "../mo/usr/share/locale/${lang}/LC_MESSAGES/"
   msgfmt -o "../mo/usr/share/locale/${lang}/LC_MESSAGES/checkupdates-inf.mo" "$lang.po"
done


echo "Usage dev:"
echo 'LANGUAGE=es LOCALE_PATH="../mo/usr/share/locale" go run *.go -h'