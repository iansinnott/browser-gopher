package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
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
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)
var titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#fafafa"))
var urlStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#87BCF7"))

func getActiveStyle(style lipgloss.Style) lipgloss.Style {
	return style.Copy().Underline(true).Background(lipgloss.Color("#D8D7A0")).Foreground(lipgloss.Color("#000000"))
}

func getTitleString(title string) string {
	if title == UNTITLED {
		return title
	}
	return title
}

const UNTITLED = "<UNTITLED>"

type item struct {
	title, desc, query string
	date               *time.Time
}

func (i item) Title() string {
	var sb strings.Builder

	if i.date != nil {
		sb.WriteString(i.date.Format(util.FormatDateOnly))
		sb.WriteString(" ")
	}

	title := getTitleString(i.title)

	if i.query != "" {
		title = strings.ReplaceAll(title, i.query, getActiveStyle(titleStyle).Render(i.query))
	}

	sb.WriteString(title)

	return sb.String()
}
func (i item) Description() string {
	desc := urlStyle.Render(i.desc)

	if i.query != "" {
		desc = strings.ReplaceAll(desc, i.query, getActiveStyle(urlStyle).Render(i.query))
	}

	return desc
}
func (i item) FilterValue() string { return i.title + i.desc }

// @todo Support other systems that don't have `open`
// @todo should prob store a list of the `item` structs that have the URL rather than doing this string manipulation
func OpenItem(item list.Item) error {
	filterVal := item.FilterValue()
	re := regexp.MustCompile(`https?://`)
	loc := re.FindStringIndex(filterVal)
	url := filterVal[loc[0]:]
	fmt.Println("open", url)
	return exec.Command("open", url).Run()
}

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
		case "enter":
			item := m.list.SelectedItem()
			OpenItem(item) // @todo wrap this in a tea.Cmd to preserve purity
			return m, tea.Quit
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
			items := urlsToItems(result.Urls, query)
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
	listView := m.list.View()
	return docStyle.Render(m.input.View()) + "\n" + listView
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

		items := urlsToItems(result.Urls, "")

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

func urlsToItems(urls []types.UrlRow, query string) []list.Item {
	items := []list.Item{}
	for _, x := range urls {
		title := UNTITLED
		if x.Title != nil {
			title = *x.Title
		}
		items = append(items, item{
			title: title,
			desc:  x.Url,
			date:  x.LastVisit,
			query: query,
		})
	}
	return items
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().Bool("no-interactive", false, "disable interactive terminal interface. useful for scripting")
}
