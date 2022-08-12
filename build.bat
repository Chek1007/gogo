set name=gogo
rm ./bin/*
for /F %%i in ('git describe --abbrev^=0 --tags') do ( set gt_ver=%%i)
set gt_key=%1

go get sigs.k8s.io/yaml
go generate gogo.go

gox.exe -osarch="linux/amd64 linux/arm64 linux/arm linux/386 windows/amd64 linux/mips64 windows/386 darwin/amd64" -ldflags="-s -w -X 'gogo/v1/cmd.ver=%gt_ver%' -X 'gogo/v1/cmd.k=%gt_key%'" -gcflags="-trimpath=$GOPATH" -asmflags="-trimpath=$GOPATH" -output=".\bin\%name%_{{.OS}}_{{.Arch}}" .
