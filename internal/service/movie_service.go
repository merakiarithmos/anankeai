package service

import (
	"anankeai/internal/models"
	"anankeai/internal/repository"
)

type MovieService struct {
	repo *repository.MovieRepository
}

func NewMovieService(repo *repository.MovieRepository) *MovieService {
	return &MovieService{repo: repo}
}

func (s *MovieService) GetAllMovies() ([]models.Movie, error) {
	return s.repo.GetAllMovies()
}

func (s *MovieService) InsertMovie(m models.Movie) error {
	return s.repo.InsertMovie(m)
}
