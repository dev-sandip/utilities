package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type generateResultMsg struct {
	output     string
	err        error
	artifactID string
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Tab    key.Binding
	Shift  key.Binding
	Submit key.Binding
	Quit   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Tab, k.Submit, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab, k.Shift},
		{k.Submit, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next"),
	),
	Shift: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev"),
	),
	Submit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "generate"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type model struct {
	inputs       []textinput.Model
	focusIndex   int
	width        int
	height       int
	help         help.Model
	keys         keyMap
	spinner      spinner.Model
	loading      bool
	done         bool
	err          error
	output       string
	javaDetected string
	autoPackage  bool
}

const (
	fieldGroupID = iota
	fieldArtifactID
	fieldPackage
	fieldJava
)

func detectJavaVersion() string {
	cmd := exec.Command("java", "-version")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "17"
	}

	// Examples:
	// openjdk version "21.0.2"
	// java version "1.8.0_381"
	re := regexp.MustCompile(`"(\d+)(?:\.(\d+))?`)
	match := re.FindStringSubmatch(out.String())
	if len(match) >= 2 {
		if match[1] == "1" && len(match) >= 3 && match[2] != "" {
			return match[2]
		}
		return match[1]
	}

	return "17"
}

func sanitizePackagePart(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))

	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}

	result := b.String()
	if result == "" {
		return "app"
	}

	// Java package part should not start with a digit
	if result[0] >= '0' && result[0] <= '9' {
		result = "p" + result
	}

	return result
}

func sanitizeGroupID(groupID string) string {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(groupID)), ".")
	cleaned := make([]string, 0, len(parts))

	for _, p := range parts {
		p = sanitizePackagePart(p)
		if p != "" {
			cleaned = append(cleaned, p)
		}
	}

	if len(cleaned) == 0 {
		return "com.example"
	}

	return strings.Join(cleaned, ".")
}

func generatePackage(groupID, artifactID string) string {
	groupID = sanitizeGroupID(groupID)
	artifactPart := sanitizePackagePart(artifactID)
	return groupID + "." + artifactPart
}

func patchPomJavaVersion(path, javaVer string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(data)

	if strings.Contains(content, "<maven.compiler.source>") &&
		strings.Contains(content, "<maven.compiler.target>") {
		reSource := regexp.MustCompile(`<maven\.compiler\.source>.*?</maven\.compiler\.source>`)
		reTarget := regexp.MustCompile(`<maven\.compiler\.target>.*?</maven\.compiler\.target>`)
		content = reSource.ReplaceAllString(content, "<maven.compiler.source>"+javaVer+"</maven.compiler.source>")
		content = reTarget.ReplaceAllString(content, "<maven.compiler.target>"+javaVer+"</maven.compiler.target>")
		return os.WriteFile(path, []byte(content), 0644)
	}

	props := fmt.Sprintf(`
  <properties>
    <maven.compiler.source>%s</maven.compiler.source>
    <maven.compiler.target>%s</maven.compiler.target>
  </properties>`, javaVer, javaVer)

	content = strings.Replace(content, "</project>", props+"\n</project>", 1)
	return os.WriteFile(path, []byte(content), 0644)
}

func generateProject(groupID, artifactID, pkg, javaVer string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(
			"mvn",
			"archetype:generate",
			"-DgroupId="+groupID,
			"-DartifactId="+artifactID,
			"-Dpackage="+pkg,
			"-DarchetypeArtifactId=maven-archetype-quickstart",
			"-DinteractiveMode=false",
		)

		out, err := cmd.CombinedOutput()
		if err != nil {
			return generateResultMsg{
				output:     string(out),
				err:        err,
				artifactID: artifactID,
			}
		}

		pomPath := artifactID + "/pom.xml"
		_ = patchPomJavaVersion(pomPath, javaVer)

		return generateResultMsg{
			output:     string(out),
			err:        nil,
			artifactID: artifactID,
		}
	}
}

func initialModel() model {
	javaVer := detectJavaVersion()
	groupID := "com.example"
	artifactID := "demo-app"
	packageName := generatePackage(groupID, artifactID)

	inputs := make([]textinput.Model, 4)

	inputs[fieldGroupID] = textinput.New()
	inputs[fieldGroupID].Prompt = ""
	inputs[fieldGroupID].SetValue(groupID)
	inputs[fieldGroupID].Placeholder = "com.example"
	inputs[fieldGroupID].Width = 38
	inputs[fieldGroupID].Focus()

	inputs[fieldArtifactID] = textinput.New()
	inputs[fieldArtifactID].Prompt = ""
	inputs[fieldArtifactID].SetValue(artifactID)
	inputs[fieldArtifactID].Placeholder = "demo-app"
	inputs[fieldArtifactID].Width = 38

	inputs[fieldPackage] = textinput.New()
	inputs[fieldPackage].Prompt = ""
	inputs[fieldPackage].SetValue(packageName)
	inputs[fieldPackage].Placeholder = "com.example.demoapp"
	inputs[fieldPackage].Width = 38

	inputs[fieldJava] = textinput.New()
	inputs[fieldJava].Prompt = ""
	inputs[fieldJava].SetValue(javaVer)
	inputs[fieldJava].Placeholder = "17"
	inputs[fieldJava].Width = 38

	s := spinner.New()
	s.Spinner = spinner.Dot

	return model{
		inputs:       inputs,
		help:         help.New(),
		keys:         keys,
		spinner:      s,
		javaDetected: javaVer,
		autoPackage:  true,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i == m.focusIndex {
			cmds[i] = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func isValidJavaPackage(pkg string) bool {
	if pkg == "" {
		return false
	}

	parts := strings.Split(pkg, ".")
	if len(parts) == 0 {
		return false
	}

	for _, p := range parts {
		if p == "" {
			return false
		}

		for i, r := range p {
			if i == 0 && r >= '0' && r <= '9' {
				return false
			}
			if !((r >= 'a' && r <= 'z') ||
				(r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') ||
				r == '_') {
				return false
			}
		}
	}

	return true
}

func isValidArtifactID(s string) bool {
	s = strings.TrimSpace(s)
	return s != "" && !strings.Contains(s, " ")
}

func isValidJavaVersion(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case generateResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = fmt.Errorf("maven failed\n\n%s", strings.TrimSpace(msg.output))
			return m, nil
		}
		m.done = true
		m.output = msg.output
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}

		if m.done {
			switch msg.String() {
			case "enter", "q", "ctrl+c":
				return m, tea.Quit
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "shift+tab":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			return m, m.updateFocus()

		case "down", "tab":
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			return m, m.updateFocus()

		case "enter":
			groupID := strings.TrimSpace(m.inputs[fieldGroupID].Value())
			artifactID := strings.TrimSpace(m.inputs[fieldArtifactID].Value())
			pkg := strings.TrimSpace(m.inputs[fieldPackage].Value())
			javaVer := strings.TrimSpace(m.inputs[fieldJava].Value())

			if groupID == "" || artifactID == "" || pkg == "" || javaVer == "" {
				m.err = fmt.Errorf("all fields are required")
				return m, nil
			}
			if !isValidArtifactID(artifactID) {
				m.err = fmt.Errorf("artifact id cannot be empty or contain spaces")
				return m, nil
			}
			if !isValidJavaPackage(pkg) {
				m.err = fmt.Errorf("invalid java package name")
				return m, nil
			}
			if !isValidJavaVersion(javaVer) {
				m.err = fmt.Errorf("java version must be numeric")
				return m, nil
			}

			m.err = nil
			m.loading = true
			return m, tea.Batch(
				m.spinner.Tick,
				generateProject(groupID, artifactID, pkg, javaVer),
			)
		}
	}

	if !m.loading && !m.done {
		cmds := make([]tea.Cmd, len(m.inputs))

		for i := range m.inputs {
			prev := m.inputs[i].Value()
			var cmd tea.Cmd
			m.inputs[i], cmd = m.inputs[i].Update(msg)
			cmds[i] = cmd

			// If user manually edits package field, disable auto mode.
			if i == fieldPackage && m.inputs[i].Value() != prev {
				m.autoPackage = false
			}
		}

		// Re-enable auto package if package field is cleared.
		if strings.TrimSpace(m.inputs[fieldPackage].Value()) == "" {
			m.autoPackage = true
		}

		// Auto-generate package from groupId + artifactId until user edits package manually.
		if m.autoPackage {
			groupID := strings.TrimSpace(m.inputs[fieldGroupID].Value())
			artifactID := strings.TrimSpace(m.inputs[fieldArtifactID].Value())
			m.inputs[fieldPackage].SetValue(generatePackage(groupID, artifactID))
		}

		return m, tea.Batch(cmds...)
	}

	return m, nil
}

var (
	bgStyle = lipgloss.NewStyle().
		Padding(1, 2)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Width(72)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Width(16).
			Foreground(lipgloss.Color("111"))

	activeLabelStyle = lipgloss.NewStyle().
				Width(16).
				Bold(true).
				Foreground(lipgloss.Color("205"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	focusedInputBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("205")).
				Padding(0, 1)
)

func (m model) View() string {
	var body strings.Builder

	body.WriteString(titleStyle.Render("Java Project Generator"))
	body.WriteString("\n")
	body.WriteString(subtitleStyle.Render("Bubble Tea TUI using installed Maven and detected Java version."))
	body.WriteString("\n")

	if m.done {
		body.WriteString(successStyle.Render("Project created successfully"))
		body.WriteString("\n\n")
		body.WriteString(infoStyle.Render("Folder: " + m.inputs[fieldArtifactID].Value()))
		body.WriteString("\n")
		body.WriteString(mutedStyle.Render("Next: cd "+m.inputs[fieldArtifactID].Value()+" && mvn package"))
		body.WriteString("\n\n")
		body.WriteString("Press Enter or q to quit.")
		return m.center(cardStyle.Render(body.String()))
	}

	if m.loading {
		body.WriteString(fmt.Sprintf("%s Generating project with Maven...", m.spinner.View()))
		body.WriteString("\n\n")
		body.WriteString(infoStyle.Render("This uses mvn archetype:generate"))
		return m.center(cardStyle.Render(body.String()))
	}

	body.WriteString(infoStyle.Render("Detected Java: " + m.javaDetected + "  (editable)"))
	body.WriteString("\n")
	if m.autoPackage {
		body.WriteString(mutedStyle.Render("Package auto-generated from artifact id"))
	} else {
		body.WriteString(mutedStyle.Render("Package is in manual mode"))
	}
	body.WriteString("\n\n")

	labels := []string{
		"Group ID",
		"Artifact ID",
		"Package",
		"Java Version",
	}

	for i := range m.inputs {
		label := labelStyle.Render(labels[i])
		box := inputBoxStyle.Render(m.inputs[i].View())

		if i == m.focusIndex {
			label = activeLabelStyle.Render("→ " + labels[i])
			box = focusedInputBoxStyle.Render(m.inputs[i].View())
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top, label, box)
		body.WriteString(row)
		body.WriteString("\n\n")
	}

	body.WriteString(mutedStyle.Render("artifact: my-app  →  package: com.example.myapp"))
	body.WriteString("\n")
	body.WriteString(mutedStyle.Render("Clear package field to enable auto mode again."))
	body.WriteString("\n\n")

	if m.err != nil {
		body.WriteString(errorStyle.Render(m.err.Error()))
		body.WriteString("\n\n")
	}

	body.WriteString(m.help.View(m.keys))

	return m.center(bgStyle.Render(cardStyle.Render(body.String())))
}

func (m model) center(content string) string {
	if m.width == 0 || m.height == 0 {
		return "\n" + content + "\n"
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
