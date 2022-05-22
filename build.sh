pushd ../pflow-editor/
npm run build
popd
rm -rf ./public/editor
mv ../pflow-editor/build ./public/editor
rice embed-go
go build #-ldflags "-s"

