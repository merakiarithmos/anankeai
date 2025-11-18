package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func main() {
	godotenv.Load()

	ctx := context.Background()
	llm, err := ollama.New(ollama.WithModel("mistral:7b"))

	if err != nil {
		log.Fatal("couldn't get the model")
	}

	completion, err := llms.GenerateFromSinglePrompt(ctx,
		llm,
		"Human: Write a poem about world peace.\nAssistant:\n",
		llms.WithTemperature(0.7),
	)

	if err != nil {
		log.Fatal("could not call LLM")
	}
	fmt.Println(completion)

}
