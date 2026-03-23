package qtui

import (
	_ "embed"
	"strings"
)

//go:embed style.tqss
var styleTemplate string

type Theme struct {
	Font          string
	Bg            string
	Surface       string
	SurfaceAlt    string
	Sidebar       string
	Toolbar       string
	Border        string
	BorderLight   string
	Fg            string
	FgMuted       string
	FgDisabled    string
	Accent        string
	AccentHover   string
	AccentPressed string
	AccentFg      string
	BtnBg         string
	BtnBgHover    string
	BtnBgPressed  string
	BtnBorder     string
	BtnGradTop    string
	BtnGradBot    string
	Selection     string
	SelectionFg   string
	Shadow        string
}

var AquaLight = Theme{
	Font:          ".AppleSystemUIFont",
	Bg:            "#ececec",
	Surface:       "#ffffff",
	SurfaceAlt:    "#f6f6f6",
	Sidebar:       "#e8e4df",
	Toolbar:       "#e8e8e8",
	Border:        "#c4c4c4",
	BorderLight:   "#d6d6d6",
	Fg:            "#262626",
	FgMuted:       "#6e6e6e",
	FgDisabled:    "#a5a5a5",
	Accent:        "#007aff",
	AccentHover:   "#3395ff",
	AccentPressed: "#0062cc",
	AccentFg:      "#ffffff",
	BtnBg:         "#ffffff",
	BtnBgHover:    "#f0f0f0",
	BtnBgPressed:  "#d8d8d8",
	BtnBorder:     "#b6b6b6",
	BtnGradTop:    "#fefefe",
	BtnGradBot:    "#f0f0f0",
	Selection:     "#007aff",
	SelectionFg:   "#ffffff",
	Shadow:        "rgba(0, 0, 0, 30)",
}

var AquaDark = Theme{
	Font:          ".AppleSystemUIFont",
	Bg:            "#2d2d2d",
	Surface:       "#3a3a3a",
	SurfaceAlt:    "#323232",
	Sidebar:       "#2e2b28",
	Toolbar:       "#353535",
	Border:        "#4a4a4a",
	BorderLight:   "#505050",
	Fg:            "#e5e5e5",
	FgMuted:       "#9a9a9a",
	FgDisabled:    "#5c5c5c",
	Accent:        "#0a84ff",
	AccentHover:   "#409cff",
	AccentPressed: "#0060df",
	AccentFg:      "#ffffff",
	BtnBg:         "#4c4c4c",
	BtnBgHover:    "#555555",
	BtnBgPressed:  "#3c3c3c",
	BtnBorder:     "#5a5a5a",
	BtnGradTop:    "#555555",
	BtnGradBot:    "#464646",
	Selection:     "#0a84ff",
	SelectionFg:   "#ffffff",
	Shadow:        "rgba(0, 0, 0, 60)",
}

func (t Theme) QSS() string {
	r := strings.NewReplacer(
		"{{font}}", t.Font,
		"{{bg}}", t.Bg,
		"{{surface}}", t.Surface,
		"{{surface_alt}}", t.SurfaceAlt,
		"{{sidebar}}", t.Sidebar,
		"{{toolbar}}", t.Toolbar,
		"{{border}}", t.Border,
		"{{border_light}}", t.BorderLight,
		"{{fg}}", t.Fg,
		"{{fg_muted}}", t.FgMuted,
		"{{fg_disabled}}", t.FgDisabled,
		"{{accent}}", t.Accent,
		"{{accent_hover}}", t.AccentHover,
		"{{accent_pressed}}", t.AccentPressed,
		"{{accent_fg}}", t.AccentFg,
		"{{btn_bg}}", t.BtnBg,
		"{{btn_bg_hover}}", t.BtnBgHover,
		"{{btn_bg_pressed}}", t.BtnBgPressed,
		"{{btn_border}}", t.BtnBorder,
		"{{btn_grad_top}}", t.BtnGradTop,
		"{{btn_grad_bot}}", t.BtnGradBot,
		"{{selection}}", t.Selection,
		"{{selection_fg}}", t.SelectionFg,
		"{{shadow}}", t.Shadow,
	)
	return r.Replace(styleTemplate)
}

