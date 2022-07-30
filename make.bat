set name=getitle
rm ./bin/*
for /F %%i in ('git describe --abbrev^=0 --tags') do ( set gt_ver=%%i)
set gt_key=%1
go get sigs.k8s.io/yaml
go generate getitle.go

gox.exe -osarch="linux/amd64 linux/arm64 linux/386 windows/amd64 linux/mips64 windows/386 darwin/amd64 darwin/arm64" -ldflags="-s -w -X 'getitle/v1/cmd.ver=%gt_ver%' -X 'getitle/v1/cmd.k=%gt_key%'" -gcflags="-trimpath=$GOPATH" -asmflags="-trimpath=$GOPATH" -output=".\bin\%name%_{{.OS}}_{{.Arch}}" .

