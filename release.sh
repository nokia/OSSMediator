#!/bin/sh
echo Releasing OSSMediator
VERSION=$(cat VERSION | sed -e "s/-SNAPSHOT$//")
echo Releasing OSSMediator-$VERSION
echo $VERSION > VERSION
git add VERSION
git commit -m "Bumping OSSMediator to $VERSION"
git push
git tag $VERSION
git push origin $VERSION
VERSION=$(echo $VERSION | awk -F. '{$NF = $NF + 1;} 1' | sed 's/ /./g')
echo Bumping OSSMediator to $VERSION-SNAPSHOT
echo $VERSION-SNAPSHOT > VERSION
git add VERSION
git commit -m "Bumping OSSMediator to $VERSION-SNAPSHOT"
git push
echo Release successful.
