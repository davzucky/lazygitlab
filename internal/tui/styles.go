package tui

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const envAccentColor = "LAZYGITLAB_ACCENT"
const envThemeForeground = "LAZYGITLAB_THEME_FG"
const envThemeBackground = "LAZYGITLAB_THEME_BG"

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
	palette := resolveThemePalette()

	return styles{
		app: lipgloss.NewStyle().Padding(0, 1),
		sidebar: lipgloss.NewStyle().
			Padding(1, 1).
			Foreground(palette.normalText),
		sidebarActive: lipgloss.NewStyle().
			PaddingLeft(0).
			Foreground(palette.accent).
			Bold(true),
		panel: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(palette.border).
			Padding(0, 1),
		header: lipgloss.NewStyle().
			Foreground(palette.accent).
			Bold(true),
		status: lipgloss.NewStyle().
			Foreground(palette.statusText).
			Background(palette.accent).
			Padding(0, 1),
		helpPopup: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(palette.accent).
			Background(palette.bg).
			Padding(1, 2).
			Width(70),
		errorPopup: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(palette.errorAccent).
			Background(palette.errorBg).
			Foreground(palette.statusText).
			Padding(1, 2).
			Width(70),
		selectedRow: lipgloss.NewStyle().
			Foreground(palette.accent).
			Bold(true),
		normalRow: lipgloss.NewStyle().Foreground(palette.normalText),
		secondary: lipgloss.NewStyle().Foreground(palette.muted),
		title:     lipgloss.NewStyle().Bold(true).Foreground(palette.accent),
		dim:       lipgloss.NewStyle().Foreground(palette.muted),
		errorText: lipgloss.NewStyle().Foreground(palette.errorAccent),
		markdown: markdownStyles{
			heading: lipgloss.NewStyle().Bold(true).Foreground(palette.accent),
			em:      lipgloss.NewStyle().Italic(true),
			strong:  lipgloss.NewStyle().Bold(true),
			code: lipgloss.NewStyle().
				Foreground(palette.codeText).
				Background(palette.codeBg),
			link:    lipgloss.NewStyle().Foreground(palette.link).Underline(true),
			quote:   lipgloss.NewStyle().Foreground(palette.muted),
			fence:   lipgloss.NewStyle().Foreground(palette.fence),
			mermaid: lipgloss.NewStyle().Foreground(palette.mermaid).Bold(true),
			warn:    lipgloss.NewStyle().Foreground(palette.warn).Italic(true),
		},
	}
}

type themePalette struct {
	accent      lipgloss.TerminalColor
	bg          lipgloss.TerminalColor
	normalText  lipgloss.TerminalColor
	muted       lipgloss.TerminalColor
	border      lipgloss.TerminalColor
	errorAccent lipgloss.TerminalColor
	errorBg     lipgloss.TerminalColor
	statusText  lipgloss.TerminalColor
	codeText    lipgloss.TerminalColor
	codeBg      lipgloss.TerminalColor
	link        lipgloss.TerminalColor
	fence       lipgloss.TerminalColor
	mermaid     lipgloss.TerminalColor
	warn        lipgloss.TerminalColor
}

func resolveThemePalette() themePalette {
	fg, bg, ok := resolveThemeFGAndBG()
	if !ok {
		accent := resolveAccentColor()
		muted := lipgloss.AdaptiveColor{Light: "242", Dark: "245"}
		return themePalette{
			accent:      accent,
			bg:          lipgloss.AdaptiveColor{Light: "255", Dark: "236"},
			normalText:  lipgloss.AdaptiveColor{Light: "238", Dark: "252"},
			muted:       muted,
			border:      lipgloss.AdaptiveColor{Light: "248", Dark: "240"},
			errorAccent: lipgloss.AdaptiveColor{Light: "160", Dark: "196"},
			errorBg:     lipgloss.AdaptiveColor{Light: "224", Dark: "52"},
			statusText:  lipgloss.AdaptiveColor{Light: "255", Dark: "255"},
			codeText:    lipgloss.AdaptiveColor{Light: "236", Dark: "229"},
			codeBg:      lipgloss.AdaptiveColor{Light: "252", Dark: "236"},
			link:        lipgloss.AdaptiveColor{Light: "26", Dark: "81"},
			fence:       lipgloss.AdaptiveColor{Light: "243", Dark: "244"},
			mermaid:     lipgloss.AdaptiveColor{Light: "30", Dark: "86"},
			warn:        lipgloss.AdaptiveColor{Light: "166", Dark: "214"},
		}
	}

	accent := resolveAccentColor()
	if _, isAdaptive := accent.(lipgloss.AdaptiveColor); isAdaptive {
		accent = lipgloss.Color(deriveAccentFromFG(fg))
	}

	statusText := "#ffffff"
	if !isDarkColor(fmt.Sprint(accent)) {
		statusText = "#111111"
	}

	return themePalette{
		accent:      accent,
		bg:          lipgloss.Color(bg),
		normalText:  lipgloss.Color(fg),
		muted:       lipgloss.Color(mixHex(fg, bg, 0.50)),
		border:      lipgloss.Color(mixHex(fg, bg, 0.33)),
		errorAccent: lipgloss.Color("#d23f31"),
		errorBg:     lipgloss.Color(mixHex(bg, "#d23f31", 0.12)),
		statusText:  lipgloss.Color(statusText),
		codeText:    lipgloss.Color(mixHex(fg, bg, 0.12)),
		codeBg:      lipgloss.Color(mixHex(bg, fg, 0.08)),
		link:        lipgloss.Color(deriveLinkColor(fg, bg)),
		fence:       lipgloss.Color(mixHex(fg, bg, 0.45)),
		mermaid:     lipgloss.Color(mixHex(deriveAccentFromFG(fg), fg, 0.65)),
		warn:        lipgloss.Color("#c67a00"),
	}
}

func resolveThemeFGAndBG() (string, string, bool) {
	fg := strings.TrimSpace(os.Getenv(envThemeForeground))
	bg := strings.TrimSpace(os.Getenv(envThemeBackground))
	if isLikelyTerminalColor(fg) && isLikelyTerminalColor(bg) {
		return normalizeColorToHex(fg), normalizeColorToHex(bg), true
	}
	if guessedFG, guessedBG, ok := parseColorFGBG(os.Getenv("COLORFGBG")); ok {
		return guessedFG, guessedBG, true
	}
	return "", "", false
}

func parseColorFGBG(value string) (string, string, bool) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return "", "", false
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool { return r == ';' || r == ':' })
	if len(parts) < 2 {
		return "", "", false
	}
	fgPart := parts[len(parts)-2]
	bgPart := parts[len(parts)-1]
	fg := normalizeColorToHex(fgPart)
	bg := normalizeColorToHex(bgPart)
	if fg == "" || bg == "" {
		return "", "", false
	}
	return fg, bg, true
}

func normalizeColorToHex(raw string) string {
	value := strings.TrimSpace(raw)
	if hexColorPattern.MatchString(value) {
		return strings.ToLower(value)
	}
	idx, err := strconv.Atoi(value)
	if err != nil || idx < 0 || idx > 255 {
		return ""
	}
	r, g, b := ansi256ToRGB(idx)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func ansi256ToRGB(index int) (int, int, int) {
	basic := [16][3]int{
		{0, 0, 0}, {205, 49, 49}, {13, 188, 121}, {229, 229, 16},
		{36, 114, 200}, {188, 63, 188}, {17, 168, 205}, {229, 229, 229},
		{102, 102, 102}, {241, 76, 76}, {35, 209, 139}, {245, 245, 67},
		{59, 142, 234}, {214, 112, 214}, {41, 184, 219}, {255, 255, 255},
	}
	if index < 16 {
		entry := basic[index]
		return entry[0], entry[1], entry[2]
	}
	if index >= 232 {
		v := 8 + (index-232)*10
		return v, v, v
	}
	idx := index - 16
	r := idx / 36
	g := (idx % 36) / 6
	b := idx % 6
	steps := [6]int{0, 95, 135, 175, 215, 255}
	return steps[r], steps[g], steps[b]
}

func deriveAccentFromFG(fgHex string) string {
	return mixHex(fgHex, "#3aa0ff", 0.58)
}

func deriveLinkColor(fgHex string, bgHex string) string {
	base := "#2f79d9"
	if !isDarkColor(bgHex) {
		base = "#0055aa"
	}
	return mixHex(base, fgHex, 0.20)
}

func mixHex(a string, b string, weightB float64) string {
	ar, ag, ab, okA := hexToRGB(a)
	br, bg, bb, okB := hexToRGB(b)
	if !okA || !okB {
		if okA {
			return strings.ToLower(a)
		}
		if okB {
			return strings.ToLower(b)
		}
		return "#888888"
	}
	if weightB < 0 {
		weightB = 0
	}
	if weightB > 1 {
		weightB = 1
	}
	weightA := 1 - weightB
	r := int(float64(ar)*weightA + float64(br)*weightB)
	g := int(float64(ag)*weightA + float64(bg)*weightB)
	bv := int(float64(ab)*weightA + float64(bb)*weightB)
	return fmt.Sprintf("#%02x%02x%02x", clampByte(r), clampByte(g), clampByte(bv))
}

func hexToRGB(value string) (int, int, int, bool) {
	if !hexColorPattern.MatchString(value) {
		return 0, 0, 0, false
	}
	r, err := strconv.ParseInt(value[1:3], 16, 64)
	if err != nil {
		return 0, 0, 0, false
	}
	g, err := strconv.ParseInt(value[3:5], 16, 64)
	if err != nil {
		return 0, 0, 0, false
	}
	b, err := strconv.ParseInt(value[5:7], 16, 64)
	if err != nil {
		return 0, 0, 0, false
	}
	return int(r), int(g), int(b), true
}

func isDarkColor(value string) bool {
	r, g, b, ok := hexToRGB(normalizeColorToHex(value))
	if !ok {
		return true
	}
	luminance := (0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)) / 255.0
	return luminance < 0.5
}

func clampByte(value int) int {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return value
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
