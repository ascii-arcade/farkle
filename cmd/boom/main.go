package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
)

type particle struct {
	projectile *harmonica.Projectile
	symbol     string
	color      lipgloss.Color
	ttl        int
}

type model struct {
	width, height  int
	exploding      bool
	particles      []*particle
	frame          int
	nextBurstFrame int
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*30, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "e":
			if !m.exploding {
				m.exploding = true
				m.frame = 0
				m.nextBurstFrame = 0
				return m, tick()
			}
		case "q":
			return m, tea.Quit
		}

	case tickMsg:
		if m.exploding {
			m.frame++

			if m.frame >= m.nextBurstFrame {
				x := float64(rand.IntN(m.width))
				y := float64(rand.IntN(m.height / 2)) // stay in upper half
				m = m.addFireworkAt(x, y)
				m.nextBurstFrame = m.frame + rand.IntN(20) + 30 // new: 30–50 frames
			}

			for _, p := range m.particles {
				p.projectile.Update()
				p.ttl--
			}

			live := m.particles[:0]
			for _, p := range m.particles {
				if p.ttl > 0 {
					live = append(live, p)
				}
			}
			m.particles = live

			return m, tick()
		}
	}
	return m, nil
}

func (m model) View() string {
	if !m.exploding {
		return "🎆 Press 'e' to launch fireworks! Press 'q' to quit."
	}

	screen := make([][]string, m.height)
	for i := range screen {
		screen[i] = make([]string, m.width)
		for j := range screen[i] {
			screen[i][j] = " "
		}
	}

	for _, p := range m.particles {
		pos := p.projectile.Position()
		x := int(pos.X)
		y := int(pos.Y)
		if x >= 0 && x < m.width && y >= 0 && y < m.height {
			style := lipgloss.NewStyle().Foreground(p.color)
			screen[y][x] = style.Render(p.symbol)
		}
	}

	var b strings.Builder
	for _, row := range screen {
		b.WriteString(strings.Join(row, ""))
		b.WriteRune('\n')
	}
	return b.String()
}

func (m model) addFireworkAt(centerX, centerY float64) model {
	symbols := []string{"*", "+", "@", "#", "%", "⚡", "💥", "✴", "☄"}

	numParticles := 200

	// Extended and vibrant color palettes
	palettes := [][]int{
		{196, 202, 208}, // red/orange
		{33, 39, 45},    // deep blue
		{40, 46, 82},    // green/teal
		{129, 135, 141}, // purple/magenta
		{220, 226, 228}, // bright yellows
		{51, 45, 39},    // icy blue
		{160, 161, 125}, // earthy
		{201, 207, 213}, // pastel pink/lavender
		{118, 154, 190}, // seafoam
		{199, 200, 201}, // hot pinks
	}
	palette := palettes[rand.IntN(len(palettes))]

	for i := 0; i < numParticles; i++ {
		// Spiral shape
		angle := float64(i) * 0.3
		speed := 0.1 * float64(i)
		vx := speed * math.Cos(angle)
		vy := speed * math.Sin(angle)

		p := harmonica.NewProjectile(
			harmonica.FPS(60),
			harmonica.Point{X: centerX, Y: centerY, Z: 0},
			harmonica.Vector{X: vx, Y: vy, Z: 0},
			harmonica.TerminalGravity,
		)

		// Dynamic color shimmer
		base := palette[i%len(palette)]
		offset := rand.IntN(2) * 36 // slight brightness variance
		colorCode := base + offset
		if colorCode > 255 {
			colorCode = 255
		}
		color := lipgloss.Color(fmt.Sprintf("%d", colorCode))

		m.particles = append(m.particles, &particle{
			projectile: p,
			symbol:     symbols[rand.IntN(len(symbols))],
			color:      color,
			ttl:        rand.IntN(40) + 50,
		})
	}

	return m
}

func main() {
	tea.NewProgram(model{}, tea.WithAltScreen()).Run()
}
