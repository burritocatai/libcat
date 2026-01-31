package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/libcat/libcat/internal/models"
)

const (
	baseURL     = "https://openlibrary.org"
	coversURL   = "https://covers.openlibrary.org"
	searchLimit = 20
)

// Client is the Open Library API client
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Open Library API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchResult represents a search result from Open Library
type SearchResult struct {
	NumFound int          `json:"numFound"`
	Start    int          `json:"start"`
	Docs     []SearchBook `json:"docs"`
}

// SearchBook represents a book in search results
type SearchBook struct {
	Key             string   `json:"key"`
	Title           string   `json:"title"`
	AuthorName      []string `json:"author_name"`
	FirstPublishYear int     `json:"first_publish_year"`
	ISBN            []string `json:"isbn"`
	Publisher       []string `json:"publisher"`
	NumberOfPages   int      `json:"number_of_pages_median"`
	CoverI          int      `json:"cover_i"`
}

// ISBNResponse represents the response from ISBN lookup
type ISBNResponse struct {
	Title       string `json:"title"`
	Authors     []struct {
		Key string `json:"key"`
	} `json:"authors"`
	Publishers    []string `json:"publishers"`
	PublishDate   string   `json:"publish_date"`
	NumberOfPages int      `json:"number_of_pages"`
	Covers        []int    `json:"covers"`
	Key           string   `json:"key"`
	Description   interface{} `json:"description"`
}

// AuthorResponse represents author data
type AuthorResponse struct {
	Name string `json:"name"`
}

// SearchBooks searches for books by title or author
func (c *Client) SearchBooks(query string) ([]models.Book, error) {
	encodedQuery := url.QueryEscape(query)
	reqURL := fmt.Sprintf("%s/search.json?q=%s&limit=%d", baseURL, encodedQuery, searchLimit)

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search books: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	books := make([]models.Book, 0, len(result.Docs))
	for _, doc := range result.Docs {
		book := models.Book{
			Title:       doc.Title,
			Authors:     doc.AuthorName,
			PublishYear: doc.FirstPublishYear,
			PageCount:   doc.NumberOfPages,
			Status:      models.StatusUnread,
			DateAdded:   time.Now(),
		}

		// Extract Open Library ID from key (e.g., "/works/OL123456W" -> "OL123456W")
		if doc.Key != "" {
			parts := strings.Split(doc.Key, "/")
			if len(parts) > 0 {
				book.OpenLibID = parts[len(parts)-1]
				book.ID = book.OpenLibID
			}
		}

		// Use first ISBN if available
		if len(doc.ISBN) > 0 {
			book.ISBN = doc.ISBN[0]
			book.ID = doc.ISBN[0]
		}

		// Get first publisher
		if len(doc.Publisher) > 0 {
			book.Publisher = doc.Publisher[0]
		}

		// Cover URL
		if doc.CoverI > 0 {
			book.CoverURL = fmt.Sprintf("%s/b/id/%d-M.jpg", coversURL, doc.CoverI)
		}

		books = append(books, book)
	}

	return books, nil
}

// GetBookByISBN fetches book details by ISBN
func (c *Client) GetBookByISBN(isbn string) (*models.Book, error) {
	// Clean ISBN (remove dashes)
	isbn = strings.ReplaceAll(isbn, "-", "")

	reqURL := fmt.Sprintf("%s/isbn/%s.json", baseURL, isbn)

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch book: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("book not found for ISBN: %s", isbn)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var isbnResp ISBNResponse
	if err := json.NewDecoder(resp.Body).Decode(&isbnResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	book := &models.Book{
		ID:            isbn,
		ISBN:          isbn,
		Title:         isbnResp.Title,
		PageCount:     isbnResp.NumberOfPages,
		Status:        models.StatusUnread,
		DateAdded:     time.Now(),
	}

	// Extract Open Library ID
	if isbnResp.Key != "" {
		parts := strings.Split(isbnResp.Key, "/")
		if len(parts) > 0 {
			book.OpenLibID = parts[len(parts)-1]
		}
	}

	// Get first publisher
	if len(isbnResp.Publishers) > 0 {
		book.Publisher = isbnResp.Publishers[0]
	}

	// Get cover URL
	if len(isbnResp.Covers) > 0 {
		book.CoverURL = fmt.Sprintf("%s/b/id/%d-M.jpg", coversURL, isbnResp.Covers[0])
	}

	// Parse description
	if isbnResp.Description != nil {
		switch desc := isbnResp.Description.(type) {
		case string:
			book.Description = desc
		case map[string]interface{}:
			if value, ok := desc["value"].(string); ok {
				book.Description = value
			}
		}
	}

	// Fetch author names
	for _, author := range isbnResp.Authors {
		authorName, err := c.getAuthorName(author.Key)
		if err == nil && authorName != "" {
			book.Authors = append(book.Authors, authorName)
		}
	}

	return book, nil
}

// getAuthorName fetches author name by their key
func (c *Client) getAuthorName(key string) (string, error) {
	reqURL := fmt.Sprintf("%s%s.json", baseURL, key)

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch author")
	}

	var author AuthorResponse
	if err := json.NewDecoder(resp.Body).Decode(&author); err != nil {
		return "", err
	}

	return author.Name, nil
}
