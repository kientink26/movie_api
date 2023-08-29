package data

import (
	"database/sql"
	"errors"
	"fmt"
	"movie_api/internal/validator"
	"strings"
	"time"
)

var ErrNonExistMovie = errors.New("movie id doesn't exist")

type Comment struct {
	ID        int64         `json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	Body      string        `json:"body"`
	Commenter *UserIdentity `json:"commenter,omitempty"`
}

func ValidateComment(v *validator.Validator, comment *Comment) {
	v.Check(comment.Body != "", "body", "must be provided")
	v.Check(len(comment.Body) <= 500, "body", "must not be more than 500 bytes long")
}

type CommentModel struct {
	DB *sql.DB
}

func (m CommentModel) Insert(comment *Comment, userID int64, movieID int64) error {
	query := `
INSERT INTO comments (body, user_id, movie_id)
VALUES ($1, $2, $3)
RETURNING id, created_at`
	args := []interface{}{comment.Body, userID, movieID}
	err := m.DB.QueryRow(query, args...).Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), `violates foreign key constraint "comments_movie_id_fkey"`):
			return ErrNonExistMovie
		case strings.Contains(err.Error(), `violates foreign key constraint "comments_user_id_fkey"`):
			panic(err)
		default:
			return err
		}
	}
	return nil
}

func (m CommentModel) GetAllForMovie(movieID int64, filters Filters) ([]*Comment, Metadata, error) {
	query := fmt.Sprintf(`
SELECT count(*) OVER(), comments.id, comments.created_at, comments.body, users.id, users.name, users.email
FROM comments
INNER JOIN movies ON comments.movie_id = movies.id
INNER JOIN users ON comments.user_id = users.id
WHERE movies.id = $1
ORDER BY comments.%s %s, comments.id ASC
LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())
	rows, err := m.DB.Query(query, movieID, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	totalRecords := 0
	comments := []*Comment{}
	for rows.Next() {
		var comment Comment
		comment.Commenter = &UserIdentity{}
		err := rows.Scan(
			&totalRecords,
			&comment.ID,
			&comment.CreatedAt,
			&comment.Body,
			&comment.Commenter.ID,
			&comment.Commenter.Name,
			&comment.Commenter.Email,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		comments = append(comments, &comment)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return comments, metadata, nil
}
