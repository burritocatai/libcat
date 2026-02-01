package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/libcat/libcat/internal/storage"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [book-id-or-title] [page]",
	Short: "Update reading progress for a book",
	Long: `Update the current page number for a book in your library.

Examples:
  libcat update "The Hobbit" 150
  libcat update 9780345339683 42
  libcat update  # Interactive mode - lists books to choose from`,
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
			fmt.Println("Your library is empty. Use 'libcat add' to add books first.")
			os.Exit(0)
		}

		var bookIndex int
		var newPage int

		if len(args) >= 2 {
			// Non-interactive mode
			query := args[0]
			pageStr := args[1]

			newPage, err = strconv.Atoi(pageStr)
			if err != nil || newPage < 0 {
				fmt.Fprintln(os.Stderr, "Error: page must be a non-negative number")
				os.Exit(1)
			}

			// Find book by ID or title
			bookIndex = -1
			for i, book := range library.Books {
				if book.ID == query || book.ISBN == query ||
					strings.EqualFold(book.Title, query) ||
					strings.Contains(strings.ToLower(book.Title), strings.ToLower(query)) {
					bookIndex = i
					break
				}
			}

			if bookIndex == -1 {
				fmt.Fprintf(os.Stderr, "Error: book '%s' not found in library\n", query)
				os.Exit(1)
			}
		} else {
			// Interactive mode
			fmt.Println("Your Library:")
			fmt.Println("-------------")
			for i, book := range library.Books {
				progress := ""
				if book.PageCount > 0 {
					progress = fmt.Sprintf(" [%d/%d pages]", book.CurrentPage, book.PageCount)
				}
				author := ""
				if len(book.Authors) > 0 {
					author = " by " + book.Authors[0]
				}
				fmt.Printf("%2d. %s%s%s\n", i+1, book.Title, author, progress)
			}

			// Select book
			fmt.Print("\nSelect book number: ")
			var input string
			fmt.Scanln(&input)

			num, err := strconv.Atoi(strings.TrimSpace(input))
			if err != nil || num < 1 || num > len(library.Books) {
				fmt.Println("Invalid selection.")
				os.Exit(0)
			}
			bookIndex = num - 1

			// Enter page number
			book := library.Books[bookIndex]
			if book.PageCount > 0 {
				fmt.Printf("Current page: %d / %d\n", book.CurrentPage, book.PageCount)
			}
			fmt.Print("Enter new page number: ")
			fmt.Scanln(&input)

			newPage, err = strconv.Atoi(strings.TrimSpace(input))
			if err != nil || newPage < 0 {
				fmt.Println("Invalid page number.")
				os.Exit(0)
			}
		}

		// Update the book
		book := &library.Books[bookIndex]
		oldPage := book.CurrentPage
		book.UpdatePage(newPage)

		if err := store.Save(library); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving library: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Updated '%s': page %d → %d", book.Title, oldPage, book.CurrentPage)
		if book.PageCount > 0 {
			fmt.Printf(" (%.1f%%)", book.Progress())
		}
		fmt.Println()

		switch book.Status {
		case "reading":
			fmt.Println("Status: Currently reading")
		case "finished":
			fmt.Println("Status: Finished! Congratulations!")
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
