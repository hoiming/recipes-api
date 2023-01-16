// Recipe API
//
// This is a simple reicpes API. You can find out more about the API at https://github.com/PacktPublishing/Building-Distributed-Applications-in-Gin.
//
// Schemes: http
// Host: localhost:8080
// BasePath: /
// Version: 1.0.0
// Contact Haiming Liang
// <haiming_liang@hotmail.com> haiming_liang
//
// Consumes:
// - applicaiton/json
//
// Produces:
// - applicatoin/json
// swagger:meta
package main

import (
	"RecipeApi/handlers"
	"RecipeApi/models"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var recipeHandler *handlers.RecipesHandler
var ctx context.Context
var err error
var client *mongo.Client
var recipes []models.Recipe
var collection *mongo.Collection
var redisClient *redis.Client
var authHandler *handlers.AuthHandler

func init() {
	recipes = make([]models.Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)
	ctx = context.Background()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(),
		readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	var listOfRecipes []interface{}
	for _, recipe := range recipes {
		listOfRecipes = append(listOfRecipes, recipe)
	}
	collection = client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	//insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)
	if err != nil {
		log.Fatal(err)
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:         os.Getenv("REDIS_ADDR"),
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           0,
		WriteTimeout: time.Second * time.Duration(500),
		ReadTimeout:  time.Second * time.Duration(500),
		IdleTimeout:  time.Second * time.Duration(60),
		PoolSize:     64,
		MinIdleConns: 16,
	})
	recipeHandler = handlers.NewRecipesHandler(ctx, collection, redisClient)
	//pong, err := redisClient.Ping(ctx).Result()
	//fmt.Println(pong, err)
	//log.Println("Inserted recipes: ", len(insertManyResult.InsertedIDs))
	collectionUsers := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")
	authHandler = handlers.NewAuthHandler(ctx, collectionUsers)
}

func main() {
	router := gin.Default()
	router.GET("/recipes", recipeHandler.ListRecipesHandler)
	authorized := router.Group("/")
	authorized.Use(authHandler.AuthMiddleware())
	{
		authorized.POST("/recipes", recipeHandler.NewRecipeHandler)
		authorized.PUT("/recipes/:id", recipeHandler.UpdateRecipeHandler)
		authorized.DELETE("/recipes/{id}", recipeHandler.DeleteRecipeHandler)
	}
	router.GET("/recipes/search", recipeHandler.SearchRecipeHandler)
	router.POST("/signin", authHandler.SignInHandler)
	router.POST("/refresh", authHandler.RefreshHandler)
	router.Run()
}
