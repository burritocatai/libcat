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
	CurrentPage  int        `json:"current_page"`
	Status       ReadStatus `json:"status"`
	DateAdded    time.Time  `json:"date_added"`
	DateStarted  *time.Time `json:"date_started,omitempty"`
	DateFinished *time.Time `json:"date_finished,omitempty"`
	Notes        string     `json:"notes"`
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
