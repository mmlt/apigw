#!/usr/bin/env bash
#
git tag $1
git push origin master --tag

echo Add binaries to github and release, see https://help.github.com/en/articles/creating-releases
# TODO automate adding binairies https://developer.github.com/v3/repos/releases/#create-a-release
