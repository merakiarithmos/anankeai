package repository

import (
	"anankeai/internal/db"
	"anankeai/internal/models"
	"context"
	"time"
)

type MovieRepository struct{}

func NewMovieRepository() *MovieRepository {
	return &MovieRepository{}
}

func (r *MovieRepository) GetAllMovies() ([]models.Movie, error) {
	rows, err := db.Pool.Query(context.Background(), "SELECT title, director, year FROM movies;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie

	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(&m.Title, &m.Director, &m.Year); err != nil {
			continue
		}
		movies = append(movies, m)
	}

	// check for erros that may have occured during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *MovieRepository) InsertMovie(m models.Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// use parameterized query to prevent SQL injection
	_, err := db.Pool.Exec(ctx,
		"INSERT INTO movies (title, director, year) VALUES ($1, $2, $3)",
		m.Title, m.Director, m.Year,
	)
	return err
}
