#!/bin/bash

set -e

APP_NAME=screen-sleeper
DIST_DIR=dist
MAIN_DIR=.
BUILD_TIME=$(date "+%F %T %Z")

WINDOWS_ARCH=386,amd64
LINUX_ARCH=386,amd64,arm,arm64

CURRENT_ARCH=$(go env GOARCH)

APP_VERSION=

getversion() {
	git status > /dev/null 2>&1
	if [ $? -eq 0 ]; then
		head_id=$(git log --pretty=oneline | head -n 1 | cut -d " " -f 1)
		latest_tag=$(git tag --sort=v:refname | tail -n 1)
		if [ -n "$latest_tag" ]; then
			tag_id=$(git log --pretty=oneline $latest_tag | head -n 1 | cut -d " " -f 1)
			if [ "$tag_id" != "$head_id" ]; then
				APP_VERSION=$latest_tag+${head_id::12}
			else
				APP_VERSION=$latest_tag
			fi
		else
			APP_VERSION=dev-${head_id::12}
		fi
	else
		APP_VERSION=dev
	fi
}

setgitconfig() {
	if [ -z "$(git config --global --get-all versionsort.prereleaseSuffix)" ]; then
		git config --global --add versionsort.prereleaseSuffix -alpha
		[ $? -eq 0 ] && echo "added prerelease suffix: -alpha"
		git config --global --add versionsort.prereleaseSuffix -beta
		[ $? -eq 0 ] && echo "added prerelease suffix: -beta"
		git config --global --add versionsort.prereleaseSuffix -rc
		[ $? -eq 0 ] && echo "added prerelease suffix: -rc"
	fi
}

prebuild() {
	go generate
}

gobuild() {
	echo "building $1 $2"
	ext=""
	append_suffix=""
	if [ "$1" == "windows" ]; then
		ext=".exe"
	fi
	if [ -n "$exe_suffix" ]; then
		append_suffix=_$1_$2
	fi
	target_dir=$DIST_DIR/$1/$2
	rm -rf $target_dir
	if [ -z "$build_native" ]; then
		export GOOS=$1
		export GOARCH=$2
	fi
	go build -o $target_dir/${APP_NAME}${append_suffix}${ext} \
	-ldflags \
	"\
	-s -w \
	-X 'main.appName=${APP_NAME}' \
	-X 'main.appVersion=${APP_VERSION}' \
	-X 'main.buildTime=${BUILD_TIME}' \
	" \
	$MAIN_DIR
}

showhelp() {
	echo "Usage: build.sh [-m] [-w] [-l] [-s] [-p]"
	echo "    -m  build macos executable of amd64"
	echo "    -w  build windows executable of current arch ($CURRENT_ARCH)"
	echo "    -w[=<arch>,...]  build windows executables of specific arch ($WINDOWS_ARCH)"
	echo "    -l  build linux executable of current arch ($CURRENT_ARCH)"
	echo "    -l[=<arch>,...]  build linux executables of specific arch ($LINUX_ARCH)"
	echo "    -s  append os type and arch suffix of executable name (use 'foo_linux_amd64' instead of 'foo')"
	echo "    -p  do not perform prebuild"
	echo "    -n  build native arch"
	echo "    --set-git-config  set value of versionsort.prereleaseSuffix"
}

archContains() {
	str=$1
	array=(${str//,/ })
	for var in ${array[@]}
	do
		if [ "$var" == "$2" ]; then
			echo "true"
			return
		fi
	done
}

cd "$( dirname "$0" )"

if [ $# -gt 0 ]; then
	for arg in $*
	do
		case $arg in
			-m)
				build_mac=1
			;;
			-w)
				build_windows=$CURRENT_ARCH
			;;
			-w=*)
				build_windows=${arg#*-w=}
			;;
			-l)
				build_linux=$CURRENT_ARCH
			;;
			-l=*)
				build_linux=${arg#*-l=}
			;;
			-s)
				exe_suffix=1
			;;
			-p)
				donot_prebuild=1
			;;
			-n)
				build_native=1
			;;
			--set-git-config)
				setgitconfig
				exit 0
			;;
			*)
				echo "unknow arg: $arg"
				showhelp
				exit 1
			;;
		esac
	done
else
	showhelp
	exit 0
fi

getversion
if [ -n "$donot_prebuild" ]; then
	echo "skip prebuild"
else
	prebuild
fi
echo "Packing $APP_NAME with version $APP_VERSION"

if [ -n "$build_native" ]; then
	gobuild
	exit 0
fi

export CGO_ENABLED=0
if [ -n "$build_mac" ]; then
	gobuild darwin amd64
fi
if [ -n "$build_linux" ]; then
	array=(${build_linux//,/ })
	for var in ${array[@]}
	do
		cont=$(archContains $LINUX_ARCH $var)
		if [ -n "$cont" ]; then
			gobuild linux $var
		else
			echo "unknow arch $var, skip"
		fi
	done
fi
if [ -n "$build_windows" ]; then
	array=(${build_windows//,/ })
	for var in ${array[@]}
	do
		cont=$(archContains $WINDOWS_ARCH $var)
		if [ -n "$cont" ]; then
			gobuild windows $var
		else
			echo "unknow arch $var, skip"
		fi
	done
fi
