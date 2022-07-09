pushd ../pflow-editor/
npm run build
popd
rm -rf ./public/p
rm -rf ./public
mkdir ./public
mv ../pflow-editor/build ./public/p
rice embed-go
go build #-ldflags "-s"

