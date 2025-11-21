package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type Movie struct {
	Title    string `json:"title"`
	Director string `json:"director"`
	Year     int    `json:"year"`
}

func generateMovies(number int, description string) error {
	format := &openai.ResponseFormat{
		Type: "json_schema",
		JSONSchema: &openai.ResponseFormatJSONSchema{
			Name: "movies_list",
			Schema: &openai.ResponseFormatJSONSchemaProperty{
				Type: "object", // top-level must be an object
				Properties: map[string]*openai.ResponseFormatJSONSchemaProperty{
					"movies": {
						Type: "array",
						Items: &openai.ResponseFormatJSONSchemaProperty{
							Type: "object",
							Properties: map[string]*openai.ResponseFormatJSONSchemaProperty{
								"title": {
									Type:        "string",
									Description: "The title of the movie.",
								},
								"director": {
									Type:        "string",
									Description: "The director of the movie.",
								},
								"year": {
									Type:        "integer",
									Description: "The year the movie was released.",
								},
							},
							Required:             []string{"title", "director", "year"},
							AdditionalProperties: false,
						},
					},
				},
				Required: []string{"movies"},
			},
			Strict: true,
		},
	}

	llm, err := openai.New(openai.WithModel("gpt-5-mini"), openai.WithResponseFormat(format))
	if err != nil {
		log.Fatal(err)
		return err
	}
	ctx := context.Background()

	prompt := fmt.Sprintf("Can I have %d incredible movies, %s? Please give me the title, director, year.", number, description)

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful assistant who has access to a large movie database."),
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	completion, err := llm.GenerateContent(ctx, content, llms.WithJSONMode())
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Println(completion.Choices[0].Content)

	return nil
}

func main() {
	godotenv.Load()
	go generateMovies(5, "set in the 80s.")
	go generateMovies(5, "incredibly scary.")
	go generateMovies(5, "that won best picture.")
	go generateMovies(5, "cheesy romcom.")
	go generateMovies(5, "set in Toronto.")
	go generateMovies(5, "that are about hiests.")

	time.Sleep(30 * time.Second)
}
