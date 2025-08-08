package service

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"runtime"
)

func NewTaggedProgressBar(n int, tag int) *progressbar.ProgressBar {
	fmt.Println("")

	pc, _, _, _ := runtime.Caller(1) // use Caller(1) to get the calling function's name
	funcName := runtime.FuncForPC(pc).Name()

	bar := progressbar.NewOptions(n,
		progressbar.OptionSetDescription(fmt.Sprintf(" ▶ %s | Period:%d", funcName, tag)),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(100),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "#",
			SaucerHead:    ">",
			SaucerPadding: "-",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	return bar
}
