package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/libcat/libcat/internal/api"
	"github.com/libcat/libcat/internal/models"
	"github.com/libcat/libcat/internal/storage"
)

// View represents the current view mode
type View int

const (
	ViewList View = iota
	ViewDetail
	ViewSearch
	ViewSearchResults
	ViewUpdatePage
)

// Model is the main TUI model
type Model struct {
	// Data
	library *models.Library
	storage *storage.Storage
	api     *api.Client

	// UI state
	view          View
	cursor        int
	width         int
	height        int
	searchResults []models.Book
	message       string
	err           error

	// Input fields
	searchInput textinput.Model
	pageInput   textinput.Model
}

// NewModel creates a new TUI model
func NewModel(lib *models.Library, store *storage.Storage) Model {
	searchInput := textinput.New()
	searchInput.Placeholder = "Search by title or author..."
	searchInput.CharLimit = 100
	searchInput.Width = 40

	pageInput := textinput.New()
	pageInput.Placeholder = "Enter page number..."
	pageInput.CharLimit = 6
	pageInput.Width = 20

	return Model{
		library:     lib,
		storage:     store,
		api:         api.NewClient(),
		view:        ViewList,
		searchInput: searchInput,
		pageInput:   pageInput,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// For input views, handle special keys but pass others to the input
		if m.view == ViewSearch {
			switch msg.String() {
			case "enter":
				query := m.searchInput.Value()
				if query != "" {
					m.message = "Searching..."
					return m, m.searchBooks(query)
				}
				return m, nil
			case "esc":
				m.view = ViewList
				m.searchInput.Blur()
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			default:
				// Pass key to text input
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}
		}

		if m.view == ViewUpdatePage {
			switch msg.String() {
			case "enter":
				pageStr := m.pageInput.Value()
				if pageStr != "" {
					var page int
					if _, err := fmt.Sscanf(pageStr, "%d", &page); err == nil && page >= 0 {
						if m.cursor < len(m.library.Books) {
							m.library.Books[m.cursor].UpdatePage(page)
							if err := m.storage.Save(m.library); err != nil {
								m.err = err
							} else {
								m.message = fmt.Sprintf("Updated page to %d", page)
							}
						}
					} else {
						m.err = fmt.Errorf("invalid page number")
					}
				}
				m.view = ViewList
				m.pageInput.Blur()
				return m, nil
			case "esc":
				m.view = ViewList
				m.pageInput.Blur()
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			default:
				// Pass key to text input
				var cmd tea.Cmd
				m.pageInput, cmd = m.pageInput.Update(msg)
				return m, cmd
			}
		}

		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case searchResultMsg:
		m.searchResults = msg.books
		m.err = msg.err
		m.view = ViewSearchResults
		m.cursor = 0
		return m, nil

	case bookAddedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.message = ""
		} else {
			m.message = fmt.Sprintf("Added '%s' to library!", msg.book.Title)
			m.err = nil
			m.view = ViewList
			m.cursor = len(m.library.Books) - 1
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.String() {
	case "ctrl+c", "q":
		if m.view == ViewList {
			return m, tea.Quit
		}
		// Return to list view
		m.view = ViewList
		m.message = ""
		m.err = nil
		return m, nil

	case "esc":
		if m.view != ViewList {
			m.view = ViewList
			m.message = ""
			m.err = nil
			return m, nil
		}
	}

	// View-specific keys
	switch m.view {
	case ViewList:
		return m.handleListKeys(msg)
	case ViewDetail:
		return m.handleDetailKeys(msg)
	case ViewSearchResults:
		return m.handleSearchResultsKeys(msg)
	}

	return m, nil
}

func (m Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.library.Books)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.library.Books) > 0 {
			m.view = ViewDetail
		}
	case "a", "/":
		m.view = ViewSearch
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		return m, textinput.Blink
	case "u":
		if len(m.library.Books) > 0 {
			m.view = ViewUpdatePage
			m.pageInput.SetValue("")
			m.pageInput.Focus()
			return m, textinput.Blink
		}
	case "d":
		if len(m.library.Books) > 0 {
			book := m.library.Books[m.cursor]
			m.library.RemoveBook(book.ID)
			if err := m.storage.Save(m.library); err != nil {
				m.err = err
			} else {
				m.message = fmt.Sprintf("Removed '%s' from library", book.Title)
			}
			if m.cursor >= len(m.library.Books) && m.cursor > 0 {
				m.cursor--
			}
		}
	}
	return m, nil
}

func (m Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "u":
		m.view = ViewUpdatePage
		m.pageInput.SetValue("")
		m.pageInput.Focus()
		return m, textinput.Blink
	case "enter", "esc", "q":
		m.view = ViewList
	}
	return m, nil
}

func (m Model) handleSearchResultsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.searchResults)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.searchResults) > 0 {
			book := m.searchResults[m.cursor]
			return m, m.addBook(book)
		}
	case "esc", "q":
		m.view = ViewList
		m.searchResults = nil
		m.cursor = 0
	}
	return m, nil
}

// Messages
type searchResultMsg struct {
	books []models.Book
	err   error
}

type bookAddedMsg struct {
	book models.Book
	err  error
}

// Commands
func (m Model) searchBooks(query string) tea.Cmd {
	return func() tea.Msg {
		books, err := m.api.SearchBooks(query)
		return searchResultMsg{books: books, err: err}
	}
}

func (m Model) addBook(book models.Book) tea.Cmd {
	return func() tea.Msg {
		// Check if book already exists
		if existing := m.library.FindByID(book.ID); existing != nil {
			return bookAddedMsg{err: fmt.Errorf("book already in library")}
		}

		m.library.AddBook(book)
		if err := m.storage.Save(m.library); err != nil {
			return bookAddedMsg{err: err}
		}
		return bookAddedMsg{book: book}
	}
}

// View renders the TUI
func (m Model) View() string {
	var content string

	switch m.view {
	case ViewList:
		content = m.renderListView()
	case ViewDetail:
		content = m.renderDetailView()
	case ViewSearch:
		content = m.renderSearchView()
	case ViewSearchResults:
		content = m.renderSearchResultsView()
	case ViewUpdatePage:
		content = m.renderUpdatePageView()
	}

	// Add message/error at bottom
	if m.err != nil {
		content += "\n" + errorStyle.Render(m.err.Error())
	} else if m.message != "" {
		content += "\n" + successStyle.Render(m.message)
	}

	return content
}

func (m Model) renderListView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📚 LibCat - Your Book Library"))
	b.WriteString("\n\n")

	if len(m.library.Books) == 0 {
		b.WriteString(mutedColorStyle.Render("No books in your library yet.\n"))
		b.WriteString(mutedColorStyle.Render("Press 'a' or '/' to search and add books.\n"))
	} else {
		for i, book := range m.library.Books {
			cursor := "  "
			if i == m.cursor {
				cursor = "▸ "
			}

			// Status indicator
			var statusIndicator string
			switch book.Status {
			case models.StatusUnread:
				statusIndicator = statusUnreadStyle.Render("○")
			case models.StatusReading:
				statusIndicator = statusReadingStyle.Render("◐")
			case models.StatusFinished:
				statusIndicator = statusFinishedStyle.Render("●")
			}

			// Progress
			progress := ""
			if book.PageCount > 0 {
				progress = fmt.Sprintf(" [%d/%d]", book.CurrentPage, book.PageCount)
			}

			// Title and author
			author := ""
			if len(book.Authors) > 0 {
				author = " - " + book.Authors[0]
			}

			line := fmt.Sprintf("%s%s %s%s%s", cursor, statusIndicator, book.Title, author, progress)

			if i == m.cursor {
				b.WriteString(selectedItemStyle.Render(line))
			} else {
				b.WriteString(normalItemStyle.Render(line))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: navigate • enter: view • a: add book • u: update page • d: delete • q: quit"))

	return b.String()
}

func (m Model) renderDetailView() string {
	if m.cursor >= len(m.library.Books) {
		return "No book selected"
	}

	book := m.library.Books[m.cursor]
	var b strings.Builder

	b.WriteString(titleStyle.Render("📖 " + book.Title))
	b.WriteString("\n\n")

	// Authors
	if len(book.Authors) > 0 {
		b.WriteString(labelStyle.Render("Authors:"))
		b.WriteString(valueStyle.Render(strings.Join(book.Authors, ", ")))
		b.WriteString("\n")
	}

	// Publisher
	if book.Publisher != "" {
		b.WriteString(labelStyle.Render("Publisher:"))
		b.WriteString(valueStyle.Render(book.Publisher))
		b.WriteString("\n")
	}

	// Year
	if book.PublishYear > 0 {
		b.WriteString(labelStyle.Render("Year:"))
		b.WriteString(valueStyle.Render(fmt.Sprintf("%d", book.PublishYear)))
		b.WriteString("\n")
	}

	// ISBN
	if book.ISBN != "" {
		b.WriteString(labelStyle.Render("ISBN:"))
		b.WriteString(valueStyle.Render(book.ISBN))
		b.WriteString("\n")
	}

	// Status
	b.WriteString(labelStyle.Render("Status:"))
	switch book.Status {
	case models.StatusUnread:
		b.WriteString(statusUnreadStyle.Render("Unread"))
	case models.StatusReading:
		b.WriteString(statusReadingStyle.Render("Reading"))
	case models.StatusFinished:
		b.WriteString(statusFinishedStyle.Render("Finished"))
	}
	b.WriteString("\n")

	// Progress
	if book.PageCount > 0 {
		b.WriteString(labelStyle.Render("Progress:"))
		b.WriteString(valueStyle.Render(fmt.Sprintf("%d / %d pages (%.1f%%)", book.CurrentPage, book.PageCount, book.Progress())))
		b.WriteString("\n")
		b.WriteString(labelStyle.Render(""))
		b.WriteString(RenderProgress(book.CurrentPage, book.PageCount, 30))
		b.WriteString("\n")
	}

	// Dates
	b.WriteString(labelStyle.Render("Added:"))
	b.WriteString(valueStyle.Render(book.DateAdded.Format("Jan 2, 2006")))
	b.WriteString("\n")

	if book.DateStarted != nil {
		b.WriteString(labelStyle.Render("Started:"))
		b.WriteString(valueStyle.Render(book.DateStarted.Format("Jan 2, 2006")))
		b.WriteString("\n")
	}

	if book.DateFinished != nil {
		b.WriteString(labelStyle.Render("Finished:"))
		b.WriteString(valueStyle.Render(book.DateFinished.Format("Jan 2, 2006")))
		b.WriteString("\n")
	}

	// Description
	if book.Description != "" {
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Description:"))
		b.WriteString("\n")
		// Wrap description
		desc := book.Description
		if len(desc) > 500 {
			desc = desc[:500] + "..."
		}
		b.WriteString(valueStyle.Render(desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("u: update page • esc/q: back to list"))

	return boxStyle.Render(b.String())
}

func (m Model) renderSearchView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🔍 Search for Books"))
	b.WriteString("\n\n")
	b.WriteString("Enter a book title or author name:\n\n")
	b.WriteString(focusedInputStyle.Render(m.searchInput.View()))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("enter: search • esc: cancel"))

	return b.String()
}

func (m Model) renderSearchResultsView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🔍 Search Results"))
	b.WriteString("\n\n")

	if len(m.searchResults) == 0 {
		b.WriteString(mutedColorStyle.Render("No books found. Try a different search.\n"))
	} else {
		for i, book := range m.searchResults {
			cursor := "  "
			if i == m.cursor {
				cursor = "▸ "
			}

			// Title and author
			author := ""
			if len(book.Authors) > 0 {
				author = " - " + book.Authors[0]
			}

			year := ""
			if book.PublishYear > 0 {
				year = fmt.Sprintf(" (%d)", book.PublishYear)
			}

			line := fmt.Sprintf("%s%s%s%s", cursor, book.Title, author, year)

			if i == m.cursor {
				b.WriteString(selectedItemStyle.Render(line))
			} else {
				b.WriteString(normalItemStyle.Render(line))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: navigate • enter: add to library • esc: back"))

	return b.String()
}

func (m Model) renderUpdatePageView() string {
	if m.cursor >= len(m.library.Books) {
		return "No book selected"
	}

	book := m.library.Books[m.cursor]
	var b strings.Builder

	b.WriteString(titleStyle.Render("📝 Update Reading Progress"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Book: %s\n", book.Title))
	if book.PageCount > 0 {
		b.WriteString(fmt.Sprintf("Total pages: %d\n", book.PageCount))
		b.WriteString(fmt.Sprintf("Current page: %d\n\n", book.CurrentPage))
	}
	b.WriteString("Enter new page number:\n\n")
	b.WriteString(focusedInputStyle.Render(m.pageInput.View()))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("enter: save • esc: cancel"))

	return b.String()
}

// mutedColor as a lipgloss style for rendering
var mutedColorStyle = lipgloss.NewStyle().Foreground(mutedColor)
