package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	bs "github.com/blevesearch/bleve/v2/search"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/search"
	"github.com/iansinnott/browser-gopher/pkg/util"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)
var titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#fafafa"))
var urlStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#87BCF7"))

var HighlightStyle = lipgloss.NewStyle().Background(lipgloss.Color("#D8D7A0")).Foreground(lipgloss.Color("#000000"))

func highlightLocation(loc *bs.Location, text string) string {
	var sb strings.Builder

	sb.WriteString(text[:loc.Start])
	sb.WriteString(HighlightStyle.Render(text[loc.Start:loc.End]))
	sb.WriteString(text[loc.End:])

	return sb.String()
}

func highlightAll(locations bs.TermLocationMap, text string) string {
	s := text

	for _, locs := range locations {
		for _, loc := range locs {
			s = highlightLocation(loc, s)
		}
	}

	return s
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

	title := i.title

	sb.WriteString(titleStyle.Render(title))

	return sb.String()
}
func (i item) Description() string { return urlStyle.Render(i.desc) }
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
	dataProvider   search.DataProvider
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
				result, err = m.dataProvider.RecentUrls(100)
			} else {
				result, err = m.searchProvider.SearchUrls(query)
			}
			// @note we ignored parse errors since they are quite expected when a user is typing
			if err != nil && !acceptibleSearchError(err) {
				fmt.Println("search error", err)
				os.Exit(1)
			}
			items := resultToItems(result, query)
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

		dataProvider := search.NewSqlSearchProvider(cmd.Context(), config.Config)
		searchProvider := search.NewBleveSearchProvider(cmd.Context(), config.Config)
		initialQuery := ""

		if len(args) > 0 {
			initialQuery = args[0]
		}

		if noInteractive {
			if len(args) < 1 {
				fmt.Println("No search query provided.")
				os.Exit(1)
				return
			}

			result, err := searchProvider.SearchUrls(initialQuery)
			if err != nil {
				fmt.Println("search error", err)
				os.Exit(1)
				return
			}

			for _, x := range util.ReverseSlice(result.Urls) {
				fmt.Printf("%v %s %sv\n", x.LastVisit.Format("2006-01-02"), *x.Title, x.Url)
			}

			fmt.Printf("Found %d results for \"%s\"\n", result.Count, initialQuery)
			os.Exit(0)
			return
		}

		var result *search.URLQueryResult

		if initialQuery == "" {
			result, err = dataProvider.RecentUrls(100)
		} else {
			result, err = searchProvider.SearchUrls(initialQuery)
		}

		if err != nil && !acceptibleSearchError(err) {
			fmt.Println("search error", err)
			os.Exit(1)
		}

		items := resultToItems(result, "")

		// Input el
		input := textinput.New()
		input.Placeholder = "Search..."
		input.SetValue(initialQuery)
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
			searchProvider: searchProvider,
			dataProvider:   dataProvider,
		}

		p := tea.NewProgram(m, tea.WithAltScreen())

		if err := p.Start(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}

func resultToItems(result *search.URLQueryResult, query string) []list.Item {
	if result == nil || len(result.Urls) == 0 {
		return []list.Item{item{title: "No results found"}}
	}

	urls := result.Urls
	items := []list.Item{}

	for _, u := range urls {
		displayUrl := u.Url
		displayTitle := UNTITLED
		if u.Title != nil {
			displayTitle = *u.Title
		}

		// Highlighting
		if result.Meta != nil {
			hit, ok := lo.Find(result.Meta.Hits, func(x *bs.DocumentMatch) bool {
				return x.ID == u.UrlMd5
			})

			if ok {
				for k, locations := range hit.Locations {
					switch k {
					case "title":
						displayTitle = highlightAll(locations, displayTitle)
					case "url":
						displayUrl = highlightAll(locations, displayUrl)
					default:
					}
				}
			}
		}

		items = append(items, item{
			title: displayTitle,
			desc:  displayUrl,
			date:  u.LastVisit,
			query: query,
		})
	}
	return items
}

func acceptibleSearchError(err error) bool {
	return strings.Contains(err.Error(), "parse error") || strings.Contains(err.Error(), "syntax error")
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().Bool("no-interactive", false, "disable interactive terminal interface. useful for scripting")
}
