package tui

import (
	"context"
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
	"github.com/iansinnott/browser-gopher/pkg/search"
	"github.com/iansinnott/browser-gopher/pkg/util"
	"github.com/pkg/errors"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)
var titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#fafafa"))
var urlStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#87BCF7"))

var HighlightStyle = lipgloss.NewStyle().Background(lipgloss.Color("#D8D7A0")).Foreground(lipgloss.Color("#000000"))

const UNTITLED = "<UNTITLED>"

type ListItem struct {
	// @note ItemTitle is thus named so as not to conflict with the Title() method, which is used by bubbletea
	ItemTitle, Desc, query string
	Body                   *string
	Date                   *time.Time
}

func (i ListItem) Title() string {
	var sb strings.Builder

	if i.Date != nil {
		sb.WriteString(i.Date.Format(util.FormatDateOnly))
		sb.WriteString(" ")
	}

	sb.WriteString(titleStyle.Render(i.ItemTitle))

	return sb.String()
}
func (i ListItem) Description() string { return urlStyle.Render(i.Desc) }
func (i ListItem) FilterValue() string { return i.ItemTitle + i.Desc }

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
	mapItem        ItemMapping
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
			var result *search.SearchResult
			var err error
			m.input, inputCmd = m.input.Update(msg)
			query := m.input.Value()
			if query == "" {
				result, err = m.dataProvider.RecentUrls(100)
			} else {
				result, err = m.searchProvider.SearchUrls(query)
			}
			// @note we ignored parse errors since they are quite expected when a user is typing
			if err != nil && !AcceptibleSearchError(err) {
				fmt.Println("search error", err)
				os.Exit(1)
			}
			items := ResultToItems(result, query, m.mapItem)
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

type ItemMapping func(x ListItem) list.Item

var identityMapping ItemMapping = func(x ListItem) list.Item {
	return x
}

// @todo Rather than taking providers this should probably take a search
// function that can handle customized querying. I.e. if I want to return only
// full-text documents with this current setup i would need to create a new
// SearchProvider that returns full-text docs for the SearchUrls call
func GetSearchProgram(
	ctx context.Context,
	initialQuery string,
	dataProvider search.DataProvider,
	searchProvider search.SearchProvider,
	mapItem *func(x ListItem) list.Item,
) (*tea.Program, error) {
	var err error
	var result *search.SearchResult

	var mapping ItemMapping
	if mapItem != nil {
		mapping = ItemMapping(*mapItem)
	} else {
		mapping = identityMapping
	}

	if initialQuery == "" {
		result, err = dataProvider.RecentUrls(100)
	} else {
		result, err = searchProvider.SearchUrls(initialQuery)
	}

	if err != nil && !AcceptibleSearchError(err) {
		return nil, errors.Wrap(err, "failed to get initial search results")
	}

	items := ResultToItems(result, "", mapping)

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
		mapItem:        mapping,
	}

	return tea.NewProgram(m, tea.WithAltScreen()), nil
}

func ResultToItems(result *search.SearchResult, query string, mapItem ItemMapping) []list.Item {
	if result == nil || len(result.Urls) == 0 {
		return []list.Item{ListItem{ItemTitle: "No results found"}}
	}

	urls := result.Urls
	items := []list.Item{}

	for _, u := range urls {
		displayUrl := u.Url
		displayTitle := UNTITLED
		if u.Title != nil {
			displayTitle = *u.Title
		}

		// @todo commented out while moving to sqlite
		// Highlighting
		// if result.Meta != nil {
		// 	hit, ok := lo.Find(result.Meta.Hits, func(x *bs.DocumentMatch) bool {
		// 		return x.ID == u.Id
		// 	})

		// 	if ok {
		// 		for k, locations := range hit.Locations {
		// 			switch k {
		// 			case "title":
		// 				displayTitle = search.HighlightAll(locations, displayTitle, HighlightStyle.Render)
		// 			case "url":
		// 				displayUrl = search.HighlightAll(locations, displayUrl, HighlightStyle.Render)
		// 			default:
		// 			}
		// 		}
		// 	}
		// }

		items = append(items, mapItem(ListItem{
			ItemTitle: displayTitle,
			Desc:      displayUrl,
			Date:      u.LastVisit,
			query:     query,
			Body:      u.Body,
		}))
	}

	return items
}

func AcceptibleSearchError(err error) bool {
	return strings.Contains(err.Error(), "parse error") || strings.Contains(err.Error(), "syntax error")
}
