package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vibe-mvp/internal/manifest"
)

// ── modes ─────────────────────────────────────────────────────────────────────

type beMode int

const (
	beNormal beMode = iota
	beInsert
)

// ── arch options ──────────────────────────────────────────────────────────────

type archOption struct {
	value string
	label string
	desc  string
}

var beArchOptions = []archOption{
	{"monolith", "Monolith", "Single deployable unit — all features in one codebase"},
	{"modular-monolith", "Modular Monolith", "Clear domain boundaries, single deployment"},
	{"microservices", "Microservices", "Independent services communicating over a network"},
	{"event-driven", "Event-Driven", "Services communicate asynchronously via events"},
}

// ── field helpers ─────────────────────────────────────────────────────────────

func defaultBeEnvFields() []Field {
	return []Field{
		{
			Key: "compute_env", Label: "compute_env   ", Kind: KindSelect,
			Options: []string{"serverless", "containerized", "bare-metal/VM"},
			Value:   "containerized", SelIdx: 1,
		},
		{
			Key: "cloud_provider", Label: "cloud_provider", Kind: KindSelect,
			Options: []string{"N/A", "AWS", "GCP", "Azure"},
			Value:   "N/A",
		},
	}
}

func defaultBeAppFields() []Field {
	return []Field{
		{Key: "language", Label: "language      ", Kind: KindText},
		{Key: "framework", Label: "framework     ", Kind: KindText},
	}
}

func defaultSvcFields() []Field {
	return []Field{
		{Key: "name", Label: "name          ", Kind: KindText},
		{Key: "responsibility", Label: "responsibility", Kind: KindText},
		{Key: "language", Label: "language      ", Kind: KindText},
		{Key: "framework", Label: "framework     ", Kind: KindText},
	}
}

func svcFieldsFromDef(s manifest.ServiceDef) []Field {
	f := defaultSvcFields()
	vals := map[string]string{
		"name": s.Name, "responsibility": s.Responsibility,
		"language": s.Language, "framework": s.Framework,
	}
	for i := range f {
		f[i].Value = vals[f[i].Key]
	}
	return f
}

// ── BackendEditor ─────────────────────────────────────────────────────────────

// BackendEditor manages the BACK section.
//
// Phase 1 — arch selection: a collapsed dropdown row shows the current choice;
// pressing Enter/Space opens the dropdown list; j/k navigates; Enter confirms.
//
// Phase 2 — sub-tabs: ENV (always) + APP (monolith) or per-service tabs
// (microservices / modular-monolith). Press b to return to Phase 1.
type BackendEditor struct {
	// Persisted data
	ArchIdx       int
	ArchConfirmed bool
	EnvFields     []Field
	AppFields     []Field
	Services      []manifest.ServiceDef

	// Phase-1 dropdown state
	dropdownOpen bool // true while the arch dropdown list is visible
	dropdownIdx  int  // cursor position inside the open dropdown

	// Phase-2 UI state
	activeSubTab int
	activeField  int

	internalMode beMode
	formInput    textinput.Model
	width        int
}

func newBackendEditor() BackendEditor {
	fi := textinput.New()
	fi.Prompt = ""
	fi.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrFg))
	fi.CursorStyle = StyleCursor
	fi.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(clrFgDim))

	return BackendEditor{
		EnvFields: defaultBeEnvFields(),
		AppFields: defaultBeAppFields(),
		Services:  []manifest.ServiceDef{},
		formInput: fi,
	}
}

// ToManifest converts the editor state to a BackendPillar.
func (be BackendEditor) ToManifest() manifest.BackendPillar {
	envGet := func(key string) string {
		for _, f := range be.EnvFields {
			if f.Key == key {
				return f.DisplayValue()
			}
		}
		return ""
	}
	appGet := func(key string) string {
		for _, f := range be.AppFields {
			if f.Key == key {
				return f.Value
			}
		}
		return ""
	}
	return manifest.BackendPillar{
		ArchPattern:   manifest.ArchPattern(be.currentArch()),
		ComputeEnv:    manifest.ComputeEnv(envGet("compute_env")),
		CloudProvider: envGet("cloud_provider"),
		Language:      appGet("language"),
		Framework:     appGet("framework"),
		Services:      be.Services,
	}
}

// ── arch helpers ──────────────────────────────────────────────────────────────

func (be BackendEditor) currentArch() string {
	if be.ArchIdx >= 0 && be.ArchIdx < len(beArchOptions) {
		return beArchOptions[be.ArchIdx].value
	}
	return beArchOptions[0].value
}

func (be BackendEditor) isServiceArch() bool {
	a := be.currentArch()
	return a == "microservices" || a == "modular-monolith"
}

// ── sub-tab helpers ───────────────────────────────────────────────────────────

func (be BackendEditor) subTabCount() int {
	switch be.currentArch() {
	case "monolith":
		return 2 // ENV + APP
	case "event-driven":
		return 1 // ENV only
	default:
		return 1 + len(be.Services) + 1 // ENV + services + [+]
	}
}

func (be BackendEditor) subTabLabel(i int) string {
	switch be.currentArch() {
	case "monolith":
		if i == 0 {
			return "ENV"
		}
		return "APP"
	case "event-driven":
		return "ENV"
	default:
		if i == 0 {
			return "ENV"
		}
		svcIdx := i - 1
		if svcIdx < len(be.Services) {
			name := be.Services[svcIdx].Name
			if name == "" {
				name = fmt.Sprintf("#%d", svcIdx+1)
			}
			return name
		}
		return "[+]"
	}
}

func (be BackendEditor) isAddTab() bool {
	return be.isServiceArch() && be.activeSubTab == len(be.Services)+1
}

// ── Mode / HintLine ───────────────────────────────────────────────────────────

func (be BackendEditor) Mode() Mode {
	if be.internalMode == beInsert {
		return ModeInsert
	}
	return ModeNormal
}

func (be BackendEditor) HintLine() string {
	if !be.ArchConfirmed {
		if be.dropdownOpen {
			hints := []string{
				StyleHelpKey.Render("j/k") + StyleHelpDesc.Render(" navigate"),
				StyleHelpKey.Render("Enter") + StyleHelpDesc.Render(" confirm"),
				StyleHelpKey.Render("Esc") + StyleHelpDesc.Render(" close"),
			}
			return "  " + strings.Join(hints, StyleHelpDesc.Render("  ·  "))
		}
		hints := []string{
			StyleHelpKey.Render("Enter") + StyleHelpDesc.Render(" open"),
		}
		return "  " + strings.Join(hints, StyleHelpDesc.Render("  ·  "))
	}
	if be.internalMode == beInsert {
		return StyleInsertMode.Render(" -- INSERT -- ") +
			StyleHelpDesc.Render("  Esc: normal  Tab: next field")
	}
	hints := []string{
		StyleHelpKey.Render("j/k") + StyleHelpDesc.Render(" navigate"),
		StyleHelpKey.Render("h/l") + StyleHelpDesc.Render(" sub-tab"),
		StyleHelpKey.Render("b") + StyleHelpDesc.Render(" change arch"),
		StyleHelpKey.Render("i/Enter") + StyleHelpDesc.Render(" edit"),
		StyleHelpKey.Render("Space") + StyleHelpDesc.Render(" cycle"),
	}
	if be.isServiceArch() {
		hints = append(hints, StyleHelpKey.Render("a")+StyleHelpDesc.Render(" add"))
		if len(be.Services) > 0 && be.activeSubTab > 0 && !be.isAddTab() {
			hints = append(hints, StyleHelpKey.Render("d")+StyleHelpDesc.Render(" delete"))
		}
	}
	return "  " + strings.Join(hints, StyleHelpDesc.Render("  ·  "))
}

// ── Update ────────────────────────────────────────────────────────────────────

func (be BackendEditor) Update(msg tea.Msg) (BackendEditor, tea.Cmd) {
	if !be.ArchConfirmed {
		return be.updateArchSelect(msg)
	}
	if be.internalMode == beInsert {
		return be.updateInsert(msg)
	}
	return be.updateNormal(msg)
}

// updateArchSelect handles Phase-1 input (dropdown closed or open).
func (be BackendEditor) updateArchSelect(msg tea.Msg) (BackendEditor, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return be, nil
	}

	if be.dropdownOpen {
		switch key.String() {
		case "j", "down":
			if be.dropdownIdx < len(beArchOptions)-1 {
				be.dropdownIdx++
			}
		case "k", "up":
			if be.dropdownIdx > 0 {
				be.dropdownIdx--
			}
		case "g":
			be.dropdownIdx = 0
		case "G":
			be.dropdownIdx = len(beArchOptions) - 1
		case "enter", " ":
			be.ArchIdx = be.dropdownIdx
			be.dropdownOpen = false
			be.ArchConfirmed = true
			be.activeSubTab = 0
			be.activeField = 0
		case "esc":
			be.dropdownOpen = false
		}
		return be, nil
	}

	// Dropdown is closed.
	switch key.String() {
	case "enter", " ":
		be.dropdownOpen = true
		be.dropdownIdx = be.ArchIdx // pre-select the current choice
	}
	return be, nil
}

func (be BackendEditor) updateInsert(msg tea.Msg) (BackendEditor, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if ok {
		switch key.String() {
		case "esc":
			be.saveTextInput()
			be.internalMode = beNormal
			be.formInput.Blur()
			return be, nil
		case "tab":
			be.saveTextInput()
			fields := be.currentFields()
			be.activeField = (be.activeField + 1) % len(fields)
			return be.tryEnterInsert()
		case "shift+tab":
			be.saveTextInput()
			fields := be.currentFields()
			n := len(fields)
			be.activeField = (be.activeField - 1 + n) % n
			return be.tryEnterInsert()
		}
	}
	var cmd tea.Cmd
	be.formInput, cmd = be.formInput.Update(msg)
	return be, cmd
}

func (be BackendEditor) updateNormal(msg tea.Msg) (BackendEditor, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return be, nil
	}

	switch key.String() {
	// ── Back to arch selection ────────────────────────────────────────────
	case "b":
		be.ArchConfirmed = false
		be.dropdownOpen = false
		be.dropdownIdx = be.ArchIdx
		be.activeSubTab = 0
		be.activeField = 0

	// ── Sub-tab navigation ────────────────────────────────────────────────
	case "h", "left":
		if be.activeSubTab > 0 {
			be.activeSubTab--
			be.activeField = 0
		}
	case "l", "right":
		if be.activeSubTab < be.subTabCount()-1 {
			be.activeSubTab++
			be.activeField = 0
		}

	// ── Field navigation ──────────────────────────────────────────────────
	case "j", "down":
		if fields := be.currentFields(); fields != nil && be.activeField < len(fields)-1 {
			be.activeField++
		}
	case "k", "up":
		if be.activeField > 0 {
			be.activeField--
		}
	case "g":
		be.activeField = 0
	case "G":
		if fields := be.currentFields(); fields != nil {
			be.activeField = len(fields) - 1
		}

	// ── Select cycling / text edit ────────────────────────────────────────
	case "enter", " ":
		if be.isAddTab() {
			be.addService()
			return be, nil
		}
		if f := be.mutableField(); f != nil && f.Kind == KindSelect {
			f.CycleNext()
		} else {
			return be.tryEnterInsert()
		}

	case "H", "shift+left":
		if f := be.mutableField(); f != nil && f.Kind == KindSelect {
			f.CyclePrev()
		}

	// ── Insert mode ───────────────────────────────────────────────────────
	case "i":
		if !be.isAddTab() {
			return be.tryEnterInsert()
		}

	// ── Service add / delete ──────────────────────────────────────────────
	case "a":
		if be.isServiceArch() {
			be.addService()
			return be, nil
		}
	case "d":
		if be.isServiceArch() && be.activeSubTab > 0 && !be.isAddTab() {
			svcIdx := be.activeSubTab - 1
			be.Services = append(be.Services[:svcIdx], be.Services[svcIdx+1:]...)
			if be.activeSubTab >= be.subTabCount() {
				be.activeSubTab = be.subTabCount() - 1
			}
			be.activeField = 0
		}
	}

	return be, nil
}

func (be *BackendEditor) addService() {
	be.Services = append(be.Services, manifest.ServiceDef{})
	be.activeSubTab = len(be.Services)
	be.activeField = 0
}

func (be BackendEditor) tryEnterInsert() (BackendEditor, tea.Cmd) {
	fields := be.currentFields()
	if fields == nil || be.activeField >= len(fields) {
		return be, nil
	}
	f := fields[be.activeField]
	if f.Kind != KindText {
		return be, nil
	}
	be.internalMode = beInsert
	be.formInput.SetValue(f.Value)
	be.formInput.Width = be.width - 22
	be.formInput.CursorEnd()
	return be, be.formInput.Focus()
}

// currentFields returns the display fields for the active sub-tab (read-only copy).
func (be BackendEditor) currentFields() []Field {
	if be.isAddTab() {
		return nil
	}
	arch := be.currentArch()
	switch {
	case arch == "event-driven" || be.activeSubTab == 0:
		return be.EnvFields
	case arch == "monolith":
		return be.AppFields
	default:
		svcIdx := be.activeSubTab - 1
		if svcIdx >= 0 && svcIdx < len(be.Services) {
			return svcFieldsFromDef(be.Services[svcIdx])
		}
		return nil
	}
}

// mutableField returns a pointer to the active field for select cycling.
// Returns nil for service tabs (they only have text fields).
func (be *BackendEditor) mutableField() *Field {
	if be.isAddTab() {
		return nil
	}
	arch := be.currentArch()
	switch {
	case arch == "event-driven" || be.activeSubTab == 0:
		if be.activeField < len(be.EnvFields) {
			return &be.EnvFields[be.activeField]
		}
	case arch == "monolith":
		if be.activeField < len(be.AppFields) {
			return &be.AppFields[be.activeField]
		}
	}
	return nil
}

// saveTextInput writes the current formInput value back to the right store.
func (be *BackendEditor) saveTextInput() {
	val := be.formInput.Value()
	arch := be.currentArch()

	if arch == "event-driven" || be.activeSubTab == 0 {
		if be.activeField < len(be.EnvFields) && be.EnvFields[be.activeField].Kind == KindText {
			be.EnvFields[be.activeField].Value = val
		}
		return
	}
	if arch == "monolith" {
		if be.activeField < len(be.AppFields) {
			be.AppFields[be.activeField].Value = val
		}
		return
	}
	svcIdx := be.activeSubTab - 1
	if svcIdx < 0 || svcIdx >= len(be.Services) {
		return
	}
	keys := []string{"name", "responsibility", "language", "framework"}
	if be.activeField >= len(keys) {
		return
	}
	switch keys[be.activeField] {
	case "name":
		be.Services[svcIdx].Name = val
	case "responsibility":
		be.Services[svcIdx].Responsibility = val
	case "language":
		be.Services[svcIdx].Language = val
	case "framework":
		be.Services[svcIdx].Framework = val
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func (be BackendEditor) View(w, h int) string {
	be.width = w
	if !be.ArchConfirmed {
		return be.viewArchSelect(w, h)
	}
	return be.viewSubTabs(w, h)
}

// viewArchSelect renders Phase 1: a collapsed select row + optional dropdown.
func (be BackendEditor) viewArchSelect(w, h int) string {
	var lines []string

	lines = append(lines,
		StyleSectionDesc.Render("  # Phase 2 · Backend — Choose an architecture"),
		"",
	)

	// ── Select row ────────────────────────────────────────────────────────
	current := beArchOptions[be.ArchIdx]
	label := StyleFieldKey.Render("arch_pattern  ")
	val := StyleFieldValActive.Render(current.label) + StyleSelectArrow.Render(" ▾")
	row := "     " + label + StyleEquals.Render(" = ") + val
	raw := lipgloss.Width(row)
	if raw < w {
		row += strings.Repeat(" ", w-raw)
	}
	lines = append(lines, StyleCurLine.Render(row))

	// ── Dropdown list (when open) ─────────────────────────────────────────
	if be.dropdownOpen {
		lines = append(lines, "")
		for i, opt := range beArchOptions {
			isCur := i == be.dropdownIdx

			var cursor string
			if isCur {
				cursor = StyleCurLineNum.Render("  ▶ ")
			} else {
				cursor = "      "
			}

			var optRow string
			labelPart := fmt.Sprintf("%-18s", opt.label)
			if isCur {
				optRow = cursor +
					StyleFieldValActive.Render(labelPart) +
					StyleSectionDesc.Render(opt.desc)
				raw := lipgloss.Width(optRow)
				if raw < w {
					optRow += strings.Repeat(" ", w-raw)
				}
				optRow = StyleCurLine.Render(optRow)
			} else {
				optRow = cursor +
					StyleFieldKey.Render(labelPart) +
					StyleSectionDesc.Render(opt.desc)
			}
			lines = append(lines, optRow)
		}
	}

	for len(lines) < h {
		lines = append(lines, StyleTilde.Render("~"))
	}
	if len(lines) > h {
		lines = lines[:h]
	}
	return strings.Join(lines, "\n") + "\n"
}

// viewSubTabs renders Phase 2: confirmed arch header + sub-tab content.
func (be BackendEditor) viewSubTabs(w, h int) string {
	var lines []string

	// ── Confirmed arch breadcrumb ─────────────────────────────────────────
	opt := beArchOptions[be.ArchIdx]
	archStr := StyleFieldValActive.Render(opt.label)
	hint := StyleSectionDesc.Render("  (b: change)")
	lines = append(lines,
		StyleSectionDesc.Render("  # Backend · ")+archStr+hint,
		"",
	)

	// ── Sub-tab bar ───────────────────────────────────────────────────────
	var tabParts []string
	for i := 0; i < be.subTabCount(); i++ {
		lbl := be.subTabLabel(i)
		if i == be.activeSubTab {
			tabParts = append(tabParts, StyleTabActive.Render(" "+lbl+" "))
		} else {
			tabParts = append(tabParts, StyleTabInactive.Render(" "+lbl+" "))
		}
	}
	lines = append(lines, "  "+strings.Join(tabParts, ""), "")

	// ── Sub-tab content ───────────────────────────────────────────────────
	if be.isAddTab() {
		unitLabel := "service"
		if be.currentArch() == "modular-monolith" {
			unitLabel = "module"
		}
		lines = append(lines,
			StyleSectionDesc.Render(fmt.Sprintf(
				"  Press Enter or 'a' to add a new %s", unitLabel)),
		)
	} else {
		lines = append(lines, be.renderFields(w, be.currentFields())...)
	}

	for len(lines) < h {
		lines = append(lines, StyleTilde.Render("~"))
	}
	if len(lines) > h {
		lines = lines[:h]
	}
	return strings.Join(lines, "\n") + "\n"
}

func (be BackendEditor) renderFields(w int, fields []Field) []string {
	if fields == nil {
		return nil
	}
	const labelW = 14
	const eqW = 3
	valW := w - 4 - labelW - eqW - 1
	if valW < 10 {
		valW = 10
	}

	var lines []string
	for i, f := range fields {
		isCur := i == be.activeField

		lineNo := StyleLineNum.Render(fmt.Sprintf("%3d ", i+1))
		if isCur {
			lineNo = StyleCurLineNum.Render(fmt.Sprintf("%3d ", i+1))
		}

		var keyStr string
		if isCur {
			keyStr = StyleFieldKeyActive.Render(f.Label)
		} else {
			keyStr = StyleFieldKey.Render(f.Label)
		}

		eq := StyleEquals.Render(" = ")

		var valStr string
		switch {
		case be.internalMode == beInsert && isCur && f.Kind == KindText:
			valStr = be.formInput.View()
		case f.Kind == KindSelect:
			val := f.DisplayValue()
			if isCur {
				val = StyleFieldValActive.Render(val)
			} else {
				val = StyleFieldVal.Render(val)
			}
			valStr = val + StyleSelectArrow.Render(" ▾")
		default:
			dv := f.Value
			if len(dv) > valW {
				dv = dv[:valW-1] + "…"
			}
			if dv == "" {
				valStr = StyleSectionDesc.Render("_")
			} else if isCur {
				valStr = StyleFieldValActive.Render(dv)
			} else {
				valStr = StyleFieldVal.Render(dv)
			}
		}

		row := lineNo + keyStr + eq + valStr
		if isCur {
			raw := lipgloss.Width(row)
			if raw < w {
				row += strings.Repeat(" ", w-raw)
			}
			row = StyleCurLine.Render(row)
		}
		lines = append(lines, row)
	}
	return lines
}
