package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/libcat/libcat/internal/models"
	"github.com/libcat/libcat/internal/storage"
	"github.com/spf13/cobra"
)

var historyLimit int
var showStats bool

var historyCmd = &cobra.Command{
	Use:   "history [book-id-or-title]",
	Short: "View reading progress history for a book",
	Long: `View the page reading progress history for a book in your library.

By default shows the last 10 entries. Use --limit to change this.
Use --stats to see reading statistics like pages per day.

Examples:
  libcat history "The Hobbit"
  libcat history 9780345339683 --limit 5
  libcat history --stats`,
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

		var bookIndex int = -1

		if len(args) > 0 {
			query := args[0]
			// Find book by ID or title
			for i, book := range library.Books {
				if book.ID == query || book.ISBN == query ||
					strings.EqualFold(book.Title, query) ||
					strings.Contains(strings.ToLower(book.Title), strings.ToLower(query)) {
					bookIndex = i
					break
				}
			}
		}

		// If showStats is true and no book specified, show stats for all books
		if showStats && bookIndex == -1 {
			fmt.Println("Reading Statistics")
			fmt.Println("===================")
			fmt.Println()

			hasHistory := false
			for _, book := range library.Books {
				if len(book.PageHistory) > 0 {
					hasHistory = true
					totalPages, pagesPerDay, daysReading := book.GetReadingStats()
					fmt.Printf("Book: %s\n", book.Title)
					if len(book.Authors) > 0 {
						fmt.Printf("  by %s\n", book.Authors[0])
					}
					fmt.Printf("  Total pages read: %.0f\n", totalPages)
					fmt.Printf("  Pages per day: %.1f\n", pagesPerDay)
					fmt.Printf("  Days tracked: %d\n", daysReading)
					fmt.Printf("  Current progress: %d / %d (%.1f%%)\n", book.CurrentPage, book.PageCount, book.Progress())
					fmt.Printf("  History entries: %d\n", len(book.PageHistory))
					fmt.Println()
				}
			}

			if !hasHistory {
				fmt.Println("No reading history available.")
			}
			return
		}

		// If no book specified and not showing stats, list books to choose from
		if bookIndex == -1 {
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
				historyCount := ""
				if len(book.PageHistory) > 0 {
					historyCount = fmt.Sprintf(" (%d entries)", len(book.PageHistory))
				}
				fmt.Printf("%2d. %s%s%s%s\n", i+1, book.Title, author, progress, historyCount)
			}

			fmt.Print("\nSelect book number: ")
			var input string
			fmt.Scanln(&input)

			num := 0
			if _, err := fmt.Sscanf(strings.TrimSpace(input), "%d", &num); err != nil || num < 1 || num > len(library.Books) {
				fmt.Println("Invalid selection.")
				os.Exit(0)
			}
			bookIndex = num - 1
		}

		book := library.Books[bookIndex]

		fmt.Printf("Reading History for '%s'\n", book.Title)
		fmt.Println(strings.Repeat("=", len(book.Title)+20))
		fmt.Println()

		if len(book.PageHistory) == 0 {
			fmt.Println("No reading history recorded yet.")
			return
		}

		// Show statistics
		totalPages, pagesPerDay, daysReading := book.GetReadingStats()
		fmt.Printf("Statistics:\n")
		fmt.Printf("  Total pages read: %.0f\n", totalPages)
		if pagesPerDay > 0 {
			fmt.Printf("  Average pages per day: %.1f\n", pagesPerDay)
		}
		if daysReading > 0 {
			fmt.Printf("  Days tracked: %d\n", daysReading)
		}
		fmt.Printf("  Total history entries: %d\n\n", len(book.PageHistory))

		// Show recent history
		entries := book.GetRecentHistory(historyLimit)
		fmt.Printf("Recent History (last %d entries):\n", len(entries))
		fmt.Println(strings.Repeat("-", 40))

		// Show in reverse order (newest first)
		for i := len(entries) - 1; i >= 0; i-- {
			entry := entries[i]
			fmt.Printf("  %s  → page %d\n", entry.Timestamp.Format("2006-01-02 15:04"), entry.Page)
		}

		// Show first entry if not included
		if historyLimit < len(book.PageHistory) {
			fmt.Printf("\n... and %d more entries\n", len(book.PageHistory)-historyLimit)
			firstEntry := book.PageHistory[0]
			fmt.Printf("  %s  → page %d (first entry)\n", firstEntry.Timestamp.Format("2006-01-02 15:04"), firstEntry.Page)
		}
	},
}

var historyAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Show reading history for all books",
	Long:  `Show recent reading history for all books in your library.`,
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

		// Collect all history entries from all books
		type BookHistoryEntry struct {
			BookTitle string
			Entry     interface{}
			IsHistory bool
		}

		var allEntries []BookHistoryEntry

		for _, book := range library.Books {
			if len(book.PageHistory) > 0 {
				for _, entry := range book.PageHistory {
					allEntries = append(allEntries, BookHistoryEntry{
						BookTitle: book.Title,
						Entry:     entry,
						IsHistory: true,
					})
				}
			}
		}

		if len(allEntries) == 0 {
			fmt.Println("No reading history recorded yet.")
			return
		}

		// Sort by timestamp
		sort.Slice(allEntries, func(i, j int) bool {
			return allEntries[i].Entry.(models.PageHistoryEntry).Timestamp.After(
				allEntries[j].Entry.(models.PageHistoryEntry).Timestamp)
		})

		fmt.Printf("All Reading Activity (showing last %d entries)\n", historyLimit)
		fmt.Println(strings.Repeat("=", 50))
		fmt.Println()

		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		yesterday := today.AddDate(0, 0, -1)

		count := 0
		for _, he := range allEntries {
			if count >= historyLimit {
				break
			}
			entry := he.Entry.(models.PageHistoryEntry)

			timeLabel := entry.Timestamp.Format("2006-01-02 15:04")
			if entry.Timestamp.After(today) {
				timeLabel = "Today " + entry.Timestamp.Format("15:04")
			} else if entry.Timestamp.After(yesterday) {
				timeLabel = "Yesterday " + entry.Timestamp.Format("15:04")
			}

			fmt.Printf("%s  %-30s → page %d\n", timeLabel, he.BookTitle, entry.Page)
			count++
		}
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.AddCommand(historyAllCmd)

	historyCmd.Flags().IntVarP(&historyLimit, "limit", "l", 10, "Number of history entries to show")
	historyCmd.Flags().BoolVarP(&showStats, "stats", "s", false, "Show reading statistics")
}