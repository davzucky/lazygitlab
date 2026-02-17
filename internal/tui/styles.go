package tui

import (
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const envAccentColor = "LAZYGITLAB_ACCENT"

var hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type markdownStyles struct {
	heading lipgloss.Style
	em      lipgloss.Style
	strong  lipgloss.Style
	code    lipgloss.Style
	link    lipgloss.Style
	quote   lipgloss.Style
	fence   lipgloss.Style
	mermaid lipgloss.Style
	warn    lipgloss.Style
}

type styles struct {
	app            lipgloss.Style
	sidebar        lipgloss.Style
	sidebarActive  lipgloss.Style
	panel          lipgloss.Style
	header         lipgloss.Style
	status         lipgloss.Style
	helpPopup      lipgloss.Style
	errorPopup     lipgloss.Style
	selectedRow    lipgloss.Style
	normalRow      lipgloss.Style
	secondary      lipgloss.Style
	title          lipgloss.Style
	dim            lipgloss.Style
	errorText      lipgloss.Style
	topLevelBorder lipgloss.Border
	markdown       markdownStyles
}

func newStyles() styles {
	accent := resolveAccentColor()
	muted := lipgloss.AdaptiveColor{Light: "242", Dark: "245"}
	bg := lipgloss.AdaptiveColor{Light: "255", Dark: "236"}
	normalText := lipgloss.AdaptiveColor{Light: "238", Dark: "252"}
	border := lipgloss.AdaptiveColor{Light: "248", Dark: "240"}
	errorAccent := lipgloss.AdaptiveColor{Light: "160", Dark: "196"}
	errorBg := lipgloss.AdaptiveColor{Light: "224", Dark: "52"}
	statusText := lipgloss.AdaptiveColor{Light: "255", Dark: "255"}

	return styles{
		app: lipgloss.NewStyle().Padding(0, 1),
		sidebar: lipgloss.NewStyle().
			Padding(1, 1).
			Foreground(normalText),
		sidebarActive: lipgloss.NewStyle().
			PaddingLeft(0).
			Foreground(accent).
			Bold(true),
		panel: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(border).
			Padding(0, 1),
		header: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),
		status: lipgloss.NewStyle().
			Foreground(statusText).
			Background(accent).
			Padding(0, 1),
		helpPopup: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Background(bg).
			Padding(1, 2).
			Width(70),
		errorPopup: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorAccent).
			Background(errorBg).
			Foreground(statusText).
			Padding(1, 2).
			Width(70),
		selectedRow: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),
		normalRow: lipgloss.NewStyle().Foreground(normalText),
		secondary: lipgloss.NewStyle().Foreground(muted),
		title:     lipgloss.NewStyle().Bold(true).Foreground(accent),
		dim:       lipgloss.NewStyle().Foreground(muted),
		errorText: lipgloss.NewStyle().Foreground(errorAccent),
		markdown: markdownStyles{
			heading: lipgloss.NewStyle().Bold(true).Foreground(accent),
			em:      lipgloss.NewStyle().Italic(true),
			strong:  lipgloss.NewStyle().Bold(true),
			code: lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "229"}).
				Background(lipgloss.AdaptiveColor{Light: "252", Dark: "236"}),
			link:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "26", Dark: "81"}).Underline(true),
			quote:   lipgloss.NewStyle().Foreground(muted),
			fence:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "243", Dark: "244"}),
			mermaid: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "30", Dark: "86"}).Bold(true),
			warn:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "166", Dark: "214"}).Italic(true),
		},
	}
}

func resolveAccentColor() lipgloss.TerminalColor {
	if raw := strings.TrimSpace(os.Getenv(envAccentColor)); raw != "" {
		if isLikelyTerminalColor(raw) {
			return lipgloss.Color(raw)
		}
	}
	return lipgloss.AdaptiveColor{Light: "25", Dark: "39"}
}

func isLikelyTerminalColor(value string) bool {
	if hexColorPattern.MatchString(value) {
		return true
	}
	if strings.HasPrefix(value, "#") {
		return false
	}
	if len(value) == 0 || len(value) > 3 {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
