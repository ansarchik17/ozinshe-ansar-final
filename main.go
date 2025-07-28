package main

import (
	"context"
	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
	"goozinshe/config"
	"goozinshe/docs"
	"goozinshe/handlers"
	"goozinshe/logger"
	"goozinshe/middlewares"
	"goozinshe/repositories"
	"time"
)

// @title           Ozinshe API
// @version         1.0
// @description     This is a simple celler server
// @termsOfService  http://swagger.io/terms/
//
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
//
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host      localhost:8070
// @BasePath  /
//
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
//
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

func main() {
	r := gin.New()
	gin.SetMode(gin.ReleaseMode)
	logger := logger.GetLogger()
	//err := r.SetTrustedProxies(nil)
	//if err != nil {
	//	log.Fatalf("Failed to set trusted proxies: %v", err)
	//}
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

	docs.SwaggerInfo.BasePath = "/"
	unauthorized.GET("/swagger/*any", swagger.WrapHandler(swaggerfiles.Handler))

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
