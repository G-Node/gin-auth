#!/bin/bash

# run this script as normal user w/o sudo from within the gin-auth root folder.
set -e

CNOC="\033[0m"
CAOK="\033[32;01m"
CERR="\033[31;01m"
CWRN="\033[33;01m"

GOPATH=/opt/deploy/go

sudo -v -p "Certain commands require sudo access. Please enter your password: "

echo -e "Running in ${CAOK}$PWD $CNOC"
REPO=$(basename $PWD)
if [ "$REPO" != "gin-auth" ]; then
    echo -e "${CERR}* Not in gin-auth *${CNOC}"
    exit 1
fi

echo "Pulling latest changes"
BRANCH=$(sudo -u deploy git rev-parse --abbrev-ref HEAD)

if [ "$BRANCH" != "master" ]; then
    echo -e "${CERR}* Not on branch master${CNOC} [${CWRN}$BRANCH${CNOC}]"
    exit 1
fi

sudo -u deploy git pull origin master

echo "Processing dependencies"
ALLDEPS=$(sudo -u deploy -E /opt/go/bin/go list -f '{{ join .Deps "\n" }}' ./... | sort -u | grep -v -e "github.com/G-Node/gin-auth" -e "golang.org/x/");
STDDEPS=$(sudo -u deploy -E /opt/go/bin/go list std);
EXTDEPS=$(comm -23 <(echo "$ALLDEPS") <(echo "$STDDEPS"))

for dep in "$EXTDEPS"; do
    sudo -u deploy -E GOPATH=$GOPATH /opt/go/bin/go get -v $dep
done

echo "Update server specific config files for goose"
sudo -u deploy cp /opt/deploy/service_conf/gin-auth/dbconf.yml /opt/deploy/gin-auth/resources/conf

echo "Update database scheme to the latest version"
sudo -u deploy -E GOPATH=$GOPATH $GOPATH/bin/goose -path /opt/deploy/gin-auth/resources/conf up

echo "Installing gin-auth"
sudo -u deploy -E GOPATH=$GOPATH /opt/go/bin/go install

echo "Restarting gin-auth"
sudo systemctl daemon-reload
sudo systemctl restart ginauth.service

echo "Reset server specific config files"
sudo -u deploy git checkout resources/conf/dbconf.yml

echo -e "${CAOK}Done${CNOC}."
