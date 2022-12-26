package print

import "github.com/pterm/pterm"

func Page(title string) {
	print("\033[H\033[2J")
	if title == "Exited" {
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgLightRed)).WithFullWidth().Println(title)
	} else {
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).WithFullWidth().Println(title)
	}
	pterm.Println()
}

func Success(message string) {
	pterm.Success.Println(message)
}

func Error(message string) {
	pterm.Error.Println(message)
}

func Info(message string) {
	pterm.Info.Println(message)
}

func ProgressBar(title string, total int) (pb *pterm.ProgressbarPrinter) {
	pb, _ = pterm.DefaultProgressbar.WithTotal(total).WithTitle(title).WithRemoveWhenDone(true).Start()
	return pb
}
