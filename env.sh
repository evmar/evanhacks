export PATH=~/projects/src/google_appengine:$PATH
unset GOROOT
unset GOPATH
export GOPATH=$(pwd):$(go env GOPATH)
