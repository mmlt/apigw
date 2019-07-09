#!/usr/bin/env bash
#

if [[ -z $GITHUB_API_TOKEN ]]; then
    echo "error: \$GITHUB_API_TOKEN needs to be set"
    exit 1
fi

# check token validity
curl -H "Authorization: token $GITHUB_API_TOKEN" -o /dev/null https://api.github.com/repos/mmlt/apigw
if $?; then
    echo "error: login with \$GITHUB_API_TOKEN failed"
    exit 1
fi


echo "Tag and push code"
git tag $1
git push origin master --tag

echo Add binaries to github and release, see https://help.github.com/en/articles/creating-releases
# TODO automate adding binairies https://developer.github.com/v3/repos/releases/#create-a-release

echo "Upload binary"
FILE=./apigw
URL="https://uploads.github.com/repos/mmlt/apigw/releases/$1/assets?name=apigw"
curl -H "Authorization: token $GITHUB_API_TOKEN" -H "Content-Type: application/octet-stream" --data-binary @"$FILE" $URL

