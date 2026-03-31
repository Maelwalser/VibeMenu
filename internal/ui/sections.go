package ui

// FieldKind enumerates the types of form fields.
type FieldKind int

const (
	KindText     FieldKind = iota // single-line text input
	KindSelect                   // cycle through a list of options
	KindTextArea                 // multi-line text input
)

// Field represents a single form field within a section.
type Field struct {
	Key     string    // machine key (e.g. "target_env")
	Label   string    // padded display label
	Kind    FieldKind
	Value   string    // current string value
	Options []string  // KindSelect: available choices
	SelIdx  int       // KindSelect: currently selected index
}

// DisplayValue returns the rendered value string for NORMAL mode.
func (f Field) DisplayValue() string {
	if f.Kind == KindSelect {
		if len(f.Options) == 0 {
			return ""
		}
		return f.Options[f.SelIdx]
	}
	// Truncate long single-line previews for textarea
	v := f.Value
	if f.Kind == KindTextArea && len(v) > 0 {
		lines := splitLines(v)
		if len(lines) > 1 {
			return lines[0] + " …"
		}
	}
	return v
}

// CycleNext advances a KindSelect field to the next option.
func (f *Field) CycleNext() {
	if f.Kind != KindSelect || len(f.Options) == 0 {
		return
	}
	f.SelIdx = (f.SelIdx + 1) % len(f.Options)
	f.Value = f.Options[f.SelIdx]
}

// CyclePrev moves a KindSelect field to the previous option.
func (f *Field) CyclePrev() {
	if f.Kind != KindSelect || len(f.Options) == 0 {
		return
	}
	f.SelIdx = (f.SelIdx - 1 + len(f.Options)) % len(f.Options)
	f.Value = f.Options[f.SelIdx]
}

// Section groups related fields under a pillar.
type Section struct {
	ID     string  // short identifier (e.g. "arch")
	Abbr   string  // 3-char tab label
	Title  string  // full title
	Desc   string  // one-line description
	Fields []Field
}

// initSections returns the six-pillar section definitions.
func initSections() []Section {
	return []Section{
		{
			ID:   "arch",
			Abbr: "ARCH",
			Title: "System Architecture & Deployment",
			Desc:  "Where and how will this application live?",
			Fields: []Field{
				{
					Key:     "target_env",
					Label:   "target_env    ",
					Kind:    KindSelect,
					Options: []string{"cloud-native", "on-premise", "edge", "local"},
				},
				{
					Key:     "cloud_provider",
					Label:   "cloud_provider",
					Kind:    KindSelect,
					Options: []string{"AWS", "GCP", "Azure", "N/A"},
					SelIdx:  3,
					Value:   "N/A",
				},
				{
					Key:     "topology",
					Label:   "topology      ",
					Kind:    KindSelect,
					Options: []string{"monolith", "microservices", "serverless", "event-driven"},
				},
				{
					Key:     "scaling",
					Label:   "scaling       ",
					Kind:    KindSelect,
					Options: []string{"horizontal", "vertical", "both", "none"},
					SelIdx:  3,
					Value:   "none",
				},
				{
					Key:   "scaling_notes",
					Label: "scaling_notes ",
					Kind:  KindText,
				},
			},
		},
		{
			ID:   "stack",
			Abbr: "STACK",
			Title: "Technology Stack & Boundaries",
			Desc:  "Languages, frameworks, runtimes, and third-party integrations.",
			Fields: []Field{
				{Key: "fe_framework",  Label: "fe_framework  ", Kind: KindText},
				{Key: "fe_version",    Label: "fe_version    ", Kind: KindText},
				{Key: "state_mgmt",    Label: "state_mgmt    ", Kind: KindText},
				{Key: "styling",       Label: "styling       ", Kind: KindText},
				{Key: "be_language",   Label: "be_language   ", Kind: KindText},
				{Key: "be_framework",  Label: "be_framework  ", Kind: KindText},
				{Key: "runtime",       Label: "runtime       ", Kind: KindText},
				{Key: "integrations",  Label: "integrations  ", Kind: KindTextArea},
			},
		},
		{
			ID:   "data",
			Abbr: "DATA",
			Title: "Data Architecture",
			Desc:  "Database schemas, relationships, and API paradigms.",
			Fields: []Field{
				{
					Key:     "db_type",
					Label:   "db_type       ",
					Kind:    KindSelect,
					Options: []string{"PostgreSQL", "MySQL", "MongoDB", "DynamoDB", "Redis", "SQLite", "other"},
				},
				{Key: "secondary_db",  Label: "secondary_db  ", Kind: KindText},
				{Key: "entities",      Label: "entities      ", Kind: KindTextArea},
				{
					Key:     "api_paradigm",
					Label:   "api_paradigm  ",
					Kind:    KindSelect,
					Options: []string{"REST", "GraphQL", "gRPC", "tRPC", "mixed"},
				},
				{Key: "caching",       Label: "caching       ", Kind: KindText},
			},
		},
		{
			ID:   "features",
			Abbr: "FEAT",
			Title: "Functional Specifications",
			Desc:  "User roles, core journeys, and error handling strategies.",
			Fields: []Field{
				{Key: "user_roles",    Label: "user_roles    ", Kind: KindTextArea},
				{Key: "core_journeys", Label: "core_journeys ", Kind: KindTextArea},
				{Key: "error_handling",Label: "error_handling", Kind: KindTextArea},
			},
		},
		{
			ID:   "nfr",
			Abbr: "NFR",
			Title: "Non-Functional Requirements",
			Desc:  "Security, performance, compliance, and accessibility.",
			Fields: []Field{
				{
					Key:     "encryption",
					Label:   "encryption    ",
					Kind:    KindSelect,
					Options: []string{"AES-256 + TLS 1.3", "TLS 1.3 only", "AES-256 only", "none"},
				},
				{
					Key:     "sanitization",
					Label:   "sanitization  ",
					Kind:    KindSelect,
					Options: []string{"yes", "no"},
				},
				{Key: "rate_limiting",  Label: "rate_limiting ", Kind: KindText},
				{Key: "latency_target", Label: "latency_target", Kind: KindText},
				{Key: "compliance",     Label: "compliance    ", Kind: KindText},
				{
					Key:     "accessibility",
					Label:   "accessibility ",
					Kind:    KindSelect,
					Options: []string{"WCAG 2.1 AA", "WCAG 2.1 AAA", "none"},
					SelIdx:  2,
					Value:   "none",
				},
			},
		},
		{
			ID:   "workflow",
			Abbr: "FLOW",
			Title: "Development Workflow & Tooling",
			Desc:  "Project structure, testing strategy, CI, and linting.",
			Fields: []Field{
				{
					Key:     "project_structure",
					Label:   "proj_structure",
					Kind:    KindSelect,
					Options: []string{"feature-based", "layer-based", "domain-driven", "standard Go"},
				},
				{Key: "test_framework",  Label: "test_framework", Kind: KindText},
				{Key: "coverage_target", Label: "coverage      ", Kind: KindText, Value: "80%"},
				{
					Key:     "ci_platform",
					Label:   "ci_platform   ",
					Kind:    KindSelect,
					Options: []string{"GitHub Actions", "GitLab CI", "CircleCI", "Jenkins", "none"},
					SelIdx:  4,
					Value:   "none",
				},
				{Key: "linting",     Label: "linting       ", Kind: KindText},
				{Key: "formatting",  Label: "formatting    ", Kind: KindText},
			},
		},
	}
}

// splitLines splits a string into lines.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}
