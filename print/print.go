package print

import (
	"bufio"
	"context"
	"fmt"
	"markov-generator/markov"
	"markov-generator/stats"
	"os"
	"strings"

	"github.com/pterm/pterm"
)

var started string

func Page(title string) {
	print("\033[H\033[2J")
	if title == "Exited" {
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgLightRed)).WithFullWidth().Println(title)
	} else if title == "Started" {
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgGreen)).WithFullWidth().Println(title)
	} else {
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).WithFullWidth().Println(title)
	}
	pterm.Println()
}

func Success(message string) {
	pterm.Success.Println(message)
	stats.Log(message)
	fmt.Println()
}

func Error(message string) {
	pterm.Error.Println(message)
	stats.Log(message)
	fmt.Println()
}

func Info(message string) {
	pterm.Info.Println(message)
	fmt.Println()
}

func ProgressBar(title string, total int) (pb *pterm.ProgressbarPrinter) {
	pb, _ = pterm.DefaultProgressbar.WithTotal(total).WithTitle(title).WithRemoveWhenDone(true).Start()
	return pb
}

func Started(text string) {
	Info(text)
	started = text
}

func TerminalInput(cancel context.CancelFunc) {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.ToLower(scanner.Text())

		fmt.Println()

		switch input {
		default:
			Info("Not a command")
		case "exit":
			cancel()
		case "clear":
			Page("Twitch Message Generator")
		case "help":
			t := fmt.Sprintln("[write duration] for latest duration of write loop")
			t += fmt.Sprintln("[zip duration] for latest duration of zip process")
			t += fmt.Sprintln("[started] for when the program started")
			t += fmt.Sprintln("[help] for list of commands")
			t += fmt.Sprintln("[clear] to clear the screen")
			t += fmt.Sprintln("[exit] to exit the program")

			Info(t)
		case "started":
			Info(started)
		case "write duration":
			reports := markov.ReportDurations()
			Info("Latest writing loop duration: " + fmt.Sprint(reports["write duration"].Duration))
		case "zip duration":
			reports := markov.ReportDurations()
			Info("Latest zip duration: " + fmt.Sprint(reports["zip duration"].Duration))
		}
	}
}
