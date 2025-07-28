package repositories

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"goozinshe/models"
	"time"
)

type WatchlistRepository struct {
	db *pgxpool.Pool
}

func NewWatchlistRepository(db *pgxpool.Pool) *WatchlistRepository {
	return &WatchlistRepository{db: db}
}

func (r *WatchlistRepository) GetMoviesFromWatchlist(c context.Context) ([]models.Movie, error) {
	sql := `
select m.id, 
       m.title, 
       m.description, 
       m.release_year, 
       m.director, 
       m.rating, 
       m.trailer_url, 
       m.poster_url,
       g.id,
       g.title
from watchlist wl
join movies m on wl.movie_id = m.id
join movies_genres mg on m.id = mg.movie_id
join genres g on mg.genre_id = g.id
order by wl.added_at
`

	rows, err := r.db.Query(c, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	moviesMap := make(map[int]*models.Movie)
	movies := make([]*models.Movie, 0)
	for rows.Next() {
		var movie models.Movie
		var genre models.Genre
		err := rows.Scan(&movie.Id, &movie.Title, &movie.Description, &movie.ReleaseYear, &movie.Director,
			&movie.Rating, &movie.TrailerUrl, &movie.PosterUrl, &genre.Id, &genre.Title)
		if err != nil {
			return nil, err
		}

		if _, exists := moviesMap[movie.Id]; !exists {
			moviesMap[movie.Id] = &movie
			movies = append(movies, &movie)
		}

		moviesMap[movie.Id].Genres = append(moviesMap[movie.Id].Genres, genre)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	concreteMovies := make([]models.Movie, 0, len(movies))
	for _, m := range movies {
		concreteMovies = append(concreteMovies, *m)
	}

	return concreteMovies, nil

}

func (r *WatchlistRepository) AddToWatchlist(c context.Context, movieId int) error {
	_, err := r.db.Exec(c, "insert into watchlist(movie_id, added_at) values($1, $2)", movieId, time.Now())
	return err
}

func (r *WatchlistRepository) RemoveFromWatchlist(c context.Context, movieId int) error {
	_, err := r.db.Exec(c, "delete from watchlist where movie_id = $1", movieId)
	return err
}
