package models

import "time"

// Book represents a book in the user's library
type Book struct {
	// Book identification
	ID        string `json:"id"`         // Internal ID (ISBN or generated)
	ISBN      string `json:"isbn"`       // ISBN-10 or ISBN-13
	OpenLibID string `json:"open_lib_id"` // Open Library ID (e.g., OL123456W)

	// Book metadata from API
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	Publisher   string   `json:"publisher"`
	PublishYear int      `json:"publish_year"`
	PageCount   int      `json:"page_count"`
	CoverURL    string   `json:"cover_url"`
	Description string   `json:"description"`

	// User tracking data
	CurrentPage  int                `json:"current_page"`
	PageHistory  []PageHistoryEntry `json:"page_history"`
	Status       ReadStatus         `json:"status"`
	DateAdded    time.Time          `json:"date_added"`
	DateStarted  *time.Time         `json:"date_started,omitempty"`
	DateFinished *time.Time         `json:"date_finished,omitempty"`
	Notes        string             `json:"notes"`
}

// ReadStatus represents the reading status of a book
type ReadStatus string

const (
	StatusUnread   ReadStatus = "unread"
	StatusReading  ReadStatus = "reading"
	StatusFinished ReadStatus = "finished"
)

// Progress returns the reading progress as a percentage (0-100)
func (b *Book) Progress() float64 {
	if b.PageCount == 0 {
		return 0
	}
	return float64(b.CurrentPage) / float64(b.PageCount) * 100
}

// UpdatePage updates the current page and status
func (b *Book) UpdatePage(page int) {
	// Record history before updating (only if page actually changes)
	if page != b.CurrentPage {
		b.PageHistory = append(b.PageHistory, PageHistoryEntry{
			Timestamp: time.Now(),
			Page:      page,
		})
	}

	b.CurrentPage = page

	if page == 0 {
		b.Status = StatusUnread
		b.DateStarted = nil
	} else if page >= b.PageCount && b.PageCount > 0 {
		b.CurrentPage = b.PageCount
		b.Status = StatusFinished
		now := time.Now()
		b.DateFinished = &now
	} else {
		if b.Status == StatusUnread {
			now := time.Now()
			b.DateStarted = &now
		}
		b.Status = StatusReading
	}
}

// GetReadingStats calculates reading statistics based on page history
func (b *Book) GetReadingStats() (totalPagesRead, pagesPerDay float64, daysReading int) {
	if len(b.PageHistory) == 0 {
		return 0, 0, 0
	}

	// Find the first positive page count
	firstPage := 0
	for _, entry := range b.PageHistory {
		if entry.Page > 0 {
			firstPage = entry.Page
			break
		}
	}

	totalPagesRead = float64(b.CurrentPage - firstPage)
	if totalPagesRead < 0 {
		totalPagesRead = 0
	}

	// Calculate days between first and last entry
	firstTime := b.PageHistory[0].Timestamp
	lastTime := b.PageHistory[len(b.PageHistory)-1].Timestamp

	duration := lastTime.Sub(firstTime).Hours() / 24
	if duration > 0 {
		pagesPerDay = totalPagesRead / duration
		daysReading = int(duration) + 1
	}

	return
}

// GetRecentHistory returns the last N history entries
func (b *Book) GetRecentHistory(n int) []PageHistoryEntry {
	if len(b.PageHistory) == 0 {
		return nil
	}
	start := len(b.PageHistory) - n
	if start < 0 {
		start = 0
	}
	return b.PageHistory[start:]
}

// PageHistoryEntry represents a single page progress entry
type PageHistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Page      int       `json:"page"`
}

// Library represents the user's book collection
type Library struct {
	Books   []Book    `json:"books"`
	Updated time.Time `json:"updated"`
}

// FindByID finds a book by its ID
func (l *Library) FindByID(id string) *Book {
	for i := range l.Books {
		if l.Books[i].ID == id {
			return &l.Books[i]
		}
	}
	return nil
}

// FindByISBN finds a book by its ISBN
func (l *Library) FindByISBN(isbn string) *Book {
	for i := range l.Books {
		if l.Books[i].ISBN == isbn {
			return &l.Books[i]
		}
	}
	return nil
}

// AddBook adds a book to the library
func (l *Library) AddBook(book Book) {
	l.Books = append(l.Books, book)
	l.Updated = time.Now()
}

// RemoveBook removes a book from the library by ID
func (l *Library) RemoveBook(id string) bool {
	for i := range l.Books {
		if l.Books[i].ID == id {
			l.Books = append(l.Books[:i], l.Books[i+1:]...)
			l.Updated = time.Now()
			return true
		}
	}
	return false
}
