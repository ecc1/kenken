#!/bin/sh -e

usage() {
    echo "Usage: $0 [date]"
    exit 1
}

base_url='http://app.kenkenpuzzle.com/kenken/puzzles/NYTimes'

case $# in
    0) date="today" ;;
    1) case "$1" in
	   -*) usage ;;
	   *) date="$1" ;;
       esac
       ;;
    *) usage ;;
esac

cd /tmp

year=$(date --date "$date" +'%Y')
month=$(date --date "$date" +'%b')
month_day=$(date --date "$date" +'%b00%d')

kinds="4x4Easy 4x4Medium 6x6Easy 6x6Medium 6x6Hard 8x8Hard"

for k in $kinds; do
    file="NYTimes$k$month_day.txt"
    if [ ! -f "$file" ]; then
	url="$base_url/$year/$month/$file"
	echo "[downloading $url]"
	wget --quiet "$url"
    fi
    kenken "$file"
done
