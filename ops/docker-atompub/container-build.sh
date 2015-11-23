#!/usr/bin/env sh

cd /go/src/github.com/ironcamel/go.atompub
echo "Fetching..."
go get
echo "Installing..."
go install

echo "Copying bin..."
mkdir -p bin
cp /go/bin/go.atompub bin/

echo "Done."

exit 0;
