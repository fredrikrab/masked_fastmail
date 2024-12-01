package main

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

// Version information
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// statePriority defines the precedence of alias states for selection
var statePriority = map[AliasState]int{
	AliasEnabled:  0,
	AliasPending:  1,
	AliasDisabled: 2,
	AliasDeleted:  3,
}

func main() {
	rootCmd := &cobra.Command{
		Use: `masked_fastmail <url>   (no flags)
  manage_fastmail <alias>`,
		Short: "Manage masked email aliases",
		Long: `A command-line tool to manage Fastmail.com masked email addresses.
Requires FASTMAIL_ACCOUNT_ID and FASTMAIL_API_KEY environment variables to be set.`,
		Example: `  # Create or get alias for a website:
  masked_fastmail example.com

  # Enable an existing alias:
  masked_fastmail --enable user.1234@fastmail.com`,

		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			showVersion, _ := cmd.Flags().GetBool("version")
			if showVersion {
				fmt.Printf("Version:\t%s\nCommit:\t\t%s\nBuild date:\t%s\n", version, commit, date)
				return nil
			}
			return runMaskedFastmail(cmd, args)
		},
	}

	rootCmd.Flags().BoolP("version", "v", false, "show version information")
	rootCmd.Flags().BoolP("enable", "e", false, "enable alias")
	rootCmd.Flags().BoolP("disable", "d", false, "disable alias (send to trash)")
	rootCmd.Flags().Bool("delete", false, "delete alias (bounce messages)")

	// Make flags mutually exclusive
	rootCmd.MarkFlagsMutuallyExclusive("enable", "disable", "delete")

	// Add completion support
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// selectPreferredAlias selects the best alias based on state priority
// Priority order: enabled > pending > disabled > deleted
// Returns nil if the input slice is empty.
func selectPreferredAlias(aliases []MaskedEmailInfo) *MaskedEmailInfo {
	if len(aliases) == 0 {
		return nil
	}

	// Validate all states are recognized
	for _, alias := range aliases {
		if _, ok := statePriority[alias.State]; !ok {
			// Log warning but continue with known states
			fmt.Fprintf(os.Stderr, "Warning: unknown alias state: %s\n", alias.State)
		}
	}

	selected := &aliases[0]
	selectedPriority := statePriority[selected.State]

	for i := 1; i < len(aliases); i++ {
		priority := statePriority[aliases[i].State]
		if priority < selectedPriority {
			selected = &aliases[i]
			selectedPriority = priority
		}
	}

	return selected
}

// runMaskedFastmail is the main command handler for the CLI application.
// It handles both alias creation/lookup and state management operations.
func runMaskedFastmail(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("exactly one URL or alias must be specified\n\n%s", cmd.UsageString())
	}

	// Note: We can remove this validation since we're using cobra's MarkFlagsMutuallyExclusive

	client, err := NewFastmailClient()
	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	identifier := args[0]

	// Check for state update flags
	enable, _ := cmd.Flags().GetBool("enable")
	disable, _ := cmd.Flags().GetBool("disable")
	delete, _ := cmd.Flags().GetBool("delete")

	if enable || disable || delete {
		return handleStateUpdate(client, identifier, enable, disable, delete)
	}
	return handleAliasCreation(client, identifier)
}

// handleStateUpdate manages the state changes of existing aliases
func handleStateUpdate(client *FastmailClient, identifier string, enable, disable, delete bool) error {
	var newState AliasState
	switch {
	case enable:
		newState = AliasEnabled
	case disable:
		newState = AliasDisabled
	case delete:
		newState = AliasDeleted
	}

	// Get current state
	targetAlias, err := client.GetAliasByEmail(identifier)
	if err != nil {
		return fmt.Errorf("failed to get alias: %w", err)
	}

	err = client.UpdateAliasStatus(targetAlias, newState)
	if err != nil {
		return fmt.Errorf("failed to update alias status: %w", err)
	}
	return nil
}

// Handle get/create alias
func handleAliasCreation(client *FastmailClient, identifier string) error {
	aliases, err := client.GetAliases(identifier)
	if err != nil {
		return fmt.Errorf("failed to get aliases: %w", err)
	}
	selectedAlias := selectPreferredAlias(aliases)

	if selectedAlias == nil {
		// Create new alias
		fmt.Printf("No alias found for %s, creating new one...\n", identifier)
		newAlias, err := client.CreateAlias(identifier)
		if err != nil {
			return fmt.Errorf("failed to create alias: %w", err)
		}
		selectedAlias = newAlias
	} else if len(aliases) > 1 {
		fmt.Printf("Found %d aliases for %s:\n", len(aliases), identifier)
		for _, alias := range aliases {
			fmt.Printf("- %s (state: %s)\n", alias.Email, alias.State)
		}
		fmt.Println("\nSelected alias:")
	}

	fmt.Printf("%s (state: %s)", selectedAlias.Email, selectedAlias.State)
	if err := copyToClipboard(selectedAlias.Email); err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: Could not copy to clipboard: %v\n", err)
	} else {
		fmt.Println(" (copied to clipboard)")
	}
	return nil
}

// copyToClipboard attempts to copy the given text to the system clipboard
func copyToClipboard(text string) error {
	if err := clipboard.WriteAll(text); err != nil {
		return fmt.Errorf("clipboard operation failed: %w", err)
	}
	return nil
}
