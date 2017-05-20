go get github.com/Masterminds/glide
cd $GOPATH/src/github.com/Masterminds/glide
git checkout tags/v0.11.1
go install

glide --version

cd $GOPATH/src/github.com/ubclaunchpad/cumulus
