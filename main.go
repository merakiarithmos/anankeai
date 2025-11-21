package main

import (
	"anankeai/internal/db"
	"anankeai/internal/models"
	"anankeai/internal/repository"
	"anankeai/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type MovieResponse struct {
	Movies []models.Movie `json:"movies"`
}

func generateMovies(number int, description string, movieChan chan models.Movie) error {
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

	jsonData := completion.Choices[0].Content

	var movieResponse MovieResponse

	if err := json.Unmarshal([]byte(jsonData), &movieResponse); err != nil {
		log.Fatal(err)
		return err
	}

	// print results
	for _, m := range movieResponse.Movies {
		fmt.Printf("%s (%d), %s\n", m.Title, m.Year, m.Director)
		movieChan <- m
	}

	return nil
}

func main() {
	// first load env variables
	godotenv.Load()

	db.Connect()
	defer db.Close()

	movieRepo := repository.NewMovieRepository()
	movieService := service.NewMovieService(movieRepo)

	existingMovies, err := movieService.GetAllMovies()
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range existingMovies {
		fmt.Printf("%s (%d), %s\n", m.Title, m.Year, m.Director)
	}

	var wg sync.WaitGroup
	movieChan := make(chan models.Movie, 5)

	wg.Add(2)

	go func() {
		defer wg.Done()
		generateMovies(10, "set in the 70s.", movieChan)
	}()
	go func() {
		defer wg.Done()
		generateMovies(10, "set in the 60s.", movieChan)
	}()

	// close the channel after all the goroutines are done
	go func() {
		wg.Wait()
		close(movieChan)
	}()

	for movie := range movieChan {
		movieService.InsertMovie(movie)
	}

	existingMovies, err = movieService.GetAllMovies()
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range existingMovies {
		fmt.Printf("%s (%d), %s\n", m.Title, m.Year, m.Director)
	}

}
