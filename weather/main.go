package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Models ---

type weatherData struct {
	Name     string `json:"name"`
	Timezone int    `json:"timezone"`
	Main     struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Humidity  int     `json:"humidity"`
		Pressure  int     `json:"pressure"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Visibility int `json:"visibility"`
	Sys        struct {
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	} `json:"sys"`
}

// --- Bubble Tea Model ---

type model struct {
	textInput textinput.Model
	weather   *weatherData
	err       error
	loading   bool
}

type weatherMsg weatherData
type errMsg error

// --- Init ---
func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter city name..."
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 30

	return model{
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// --- Update ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			city := strings.TrimSpace(m.textInput.Value())
			if city != "" {
				m.loading = true
				return m, fetchWeather(city)
			}
		}

	case weatherMsg:
		m.weather = (*weatherData)(&msg)
		m.loading = false
		return m, nil

	case errMsg:
		m.err = msg
		m.loading = false
		return m, nil
	}

	return m, cmd
}

// --- View ---

// --- View ---

func (m model) View() string {
	style := lipgloss.NewStyle().Padding(1, 2)

	if m.loading {
		return style.Render("Fetching weather data... ☁️")
	}

	if m.err != nil {
		return style.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.weather != nil {
		tempC := m.weather.Main.Temp - 273.15
		feelsLikeC := m.weather.Main.FeelsLike - 273.15
		tempMinC := m.weather.Main.TempMin - 273.15
		tempMaxC := m.weather.Main.TempMax - 273.15

		color := lipgloss.NewStyle()
		switch {
		case tempC < 10:
			color = color.Foreground(lipgloss.Color("#00BFFF"))
		case tempC < 25:
			color = color.Foreground(lipgloss.Color("#00FF7F"))
		default:
			color = color.Foreground(lipgloss.Color("#FF6347"))
		}

		// Convert sunrise/sunset from UNIX timestamp to local time
		sunrise := fmt.Sprintf("%v", unixToTime(m.weather.Sys.Sunrise+m.weather.Timezone))
		sunset := fmt.Sprintf("%v", unixToTime(m.weather.Sys.Sunset+m.weather.Timezone))

		return style.Render(fmt.Sprintf(
			"📍 City: %s, %s\n"+
				"🌡️  Temp: %s%.1f°C (Feels like %.1f°C)\n"+
				"🌡️  Min: %.1f°C, Max: %.1f°C\n"+
				"💧 Humidity: %d%%\n"+
				"🧭 Pressure: %dhPa\n"+
				"💨 Wind: %.1fm/s at %d°\n"+
				"☁️  Cloudiness: %d%%\n"+
				"👀 Visibility: %dm\n"+
				"🌤️  Condition: %s\n"+
				"🌅 Sunrise: %s, 🌇 Sunset: %s\n\n"+
				"(Press q to quit)",
			m.weather.Name,
			m.weather.Sys.Country,
			color.Render(""),
			tempC,
			feelsLikeC,
			tempMinC,
			tempMaxC,
			m.weather.Main.Humidity,
			m.weather.Main.Pressure,
			m.weather.Wind.Speed,
			m.weather.Wind.Deg,
			m.weather.Clouds.All,
			m.weather.Visibility,
			m.weather.Weather[0].Description,
			sunrise,
			sunset,
		))
	}

	return style.Render(fmt.Sprintf("City: %s\n(Press Enter to get weather)", m.textInput.View()))
}

// Helper to convert UNIX timestamp to time string
func unixToTime(ts int) string {
	t := time.Unix(int64(ts), 0)
	return t.Format("15:04:05")
}

// --- Fetch Weather ---

func fetchWeather(city string) tea.Cmd {

	return func() tea.Msg {
		API_KEY := os.Getenv("OPENWEATHER_API_KEY")
		if API_KEY == "" {
			return errMsg(fmt.Errorf("missing OPENWEATHER_API_KEY environment variable"))
		}

		url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s", city, API_KEY)
		resp, err := http.Get(url)
		if err != nil {
			return errMsg(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errMsg(err)
		}

		if resp.StatusCode != 200 {
			return errMsg(fmt.Errorf("API error: %s", resp.Status))
		}

		var data weatherData
		if err := json.Unmarshal(body, &data); err != nil {
			return errMsg(err)
		}
		return weatherMsg(data)
	}
}

// --- Main ---

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
