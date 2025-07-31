package main

import (
	"context"
	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"goozinshe/config"
	"goozinshe/handlers"
	"goozinshe/logger"
	"goozinshe/middlewares"
	"goozinshe/repositories"
	"time"
)

func main() {
	r := gin.New()

	logger := logger.GetLogger()
	r.Use(
		ginzap.Ginzap(logger, time.RFC3339, true),
		ginzap.RecoveryWithZap(logger, true),
	)

	corsConfig := cors.Config{
		AllowAllOrigins: true,
		AllowHeaders:    []string{"*"},
		AllowMethods:    []string{"*"},
	}
	r.Use(cors.New(corsConfig))

	err := loadConfig()
	if err != nil {
		panic(err)
	}

	conn, err := connectToDb()
	if err != nil {
		panic(err)
	}

	moviesRepository := repositories.NewMoviesRepository(conn)
	genresRepository := repositories.NewGenresRepository(conn)
	watchlistRepository := repositories.NewWatchlistRepository(conn)
	usersRepository := repositories.NewUsersRepository(conn)
	moviesHandler := handlers.NewMoviesHandler(
		moviesRepository,
		genresRepository,
	)
	genresHandler := handlers.NewGenreHandlers(genresRepository)
	imageHandler := handlers.NewImageHandlers()
	watchlistHandlers := handlers.NewWatchlistHandler(moviesRepository, watchlistRepository)
	userHandlers := handlers.NewUsersHandlers(usersRepository)
	authHandlers := handlers.NewAuthHandlers(usersRepository)
	authorized := r.Group("")
	authorized.Use(middlewares.AuthMiddleware)
	//Movie handlers
	authorized.POST("/movies", moviesHandler.Create)
	authorized.GET("/movies/:id", moviesHandler.FindById)
	authorized.GET("/movies", moviesHandler.FindAll)
	authorized.PUT("/movies/:id", moviesHandler.Update)
	authorized.DELETE("/movies/:id", moviesHandler.Delete)
	authorized.PATCH("/movies/:movieId/rate", moviesHandler.HandleSetRating)
	authorized.PATCH("/movies/:movieId/setWatched", moviesHandler.HandleSetWatched)
	//Genre handlers
	authorized.POST("/genres", genresHandler.Create)
	authorized.GET("/genres/:id", genresHandler.FindById)
	authorized.GET("/genres", genresHandler.FindAll)
	authorized.PUT("/genres/:id", genresHandler.Update)
	authorized.DELETE("/genres/:id", genresHandler.Delete)
	//Watchlist handlers
	authorized.GET("/watchlist", watchlistHandlers.HandleGetMovies)
	authorized.DELETE("/watchlist/:movieId", watchlistHandlers.HandleRemoveMovie)
	authorized.POST("/watchlist/:movieId", watchlistHandlers.HandleAddMovie)
	//Users handlers
	authorized.POST("/users", userHandlers.Create)
	authorized.GET("/users", userHandlers.FindAll)
	authorized.GET("/users/:id", userHandlers.FindById)
	authorized.PUT("/users/:id", userHandlers.Update)
	authorized.PATCH("/users/:id/changePassword", userHandlers.ChangePassword)
	authorized.DELETE("/users/:id", userHandlers.Delete)
	authorized.POST("/auth/signOut", authHandlers.SignOut)
	authorized.GET("auth/userInfo", authHandlers.GetUserInfo)
	//Authorization handlers
	unauthorized := r.Group("")
	unauthorized.POST("/auth/signIn", authHandlers.SignIn)
	unauthorized.GET("/images/:imageId", imageHandler.HandleGetImageById)

	logger.Info("Application starting...")

	r.Run(config.Config.AppHost)
}

func loadConfig() error {
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	var mapConfig config.MapConfig
	err = viper.Unmarshal(&mapConfig)
	if err != nil {
		return err
	}

	config.Config = &mapConfig

	return nil
}

func connectToDb() (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), config.Config.DbConnectionString)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return conn, nil
}
