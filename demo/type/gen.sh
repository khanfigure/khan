#!/bin/sh
go build
rm /tmp/file.txt
userdel bozo
groupdel bozo
termtosvg demo.svg -c ./run.sh -D 600000ms
