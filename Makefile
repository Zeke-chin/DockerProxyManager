.PHONY: all darwin-amd64 linux-amd64 windows-amd64 darwin-arm64 linux-arm64 windows-arm64 clean

all: darwin-amd64 linux-amd64 windows-amd64 darwin-arm64 linux-arm64 windows-arm64

darwin-amd64:
	GOOS=darwin	GOARCH=amd64	go	build	-o	target/DPM-darwin-amd64	main.go

linux-amd64:
	GOOS=linux	GOARCH=amd64	go	build	-o	target/DPM-linux-amd64	main.go

windows-amd64:
	GOOS=windows GOARCH=amd64 go build -o target/DPM-windows-amd64.exe main.go

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o target/DPM-darwin-arm64 main.go

linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o target/DPM-linux-arm64 main.go

windows-arm64:
	GOOS=windows GOARCH=arm64 go build -o target/DPM-windows-arm64.exe main.go

clean:
	rm -rf target

# 注意: Makefile中的命令行必须是tab缩进而不是空格缩进。