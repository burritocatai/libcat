package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/libcat/libcat/internal/api"
	"github.com/libcat/libcat/internal/storage"
	"github.com/spf13/cobra"
)

var (
	addISBN   string
	addSearch string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a book to your library",
	Long: `Add a book to your library by ISBN or search query.

Examples:
  libcat add --isbn 9780345339683
  libcat add --search "The Hobbit"
  libcat add -s "George Orwell 1984"`,
	Run: func(cmd *cobra.Command, args []string) {
		if addISBN == "" && addSearch == "" {
			fmt.Fprintln(os.Stderr, "Error: must provide --isbn or --search")
			cmd.Usage()
			os.Exit(1)
		}

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

		client := api.NewClient()

		if addISBN != "" {
			// Add by ISBN
			fmt.Printf("Looking up ISBN %s...\n", addISBN)
			book, err := client.GetBookByISBN(addISBN)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching book: %v\n", err)
				os.Exit(1)
			}

			// Check if already exists
			if existing := library.FindByID(book.ID); existing != nil {
				fmt.Printf("Book '%s' is already in your library.\n", book.Title)
				os.Exit(0)
			}

			library.AddBook(*book)
			if err := store.Save(library); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving library: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Added '%s' by %s to your library!\n", book.Title, strings.Join(book.Authors, ", "))
		} else {
			// Search and select
			fmt.Printf("Searching for '%s'...\n", addSearch)
			books, err := client.SearchBooks(addSearch)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error searching books: %v\n", err)
				os.Exit(1)
			}

			if len(books) == 0 {
				fmt.Println("No books found.")
				os.Exit(0)
			}

			// Display results
			fmt.Println("\nSearch Results:")
			fmt.Println("---------------")
			for i, book := range books {
				author := ""
				if len(book.Authors) > 0 {
					author = " by " + book.Authors[0]
				}
				year := ""
				if book.PublishYear > 0 {
					year = fmt.Sprintf(" (%d)", book.PublishYear)
				}
				fmt.Printf("%2d. %s%s%s\n", i+1, book.Title, author, year)
			}

			// Prompt for selection
			fmt.Print("\nEnter number to add (or 0 to cancel): ")
			var input string
			fmt.Scanln(&input)

			num, err := strconv.Atoi(strings.TrimSpace(input))
			if err != nil || num < 1 || num > len(books) {
				if num != 0 {
					fmt.Println("Invalid selection.")
				}
				os.Exit(0)
			}

			selectedBook := books[num-1]

			// Check if already exists
			if existing := library.FindByID(selectedBook.ID); existing != nil {
				fmt.Printf("Book '%s' is already in your library.\n", selectedBook.Title)
				os.Exit(0)
			}

			library.AddBook(selectedBook)
			if err := store.Save(library); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving library: %v\n", err)
				os.Exit(1)
			}

			author := ""
			if len(selectedBook.Authors) > 0 {
				author = " by " + selectedBook.Authors[0]
			}
			fmt.Printf("Added '%s'%s to your library!\n", selectedBook.Title, author)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVar(&addISBN, "isbn", "", "ISBN of the book to add")
	addCmd.Flags().StringVarP(&addSearch, "search", "s", "", "Search query (title or author)")
}
