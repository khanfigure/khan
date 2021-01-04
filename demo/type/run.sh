#!/bin/sh
rm -rf demo
mkdir demo
cp cmd*.sh demo
cd demo
for file in cmd*.sh ; do
	echo -n "$ "
	cat $file | ../type
	sh $file
	echo
done
