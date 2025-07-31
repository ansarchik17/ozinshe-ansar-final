package repositories

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"goozinshe/logger"
	"goozinshe/models"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MoviesRepository struct {
	db *pgxpool.Pool
}

func NewMoviesRepository(conn *pgxpool.Pool) *MoviesRepository {
	return &MoviesRepository{db: conn}
}

func (r *MoviesRepository) FindById(c context.Context, id int) (models.Movie, error) {
	sql :=
		`
select 
m.id,
m.title,
m.description,
m.release_year,
m.director,
m.rating,
m.is_watched,
m.trailer_url,
m.poster_url,
g.id,
g.title
from movies m
join movies_genres mg on mg.movie_id = m.id
join genres g on mg.genre_id  = g.id
where m.id = $1
	`

	logger := logger.GetLogger()

	rows, err := r.db.Query(c, sql, id)
	defer rows.Close()
	if err != nil {
		logger.Error("Could not query database", zap.String("db_msg", err.Error()))
		return models.Movie{}, err
	}

	var movie *models.Movie

	for rows.Next() {
		var m models.Movie
		var g models.Genre

		err := rows.Scan(
			&m.Id,
			&m.Title,
			&m.Description,
			&m.ReleaseYear,
			&m.Director,
			&m.Rating,
			&m.IsWatched,
			&m.TrailerUrl,
			&m.PosterUrl,
			&g.Id,
			&g.Title,
		)
		if err != nil {
			logger.Error("Could not scan query row", zap.String("db_msg", err.Error()))
			return models.Movie{}, err
		}

		if movie != nil {
			m = *movie
		}

		m.Genres = append(m.Genres, g)
		movie = &m
	}

	err = rows.Err()
	if err != nil {
		logger.Error(err.Error())
		return models.Movie{}, err
	}
	return *movie, nil
}

func (r *MoviesRepository) FindAll(c context.Context, filters models.MovieFilters) ([]models.Movie, error) {
	sql :=
		`
select 
m.id,
m.title,
m.description,
m.release_year,
m.director,
m.rating,
m.is_watched,
m.trailer_url,
m.poster_url,
g.id,
g.title
from movies m
join movies_genres mg on mg.movie_id = m.id
join genres g on mg.genre_id  = g.id
where 1 = 1
`

	logger := logger.GetLogger()

	params := pgx.NamedArgs{}
	if filters.SearchTerm != "" {
		sql = fmt.Sprintf("%s and m.title ilike @s", sql)
		params["s"] = fmt.Sprintf("%%%s%%", filters.SearchTerm)
	}
	if filters.IsWatched != "" {
		isWatched, _ := strconv.ParseBool(filters.IsWatched)
		sql = fmt.Sprintf("%s and m.is_watched = @isWatched", sql)
		params["isWatched"] = isWatched
	}
	if filters.GenreId != "" {
		sql = fmt.Sprintf("%s and g.id = @genreId", sql)
		params["genreId"] = filters.GenreId
	}
	if filters.Sort != "" {
		identifier := pgx.Identifier{filters.Sort}.Sanitize()
		sql = fmt.Sprintf("%s order by m.%s", sql, identifier)
	}
	rows, err := r.db.Query(c, sql, params)
	if err != nil {
		logger.Error("Could not query database", zap.String("db_msg", err.Error()))
		return nil, err
	}

	movies := make([]*models.Movie, 0)
	moviesMap := make(map[int]*models.Movie)

	for rows.Next() {
		var m models.Movie
		var g models.Genre

		err := rows.Scan(
			&m.Id,
			&m.Title,
			&m.Description,
			&m.ReleaseYear,
			&m.Director,
			&m.Rating,
			&m.IsWatched,
			&m.TrailerUrl,
			&m.PosterUrl,
			&g.Id,
			&g.Title,
		)
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}

		if _, exists := moviesMap[m.Id]; !exists {
			moviesMap[m.Id] = &m
			movies = append(movies, &m)
		}

		moviesMap[m.Id].Genres = append(moviesMap[m.Id].Genres, g)
	}
	err = rows.Err()
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	concreteMovies := make([]models.Movie, 0, len(movies))
	for _, v := range movies {
		concreteMovies = append(concreteMovies, *v)
	}

	return concreteMovies, nil
}

func (r *MoviesRepository) Create(c context.Context, movie models.Movie) (int, error) {
	var id int

	tx, err := r.db.Begin(c)
	if err != nil {
		return 0, nil
	}

	row := tx.QueryRow(c,
		`
insert into movies(title, description, release_year, director, trailer_url, poster_url)
values($1, $2, $3, $4, $5, $6)
returning id
	`,
		movie.Title,
		movie.Description,
		movie.ReleaseYear,
		movie.Director,
		movie.TrailerUrl,
		movie.PosterUrl,
	)
	logger := logger.GetLogger()

	err = row.Scan(&id)
	if err != nil {
		logger.Error("Could not query database", zap.String("db_msg", err.Error()))
		return 0, nil
	}

	for _, genre := range movie.Genres {
		_, err = tx.Exec(c, "insert into movies_genres(movie_id, genre_id) values($1, $2)", id, genre.Id)
		if err != nil {
			logger.Error(err.Error())
			return 0, err
		}
	}

	err = tx.Commit(c)
	if err != nil {
		logger.Error(err.Error())
		return 0, nil
	}

	return id, nil
}

func (r *MoviesRepository) Update(c context.Context, id int, updatedMovie models.Movie) error {
	tx, err := r.db.Begin(c)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		c,
		`
update movies
set 
title = $1,
description = $2,
release_year = $3,
director = $4,
trailer_url = $5,
poster_url = $6
where id = $7
	`,
		updatedMovie.Title,
		updatedMovie.Description,
		updatedMovie.ReleaseYear,
		updatedMovie.Director,
		updatedMovie.TrailerUrl,
		updatedMovie.PosterUrl,
		id)
	logger := logger.GetLogger()
	if err != nil {
		logger.Error("Could not update database", zap.String("db_msg", err.Error()))
		return err
	}

	_, err = tx.Exec(c, "delete from movies_genres where movie_id = $1", id)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	for _, genre := range updatedMovie.Genres {
		_, err = tx.Exec(c, "insert into movies_genres(movie_id, genre_id) values($1, $2)", id, genre.Id)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
	}

	err = tx.Commit(c)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil
}

func (r *MoviesRepository) Delete(c context.Context, id int) error {
	tx, err := r.db.Begin(c)
	if err != nil {
		return err
	}
	logger := logger.GetLogger()
	_, err = tx.Exec(c, "delete from movies_genres where movie_id = $1", id)
	if err != nil {
		logger.Error("Could not delete database", zap.String("db_msg", err.Error()))
		return err
	}

	_, err = tx.Exec(c, "delete from movies where id = $1", id)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	err = tx.Commit(c)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil
}

func (r *MoviesRepository) SetRating(c context.Context, id int, rating int) error {
	logger := logger.GetLogger()
	_, err := r.db.Exec(c, "update movies set rating = $1 where id = $2", rating, id)
	if err != nil {
		logger.Error("Could not set rating", zap.String("db_msg", err.Error()))
		return err
	}

	return nil
}

func (r *MoviesRepository) SetWatched(c context.Context, id int, isWatched bool) error {
	logger := logger.GetLogger()
	_, err := r.db.Exec(c, "update movies set is_watched = $1 where id = $2", isWatched, id)
	if err != nil {
		logger.Error("Could not set isWatched", zap.String("db_msg", err.Error()))
		return err
	}

	return nil
}
