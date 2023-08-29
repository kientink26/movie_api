package mock

import "movie_api/internal/data"

func NewMockModels() data.Models {
	return data.Models{
		Movies: MockMovieModel{},
	}
}
