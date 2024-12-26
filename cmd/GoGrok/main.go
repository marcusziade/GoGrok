package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"GoGrok/pkg/client"
	"GoGrok/pkg/types"
)

type chatHandler struct {
	buffer string
}

func (h *chatHandler) OnContent(content string) {
	h.buffer += content
	fmt.Print(content)
}

func (h *chatHandler) OnError(err error) {
	fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
}

func (h *chatHandler) OnComplete() {
	fmt.Println("\n")
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "gogrok",
		Short: "A CLI tool for interacting with Grok API",
	}

	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup GoGrok with your API key",
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter your xAI API key: ")
			apiKey, _ := reader.ReadString('\n')
			apiKey = strings.TrimSpace(apiKey)

			// Create .env file with API key
			envContent := fmt.Sprintf("XAI_API_KEY=%s\n", apiKey)
			if err := os.WriteFile(".env", []byte(envContent), 0o600); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating .env file: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("API key saved to .env file successfully!")
		},
	}

	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Start a chat session with Grok",
		Run: func(cmd *cobra.Command, args []string) {
			// Try to load .env file
			if err := godotenv.Load(); err != nil {
				fmt.Fprintf(os.Stderr, "Error loading .env file. Run 'gogrok setup' first.\n")
				os.Exit(1)
			}

			c, err := client.NewClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
				os.Exit(1)
			}

			handler := &chatHandler{}
			scanner := bufio.NewScanner(os.Stdin)

			// Initialize chat history
			var messages []types.Message

			// Add system message if provided
			systemMsg, _ := cmd.Flags().GetString("system")
			if systemMsg != "" {
				messages = append(messages, types.Message{
					Role:    "system",
					Content: systemMsg,
				})
			}

			fmt.Println("Starting chat session (type 'exit' to quit)")
			fmt.Println("----------------------------------------")

			for {
				fmt.Print("\nYou: ")
				if !scanner.Scan() {
					break
				}

				input := scanner.Text()
				if input == "exit" {
					break
				}

				// Add user message to history
				messages = append(messages, types.Message{
					Role:    "user",
					Content: input,
				})

				// Prepare request
				req := types.ChatRequest{
					Messages:    messages,
					Model:       "grok-2-1212",
					Stream:      true,
					Temperature: 0.7,
				}

				fmt.Print("\nGrok: ")
				if err := c.StreamChat(req, handler); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					continue
				}

				// Add assistant's response to history
				messages = append(messages, types.Message{
					Role:    "assistant",
					Content: handler.buffer,
				})
				handler.buffer = "" // Reset buffer for next response
			}
		},
	}

	// Add flags
	chatCmd.Flags().String("system", "", "Set system message for the chat session")

	// Add commands to root
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(chatCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
