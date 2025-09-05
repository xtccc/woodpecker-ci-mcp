package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFA500")).
			MarginBottom(1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0EA5E9")).
			MarginBottom(1)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			MarginLeft(2)

	instructionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")).
				MarginLeft(2)
)

func PromptForToken(woodpeckerURL string) (string, error) {
	fmt.Println(titleStyle.Render("🔐 Woodpecker CI Authentication"))
	fmt.Println(infoStyle.Render("To use this MCP server, you need a Personal Access Token (PAT) from your Woodpecker CI server."))

	fmt.Println(instructionStyle.Render("Steps to get your token:"))
	fmt.Println(instructionStyle.Render("1. Open your browser and go to: " + woodpeckerURL))
	fmt.Println(instructionStyle.Render("2. Log in to your Woodpecker CI account"))
	fmt.Println(instructionStyle.Render("3. Click on your user icon in the top right corner"))
	fmt.Println(instructionStyle.Render("4. Go to your personal profile page"))
	fmt.Println(instructionStyle.Render("5. Generate a new Personal Access Token"))
	fmt.Println(instructionStyle.Render("6. Copy the token and paste it below"))

	fmt.Println(warningStyle.Render("⚠️  The token will not be displayed as you type for security reasons."))

	fmt.Println("\n" + lipgloss.NewStyle().Bold(true).Render("Enter your Woodpecker CI Personal Access Token: "))

	// Read token securely (hidden input)
	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read token: %w", err)
	}

	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}

	fmt.Println() // Add newline after hidden input
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render("✅ Token received successfully!"))

	return token, nil
}

func PromptForURL() (string, error) {
	fmt.Println(titleStyle.Render("🛠️  Woodpecker CI Server Configuration"))
	fmt.Println(infoStyle.Render("Please enter your Woodpecker CI server URL."))

	fmt.Println(warningStyle.Render("Examples:"))
	fmt.Println(warningStyle.Render("  • https://woodpecker.example.com"))
	fmt.Println(warningStyle.Render("  • https://ci.mycompany.com"))
	fmt.Println(warningStyle.Render("  • http://localhost:8000 (for local development)"))

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n" + lipgloss.NewStyle().Bold(true).Render("Enter Woodpecker CI server URL: "))

	url, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read URL: %w", err)
	}

	url = strings.TrimSpace(url)
	if url == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	// Basic URL validation
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("URL must start with http:// or https://")
	}

	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render("✅ URL configured successfully!"))

	return url, nil
}

func ConfirmConfiguration(url, tokenPreview string) bool {
	fmt.Println(titleStyle.Render("📋 Configuration Summary"))
	fmt.Printf("Server URL: %s\n", infoStyle.Render(url))
	fmt.Printf("Token: %s\n", infoStyle.Render(tokenPreview+"***"))

	reader := bufio.NewReader(os.Stdin)
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("Is this configuration correct? (y/N): "))

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func MaskToken(token string) string {
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}
