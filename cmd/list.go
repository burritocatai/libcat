package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/libcat/libcat/internal/models"
	"github.com/libcat/libcat/internal/storage"
	"github.com/spf13/cobra"
)

var (
	listStatus string
	listJSON   bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List books in your library",
	Long: `List all books in your library with their reading status.

Examples:
  libcat list
  libcat list --status reading
  libcat list --status finished`,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		store, err := storage.NewStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
			os.Exit(1)
		}

		library, err := store.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading library: %v\n", err)
			os.Exit(1)
		}

		if len(library.Books) == 0 {
			fmt.Println("Your library is empty. Use 'libcat add' to add books.")
			return
		}

		// Filter by status if specified
		books := library.Books
		if listStatus != "" {
			filtered := make([]models.Book, 0)
			for _, book := range books {
				if string(book.Status) == listStatus {
					filtered = append(filtered, book)
				}
			}
			books = filtered
		}

		if len(books) == 0 {
			fmt.Printf("No books with status '%s'.\n", listStatus)
			return
		}

		// Print header
		fmt.Printf("📚 Your Library (%d books)\n", len(books))
		fmt.Println(strings.Repeat("─", 60))

		for _, book := range books {
			// Status indicator
			var statusIcon string
			switch book.Status {
			case models.StatusUnread:
				statusIcon = "○"
			case models.StatusReading:
				statusIcon = "◐"
			case models.StatusFinished:
				statusIcon = "●"
			}

			// Author
			author := ""
			if len(book.Authors) > 0 {
				author = book.Authors[0]
			}

			// Progress
			progress := ""
			if book.PageCount > 0 {
				progress = fmt.Sprintf("%d/%d (%.0f%%)", book.CurrentPage, book.PageCount, book.Progress())
			}

			// Print book info
			fmt.Printf("%s %s\n", statusIcon, book.Title)
			if author != "" {
				fmt.Printf("  Author: %s\n", author)
			}
			if progress != "" {
				fmt.Printf("  Progress: %s\n", progress)
			}
			if book.ISBN != "" {
				fmt.Printf("  ISBN: %s\n", book.ISBN)
			}
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (unread, reading, finished)")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
}
