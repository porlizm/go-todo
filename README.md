# Go Todo API

A simple RESTful Todo API built with Go, Chi router, and MongoDB.

![Go](https://img.shields.io/badge/Go-1.20+-00ADD8?logo=go)
![MongoDB](https://img.shields.io/badge/MongoDB-4.4+-47A248?logo=mongodb)

## Features

- RESTful API endpoints for Todo operations
- MongoDB database integration
- Chi router for HTTP routing
- Renderer for JSON responses and HTML templates
- Environment variable configuration
- Graceful server shutdown
- Static file serving

## Prerequisites

- Go 1.20+
- MongoDB (local or Atlas)
- Git

## Installation

1. Clone the repository:
```bash
git clone https://github.com/porlizm/go-todo.git
cd go-todo

2. Install dependencies:
go mod download


3. Set up environment variables:

cp .env.example .env
Edit the .env file with your MongoDB connection details.

########################
Configuration
Environment variables:

Variable	Description	Default Value
MONGODB_URI	MongoDB connection string	mongodb://localhost:27017
DB_NAME	Database name	todoapp
PORT	Server port	9000

########################
Project Structure
Copy
/go-todo/
├── .env
├── go.mod
├── go.sum
├── main.go
├── README.md
├── static/
│   └── favicon.ico
└── templates/
    └── home.html

#########################
API Endpoints
Method	Endpoint	Description
GET	/	Home page
GET	/api/v1/todos	Get all todos
POST	/api/v1/todos	Create new todo
PUT	/api/v1/todos/:id	Update todo
DELETE	/api/v1/todos/:id	Delete todo

#########################
Request/Response Examples
Create Todo:

curl -X POST http://localhost:9000/api/v1/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Buy groceries", "completed": false}'
Response:

{
  "id": "507f1f77bcf86cd799439011",
  "title": "Buy groceries",
  "completed": false,
  "createdAt": "2023-05-20T12:00:00Z"
}

Get All Todos:


curl http://localhost:9000/api/v1/todos
Response:

{
  "data": [
    {
      "id": "507f1f77bcf86cd799439011",
      "title": "Buy groceries",
      "completed": false,
      "createdAt": "2023-05-20T12:00:00Z"
    }
  ]
}

#########################
Running the Application
1. Start MongoDB (if using local instance):

mongod

2. Run the application:

go run main.go

3. Access the application:

Home page: http://localhost:9000

API base: http://localhost:9000/api/v1

#########################
Testing
To test the API endpoints, you can use:

1. cURL commands (examples above)

2. Postman

3. Any HTTP client

#########################
Deployment
Local Deployment

go build -o todo-app
./todo-app

Docker Deployment
1. Build the image:
docker build -t go-todo .

2. Run the container:
docker run -p 9000:9000 --env-file .env go-todo

#########################
Contributing
1. Fork the repository

2. Create your feature branch (git checkout -b feature/fooBar)

3. Commit your changes (git commit -am 'Add some fooBar')

4. Push to the branch (git push origin feature/fooBar)

5. Create a new Pull Request

License
- MIT

Acknowledgments
Chi router team

MongoDB Go driver team

Renderer package author

Copy

### Additional Files Needed

1. Create `.env.example`:
MONGODB_URI=mongodb://localhost:27017
DB_NAME=todoapp
PORT=9000

Copy

2. Create `Dockerfile` (optional for Docker deployment):
```dockerfile
FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o todo-app .

EXPOSE 9000

CMD ["./todo-app"]