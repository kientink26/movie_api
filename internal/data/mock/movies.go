package mock

import (
	"movie_api/internal/data"
	"time"
)

var mockMovie = &data.Movie{
	ID:        1,
	CreatedAt: time.Now(),
	Title:     "Black Panther",
	Year:      2018,
	Runtime:   134,
	Genres:    []string{"sci-fi", "action", "adventure"},
	Version:   1,
}

type MockMovieModel struct{}

func (m MockMovieModel) Insert(movie *data.Movie) error {
	return nil
}
func (m MockMovieModel) Get(id int64) (*data.Movie, error) {
	switch id {
	case 1:
		return mockMovie, nil
	default:
		return nil, data.ErrRecordNotFound
	}
}
func (m MockMovieModel) Update(movie *data.Movie) error {
	return nil
}
func (m MockMovieModel) Delete(id int64) error {
	return nil
}
func (m MockMovieModel) GetAll(title string, genres []string, filters data.Filters) ([]*data.Movie, data.Metadata, error) {
	return []*data.Movie{mockMovie}, data.Metadata{}, nil
}
