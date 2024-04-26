//go:generate goversioninfo -64 -icon=packaging/win/exe.ico -manifest=packaging/win/exe.manifest
package main

// After changing the windows .ico:
// - run `go generate` from `source` to create the .syso file
//   (it will use the parameters defined here above)
// - when `go build` finds a .syso file, it embeds it in the .exe
//
// If it says "goversioninfo": executable file not found in $PATH:
// - ensure there is a ~/go/bin/goversioninfo:
// 	 go get "github.com/josephspurrier/goversioninfo/cmd/goversioninfo" (don't use GOOS=windows)
// - ensure ~/go/bin is in the PATH when running `go generate`
//
import (
	"main/app"
)

func main() {
	app.Start()
}
