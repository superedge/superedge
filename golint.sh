#!/bin/bash
export OLDPATH=`pwd`
mkdir -p $GOPATH/src/golang.org/x/
cd $GOPATH/src/golang.org/x/ 
git clone https://github.com/golang/lint.git
cd $GOPATH/src/golang.org/x/lint/golint
go install
cd $OLDPATH
golint ./... > golint.result
count=`cat golint.result|wc -l`
if [ $count -gt 0 ]
then exit 1
else exit 0
fi
