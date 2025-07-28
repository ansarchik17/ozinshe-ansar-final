package repositories

import (
	"context"
	"goozinshe/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GenresRepository struct {
	db *pgxpool.Pool
}

func NewGenresRepository(conn *pgxpool.Pool) *GenresRepository {
	return &GenresRepository{db: conn}
}

func (r *GenresRepository) FindById(c context.Context, id int) (models.Genre, error) {
	var genre models.Genre
	row := r.db.QueryRow(c, "select id, title from genres where id = $1", id)
	err := row.Scan(&genre.Id, &genre.Title)
	if err != nil {
		return models.Genre{}, err
	}

	return genre, nil
}

func (r *GenresRepository) FindAll(c context.Context) ([]models.Genre, error) {
	rows, err := r.db.Query(c, "select id, title from genres")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	genres := make([]models.Genre, 0)

	for rows.Next() {
		var genre models.Genre
		err = rows.Scan(&genre.Id, &genre.Title)
		if err != nil {
			return nil, err
		}

		genres = append(genres, genre)
	}

	return genres, nil
}

func (r *GenresRepository) FindAllByIds(c context.Context, ids []int) ([]models.Genre, error) {
	rows, err := r.db.Query(c, "select id, title from genres where id = any($1)", ids)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	genres := make([]models.Genre, 0)

	for rows.Next() {
		var genre models.Genre
		err = rows.Scan(&genre.Id, &genre.Title)
		if err != nil {
			return nil, err
		}

		genres = append(genres, genre)
	}

	return genres, nil
}

func (r *GenresRepository) Create(c context.Context, genre models.Genre) (int, error) {
	var id int
	row := r.db.QueryRow(c, "insert into genres (title) values ($1) returning id", genre.Title)
	err := row.Scan(&id)
	if err != nil {
		return 0, nil
	}

	return id, nil
}

func (r *GenresRepository) Update(c context.Context, id int, genre models.Genre) error {
	_, err := r.db.Exec(c, "update genres set title = $1 where id = $2", genre.Title, genre.Id)
	if err != nil {
		return err
	}

	return nil
}

func (r *GenresRepository) Delete(c context.Context, id int) error {
	_, err := r.db.Exec(c, "delete from genres where id = $1", id)
	if err != nil {
		return err
	}

	return nil
}
