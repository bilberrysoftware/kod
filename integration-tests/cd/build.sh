#!/bin/sh
echo "This file must be ran from the root folder."
go build -o kod main.go
./kod
