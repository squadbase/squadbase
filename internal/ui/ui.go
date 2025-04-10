package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

var (
	primaryStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#3498db")).Bold(true)
	secondaryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#2ecc71"))
	accentStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#9b59b6"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#e74c3c")).Bold(true)
	warningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#f39c12"))
	infoStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#3498db"))
	successStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#2ecc71")).Bold(true)

	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3498db")).
			Padding(0, 1).
			MarginTop(1)
	warningBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#f39c12")).
			Padding(0, 1).
			MarginTop(1)
)

func PrintTitle(subtitle string) {
	fmt.Println()
	pterm.DefaultBigText.WithLetters(
		putils.LettersFromStringWithStyle("SQUAD", pterm.NewStyle(pterm.FgBlue)),
		putils.LettersFromStringWithStyle("BASE", pterm.NewStyle(pterm.FgLightBlue)),
	).Render()

	fmt.Println(secondaryStyle.Render(subtitle))
	fmt.Println()
}

func StartupMessage() {
	pterm.Info.Prefix = pterm.Prefix{
		Text:  "INFO",
		Style: pterm.NewStyle(pterm.FgBlue),
	}
	pterm.Info.Println("Starting up...")
	time.Sleep(700 * time.Millisecond)

	fmt.Println()
	spinner, _ := pterm.DefaultSpinner.
		WithRemoveWhenDone(true).
		WithText("Loading Squadbase CLI...").
		Start()
	time.Sleep(1 * time.Second)
	spinner.Success("Ready!")
}

func PrintStep(current, total int, title string) {
	step := fmt.Sprintf(" STEP %d/%d ", current, total)
	fmt.Printf("\n%s %s\n",
		primaryStyle.Copy().Background(lipgloss.Color("#f0f0f0")).Render(step),
		secondaryStyle.Render(title))
}

func GetPrimaryText(text string) string {
	return primaryStyle.Render(text)
}

func GetSecondaryText(text string) string {
	return secondaryStyle.Render(text)
}

func GetAccentText(text string) string {
	return accentStyle.Render(text)
}

func PrintInfo(message string) {
	fmt.Printf("%s %s\n", infoStyle.Render("ℹ"), message)
}

func PrintSuccess(message string) {
	fmt.Printf("%s %s\n", successStyle.Render("✓"), message)
}

func PrintError(message string) {
	fmt.Printf("%s %s\n", errorStyle.Render("✗"), message)
}

func PrintWarning(message string) {
	fmt.Printf("%s %s\n", warningStyle.Render("⚠"), message)
}

func PrintSummaryBox(title string, items map[string]string) {
	header := primaryStyle.Copy().Underline(true).Render(title)

	var content strings.Builder
	content.WriteString(header + "\n\n")

	longestKey := 0
	for k := range items {
		if len(k) > longestKey {
			longestKey = len(k)
		}
	}

	for k, v := range items {
		padding := longestKey - len(k) + 2
		content.WriteString(fmt.Sprintf("%s%s%s\n",
			secondaryStyle.Render(k+":"),
			strings.Repeat(" ", padding),
			accentStyle.Render(v)))
	}

	fmt.Println(boxStyle.Render(content.String()))
}

func PrintWarningBox(title string, message string) {
	header := warningStyle.Copy().Underline(true).Render(title)

	var content strings.Builder
	content.WriteString(header + "\n\n")
	content.WriteString(message)

	fmt.Println(warningBoxStyle.Render(content.String()))
}

func ShowProgressBar(title string, total int) *pterm.ProgressbarPrinter {
	pb, _ := pterm.DefaultProgressbar.
		WithTitle(title).
		WithTotal(total).
		WithShowElapsedTime(true).
		WithShowCount(true).
		Start()
	return pb
}

func ShowSpinner(text string) *pterm.SpinnerPrinter {
	spinner, _ := pterm.DefaultSpinner.
		WithRemoveWhenDone(false).
		WithText(text).
		Start()
	return spinner
}
