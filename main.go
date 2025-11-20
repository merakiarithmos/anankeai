package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func readOrCreateFile(filename string) ([]byte, error) {
	fileData, err := os.ReadFile(filename)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("%s not found creating it.\n", filename)

			file, err := os.Create(filename)
			if err != nil {
				return nil, err

			}
			file.Close()
			fmt.Printf("created empty file: %s\n", filename)
		} else {
			return nil, err
		}
	}
	return fileData, nil

}

func main() {
	godotenv.Load()

	model := "mistral:7b"

	if v := os.Getenv("OLLAMA_MODEL_NAME"); v != "" {
		model = v
	}

	ctx := context.Background()
	llm, err := ollama.New(ollama.WithModel(model))

	if err != nil {
		log.Fatal("couldn't get the model")
	}

	fileData, err := readOrCreateFile("todo.txt")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var humanMsg string

	if fileData != nil {
		humanMsg = fmt.Sprintf("Please summarize the following todo list:\n\n%s", string(fileData))
	} else {
		humanMsg = "No todo list exists."
	}

	var msgs []llms.MessageContent

	// system message defines the available tools
	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful assistant who can summarize documents."))
	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, humanMsg))

	completion, err := llm.GenerateContent(ctx, msgs)

	log.Println(msgs)

	if err != nil {
		log.Fatal(err)
	}

	llmResponse := completion.Choices[0].Content

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter any updates to your todos")
	text, _ := reader.ReadString('\n')
	fmt.Println("Updating todo with new input")

	var newMsgs []llms.MessageContent
	newMsgs = append(msgs, llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful assistant"))

	prompt := fmt.Sprintf("The following is the initial state of the todo list:\n\n%s\nApply the following updates and output a summary of the todo list:\n\n%s", llmResponse, text)

	newMsgs = append(newMsgs, llms.TextParts(llms.ChatMessageTypeHuman, prompt))

	completion, err = llm.GenerateContent(ctx, newMsgs)

	if err != nil {
		log.Fatal(err)
	}

	llmResponse = completion.Choices[0].Content
	fmt.Println(llmResponse)
	fmt.Println("writing to todo.txt..")

	data := []byte(llmResponse)

	err = os.WriteFile("todo.txt", data, 0644)
	if err != nil {
		panic(err)
	}

}
