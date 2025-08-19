# Fiber HRMS API

A simple CRUD API built with **Go (Fiber)** and **MongoDB** to manage employees.

## Run

1. Start MongoDB (local or Docker):
   ```bash
   docker run -d -p 27017:27017 --name mongodb mongo:6.0
````

2. Run the server:

   ```bash
   go run main.go
   ```

Server will be available at **[http://localhost:3000](http://localhost:3000)**

## 📡 Endpoints

* `GET /employee` → List employees
* `POST /employee` → Create employee
* `PUT /employee/:id` → Update employee
* `DELETE /employee/:id` → Delete employee

## 📊 Example Data

```json
{ "name": "Alice", "salary": 75000, "age": 29 }
```
