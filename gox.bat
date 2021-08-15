python FingerprintUpdate.py
set name=getitle
gox.exe -osarch="linux/amd64 linux/arm64 linux/386 windows/amd64 linux/mips64 windows/386 darwin/amd64" -ldflags="-s -w" -gcflags="-trimpath=$GOPATH" -asmflags="-trimpath=$GOPATH" -output=".\bin\%name%_{{.OS}}_{{.Arch}}_v%1" .\src\main\

go-strip -f ./bin/%name%_windows_386_v%1.exe -a -output ./bin/%name%_windows_386_v%1.exe > nul
go-strip -f ./bin/%name%_windows_amd64_v%1.exe -a -output ./bin/%name%_windows_amd64_v%1.exe > nul
go-strip -f ./bin/%name%_linux_386_v%1 -a -output ./bin/%name%_linux_386_v%1 > nul
go-strip -f ./bin/%name%_linux_arm64_v%1 -a -output ./bin/%name%_linux_arm64_v%1 > nul
go-strip -f ./bin/%name%_linux_amd64_v%1 -a -output ./bin/%name%_linux_amd64_v%1 > nul
go-strip -f ./bin/%name%_linux_mips64_v%1 -a -output ./bin/%name%_linux_mips64_v%1 > nul
go-strip -f ./bin/%name%_darwin_amd64_v%1 -a -output ./bin/%name%_darwin_amd64_v%1 > nul



upxs -1 -k -o ./bin/%name%_windows_386_v%1_upx.exe ./bin/%name%_windows_386_v%1.exe
upxs -1 -k -o ./bin/%name%_windows_amd64_v%1_upx.exe ./bin/%name%_windows_amd64_v%1.exe
upxs -1 -k -o ./bin/%name%_linux_386_v%1_upx ./bin/%name%_linux_386_v%1
upxs -1 -k -o ./bin/%name%_linux_amd64_v%1_upx ./bin/%name%_linux_amd64_v%1
upxs -1 -k -o ./bin/%name%_linux_arm64_v%1_upx ./bin/%name%_linux_arm64_v%1
upxs -1 -k -o ./bin/%name%_darwin_amd64_v%1_upx ./bin/%name%_darwin_amd64_v%1

limelighter -I ./bin/%name%_windows_amd64_v%1.exe -O ./bin/%name%_windows_amd64_sangfor_v%1.exe -Domain www.sangfor.com
rm *.sangfor.*

tar cvf release/%name%v%1.tar bin/* README.md gtfilter.py UPDATELOG.md