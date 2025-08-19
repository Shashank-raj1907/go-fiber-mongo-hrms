package main

import (
	"log"

	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDB struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg *mongoDB

const dbName = "fiber-hrms"

// setted mongo db without user password for simplicity
const mongoURI = "mongodb://localhost:27017" + dbName

type Employee struct {
	ID     string  `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string  `json:"name" bson:"name"`
	Salary float64 `json:"salary" bson:"salary"`
	Age    float64 `json:"age" bson:"age"`
}

func connectDB() error {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)

	db := client.Database(dbName)

	if err != nil {
		return err
	}

	mg = &mongoDB{
		Client: client,
		Db:     db,
	}

	return nil
}

func main() {

	if err := connectDB(); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	app := fiber.New()

	app.Get("/employee", func(c *fiber.Ctx) error {

		query := bson.D{}

		cursor, err := mg.Db.Collection("employees").Find(c.Context(), query)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch employees",
			})
		}

		var employees []Employee = make([]Employee, 0)

		if err := cursor.All(c.Context(), &employees); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to decode employees",
			})
		}

		return c.JSON(employees)
	})

	app.Post("/employee", func(c *fiber.Ctx) error {
		employee := new(Employee)
		// get the employees collection
		collection := mg.Db.Collection("employees")

		// parse the request body into the employee struct
		if err := c.BodyParser(employee); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse request body",
			})
		}

		employee.ID = "" // MongoDB will generate an ID automatically
		// insert the employee into the collection
		result, err := collection.InsertOne(c.Context(), employee)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to insert employee",
			})
		}

		// Find the created employee by ID
		// This is to return the full employee object with the generated ID
		filter := bson.D{{Key: "_id", Value: result.InsertedID}}
		createdRecord := collection.FindOne(c.Context(), filter)

		// Decode the created record into an Employee struct
		createdEmployee := new(Employee)
		if err := createdRecord.Decode(createdEmployee); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to decode created employee",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(createdEmployee)
	})

	app.Put("/employee/:id")
	app.Delete("/employee/:id")

}
