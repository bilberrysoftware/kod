#!/bin/sh
echo "This file must be ran from the root folder."

echo 'deb [trusted=yes] https://repo.goreleaser.com/apt/ /' | sudo tee /etc/apt/sources.list.d/goreleaser.list
sudo apt update
sudo apt install goreleaser

goreleaser release --snapshot --clean

cp ./dist/kod_linux_amd64_v1/kod ./dist/kod
chmod u+x ./dist/kod
./dist/kod version

