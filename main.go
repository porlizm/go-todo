package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/thedevsaddam/renderer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// App represents the application
type App struct {
	renderer *renderer.Render
	db       *mongo.Database
}

// Todo represents the todo model
type Todo struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title     string             `json:"title" bson:"title"`
	Completed bool               `json:"completed" bson:"completed"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize renderer with templates
	rnd := renderer.New(renderer.Options{
		ParseGlobPattern: "./templates/*.html",
	})

	// Connect to MongoDB
	client, err := connectToMongoDB()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database(os.Getenv("DB_NAME"))
	app := &App{
		renderer: rnd,
		db:       db,
	}

	// Create router
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// Static files
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(filesDir)))

	// Routes
	router.Get("/", app.homeHandler)
	router.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "static/favicon.ico"))
	})

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/todos", app.getTodos)
		r.Post("/todos", app.createTodo)
		r.Put("/todos/{id}", app.updateTodo)
		r.Delete("/todos/{id}", app.deleteTodo)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		log.Printf("Server running on http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped gracefully")
}

func connectToMongoDB() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	err := app.renderer.HTML(w, http.StatusOK, "home", nil)
	if err != nil {
		app.renderer.JSON(w, http.StatusInternalServerError, renderer.M{
			"error": "Failed to render home page",
		})
	}
}

func (app *App) getTodos(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := app.db.Collection("todos").Find(ctx, bson.M{})
	if err != nil {
		app.renderer.JSON(w, http.StatusInternalServerError, renderer.M{
			"error": "Failed to fetch todos",
		})
		return
	}
	defer cursor.Close(ctx)

	var todos []Todo
	if err = cursor.All(ctx, &todos); err != nil {
		app.renderer.JSON(w, http.StatusInternalServerError, renderer.M{
			"error": "Failed to decode todos",
		})
		return
	}

	app.renderer.JSON(w, http.StatusOK, renderer.M{
		"data": todos,
	})
}

func (app *App) createTodo(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		app.renderer.JSON(w, http.StatusBadRequest, renderer.M{
			"error": "Invalid request body",
		})
		return
	}

	if todo.Title == "" {
		app.renderer.JSON(w, http.StatusBadRequest, renderer.M{
			"error": "Title is required",
		})
		return
	}

	todo.ID = primitive.NewObjectID()
	todo.CreatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := app.db.Collection("todos").InsertOne(ctx, todo)
	if err != nil {
		app.renderer.JSON(w, http.StatusInternalServerError, renderer.M{
			"error": "Failed to create todo",
		})
		return
	}

	app.renderer.JSON(w, http.StatusCreated, todo)
}

func (app *App) updateTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		app.renderer.JSON(w, http.StatusBadRequest, renderer.M{
			"error": "Invalid ID format",
		})
		return
	}

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		app.renderer.JSON(w, http.StatusBadRequest, renderer.M{
			"error": "Invalid request body",
		})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"title":     todo.Title,
			"completed": todo.Completed,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = app.db.Collection("todos").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		app.renderer.JSON(w, http.StatusInternalServerError, renderer.M{
			"error": "Failed to update todo",
		})
		return
	}

	app.renderer.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo updated successfully",
	})
}

func (app *App) deleteTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		app.renderer.JSON(w, http.StatusBadRequest, renderer.M{
			"error": "Invalid ID format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = app.db.Collection("todos").DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		app.renderer.JSON(w, http.StatusInternalServerError, renderer.M{
			"error": "Failed to delete todo",
		})
		return
	}

	app.renderer.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo deleted successfully",
	})
}
