package main

import (
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

func callLocalModel(ctx context.Context, prompt string) (string, error) {
	model := "mistral:7b"

	if v := os.Getenv("OLLAMA_MODEL_NAME"); v != "" {
		model = v
	}

	log.Println("Extracted model name")

	llm, err := ollama.New(ollama.WithModel(model))
	if err != nil {
		log.Fatalf("failure to create model, err: %v", err)
	}
	var msgs []llms.MessageContent

	// system message defines the available tools
	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful assistant."))
	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, prompt))

	log.Println("Calling model")
	completion, err := llm.GenerateContent(ctx, msgs)
	if err != nil {
		log.Fatalf("failure to call model, err: %v", err)
	}
	llmResponse := completion.Choices[0].Content

	return llmResponse, nil
}

func main() {
	godotenv.Load()

	ctx := context.Background()
	modelResponse, err := callLocalModel(ctx, "What is the capital of France?")
	if err != nil {
		panic(err)
	}
	fmt.Println(modelResponse)

}
