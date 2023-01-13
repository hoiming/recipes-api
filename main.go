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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io/ioutil"
	"log"
	"os"
)

var recipeHandler *handlers.RecipesHandler
var ctx context.Context
var err error
var client *mongo.Client
var recipes []models.Recipe
var collection *mongo.Collection

func init() {
	recipes = make([]models.Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)
	ctx := context.Background()
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
	recipeHandler = handlers.NewRecipesHandler(ctx, collection)
	//log.Println("Inserted recipes: ", len(insertManyResult.InsertedIDs))
}

func main() {
	router := gin.Default()
	router.POST("/recipes", recipeHandler.NewRecipeHandler)
	router.GET("/recipes", recipeHandler.ListRecipesHandler)
	router.PUT("/recipes/:id", recipeHandler.UpdateRecipeHandler)
	router.DELETE("/recipes/{id}", recipeHandler.DeleteRecipeHandler)
	router.GET("/recipes/search", recipeHandler.SearchRecipeHandler)
	router.Run()
}
