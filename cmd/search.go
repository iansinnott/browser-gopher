package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/search"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
	"github.com/spf13/cobra"
	image "github.com/trashhalo/imgcat/component"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)
var titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#fafafa"))
var urlStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#87BCF7"))

func renderTitle(title string) string {
	if title == untitled {
		return title
	}

	return titleStyle.Render(title)
}

const untitled = "<UNTITLED>"

type item struct {
	title, desc string
	date        *time.Time
	img         *image.Model
}

func (i item) Title() string {
	var sb strings.Builder

	if i.img != nil {
		sb.WriteString(i.img.View())
		sb.WriteString(" ")
	}

	if i.date != nil {
		sb.WriteString(i.date.Format(util.FormatDateOnly))
		sb.WriteString(" ")
	}

	sb.WriteString(renderTitle(i.title))

	return sb.String()
}
func (i item) Description() string {
	return urlStyle.Render(i.desc)
}
func (i item) FilterValue() string { return i.title + i.desc }

type model struct {
	input          textinput.Model
	list           list.Model
	searchProvider search.SearchProvider
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl-c", "esc":
			return m, tea.Quit
		case "ctrl+n", "ctrl+j", "down":
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		case "ctrl+p", "ctrl+k", "up":
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		default:
			var inputCmd tea.Cmd
			var result *search.URLQueryResult
			var err error
			m.input, inputCmd = m.input.Update(msg)
			query := m.input.Value()
			if query == "" {
				result, err = m.searchProvider.RecentUrls(100)
			} else {
				result, err = m.searchProvider.SearchUrls(query)
			}
			if err != nil {
				fmt.Println("search error", err)
				os.Exit(1)
			}
			items := urlsToItems(result.Urls)
			listCmd := m.list.SetItems(items)
			return m, tea.Batch(inputCmd, listCmd)
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h*2, msg.Height-v*2)
	}

	return m, cmd
}

func (m model) View() string {
	// return m.input.View() + "\n" + m.list.View()
	return docStyle.Render(m.input.View()) + "\n" + m.list.View()
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Find URLs you've visited",
	Run: func(cmd *cobra.Command, args []string) {
		noInteractive, err := cmd.Flags().GetBool("no-interactive")
		if err != nil {
			fmt.Println("could not parse --no-interactive:", err)
			os.Exit(1)
		}

		searchProvider := search.NewSearchProvider(cmd.Context(), config.Config)

		if noInteractive {
			if len(args) < 1 {
				fmt.Println("No search query provided.")
				os.Exit(1)
				return
			}

			query := args[0]
			result, err := searchProvider.SearchUrls(query)
			if err != nil {
				fmt.Println("search error", err)
				os.Exit(1)
				return
			}

			for _, x := range util.ReverseSlice(result.Urls) {
				fmt.Printf("%v %s %sv\n", x.LastVisit.Format("2006-01-02"), *x.Title, x.Url)
			}

			fmt.Printf("Found %d results for \"%s\"\n", result.Count, query)
			os.Exit(0)
			return
		}

		result, err := searchProvider.RecentUrls(100)
		if err != nil {
			fmt.Println("search error", err)
			os.Exit(1)
		}

		items := urlsToItems(result.Urls)

		// Input el
		input := textinput.New()
		input.Placeholder = "Search..."
		input.Focus()

		// Search results list el
		listDelegate := list.NewDefaultDelegate()
		listDelegate.SetHeight(2)
		listDelegate.SetSpacing(1)
		list := list.New(items, listDelegate, 0, 0)
		list.SetFilteringEnabled(false)
		list.SetShowTitle(false)
		list.SetShowStatusBar(false)

		m := model{
			list:           list,
			input:          input,
			searchProvider: *searchProvider,
		}

		p := tea.NewProgram(m, tea.WithAltScreen())

		if err := p.Start(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}

func urlsToItems(urls []types.UrlRow) []list.Item {
	items := []list.Item{}
	for _, x := range urls {
		var title string = untitled
		if x.Title != nil {
			title = *x.Title
		}
		img := image.New(32, 32, "https://avatars.githubusercontent.com/u/3154865?s=40&v=4")
		items = append(items, item{
			title: title,
			desc:  x.Url,
			date:  x.LastVisit,
			img:   &img,
		})
	}
	return items
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().Bool("no-interactive", false, "disable interactive terminal interface. useful for scripting")
}
