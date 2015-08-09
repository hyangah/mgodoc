#!/bin/bash

# Copyright 2015 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
#
# Script to create assets/go.zip from the goroot.
set -e

if [ ! -f zip.bash ]; then
	echo 'zip.bash must be run from $GOPATH/src/github.com/hyangah/mgodoc/godoc' 1>&2
	exit 1
fi

if [ -z $GOROOT ]; then
	export GOROOT=`go env GOROOT`
fi

CWD=`pwd`
mkdir -p assets
rm -f assets/go.zip

pushd ${GOROOT}/..; zip -q -r ${CWD}/assets/go.zip ./go -i \*.go -i \*.html -i \*.css -i \*.js -i \*.txt -i \*.c -i \*.h -i \*.s -i \*.png -i \*.jpg -i \*.sh -i \*/favicon.ico \*.article; popd;
