package main

import (
	"log"

	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
const mongoURI = "mongodb://localhost:27017/" + dbName

type Employee struct {
	ID     string  `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string  `json:"name" bson:"name"`
	Salary float64 `json:"salary" bson:"salary"`
	Age    float64 `json:"age" bson:"age"`
}

func connectDB() error {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))

	if err != nil {
		return err
	}

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

		// Decode the cursor into the employees slice
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

	app.Put("/employee/:id", func(c *fiber.Ctx) error {
		// Get the employee ID from the URL parameters
		idParam := c.Params("id")

		// Convert the string ID to a MongoDB ObjectID
		employeeId , err := primitive.ObjectIDFromHex(idParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid employee ID",
			})
		}

		employee := new(Employee)

		// Parse the request body into the employee struct
		if err := c.BodyParser(employee); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse request body",
			})
		}

		query := bson.D{{Key: "_id", Value: employeeId}}

		// Create an update document with the new employee data
		// The $set operator is used to update the fields
		// If the field does not exist, it will be created
		update := bson.D{
			{
				Key:  "$set",
				Value : bson.D{
					{Key: "name", Value: employee.Name},
					{Key: "salary", Value: employee.Salary},
					{Key: "age", Value: employee.Age},
				},
			},
		}

		// Attempt to update the employee
		// FindOneAndUpdate returns the updated document
		err = mg.Db.Collection("employees").FindOneAndUpdate(c.Context(),query,update).Err()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Employee not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update employee",
			})
		}

		// Return the updated employee
		employee.ID = idParam
		return c.Status(fiber.StatusOK).JSON(employee)
	})

	app.Delete("/employee/:id", func(c *fiber.Ctx) error {
		// Get the employee ID from the URL parameters
		employeeId, err := primitive.ObjectIDFromHex(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid employee ID",
			})
		}

		// Create a query to find the employee by ID
		query := bson.D{{Key: "_id", Value: employeeId}}

		// Attempt to delete the employee
		result, err := mg.Db.Collection("employees").DeleteOne(c.Context(), query)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete employee",
			})
		}

		// Check if any document was deleted
		if result.DeletedCount == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Employee not found",
			})
		}

		// Return a success response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Employee deleted successfully",
		})
	})

	log.Fatal(app.Listen(":3000"))
}
