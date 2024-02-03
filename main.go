package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/lib/pq"
)

var postgres *embeddedpostgres.EmbeddedPostgres
var dbConn *sql.DB

func databaseLogic() {
	var err error

	postgres = embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().RuntimePath("embedded-postgres"))
	postgres.Start()

	dbConn, err = sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	//defer dbConn.Close()

	ping := dbConn.Ping()
	if ping != nil {
		panic(ping)
	}
}

func migrateDatabase() {
	createTable := `
	CREATE TABLE items (
		id serial PRIMARY KEY UNIQUE,
		name varchar(255) NOT NULL DEFAULT '',
		quantity int NOT NULL DEFAULT 0
	)
	`

	_, err := dbConn.Exec(createTable)
	if err != nil {
		panic(err)
	}

	populateTable := `
	INSERT INTO items (name, quantity) VALUES
	('Cheese', 5),
	('Milk', 10),
	('Bread', 15),
	('Lamb', 20)
	`

	_, err = dbConn.Exec(populateTable)
	if err != nil {
		panic(err)
	}
}

func killDatabase() {
	postgres.Stop()
}

type Item struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type client struct {
	isClosing bool
	mu        sync.Mutex
}

var clients = make(map[*websocket.Conn]*client) // Note: although large maps with pointer-like types (e.g. strings) as keys are slow, using pointers themselves as keys is acceptable and fast
var register = make(chan *websocket.Conn)
var broadcast = make(chan string)
var unregister = make(chan *websocket.Conn)

func runHub() {
	for {
		select {
		case connection := <-register:
			clients[connection] = &client{}
			log.Println("connection registered")

		case message := <-broadcast:
			log.Println("message received:", message)
			// Send the message to all clients
			for connection, c := range clients {
				go func(connection *websocket.Conn, c *client) { // send to each client in parallel so we don't block on a slow client
					c.mu.Lock()
					defer c.mu.Unlock()
					if c.isClosing {
						return
					}
					if err := connection.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
						c.isClosing = true
						log.Println("write error:", err)

						connection.WriteMessage(websocket.CloseMessage, []byte{})
						connection.Close()
						unregister <- connection
					}
				}(connection, c)
			}

		case connection := <-unregister:
			// Remove the client from the hub
			delete(clients, connection)

			log.Println("connection unregistered")
		}
	}
}

func main() {

	databaseLogic()
	defer dbConn.Close()
	migrateDatabase()
	defer killDatabase()

	app := fiber.New()

	go runHub()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	var messageChan = make(chan string)

	app.Post("/new-item", func(c *fiber.Ctx) error {
		item := new(Item)
		if err := c.BodyParser(item); err != nil {
			return c.Status(http.StatusBadRequest).SendString(err.Error())
		}

		_, err := dbConn.Exec("INSERT INTO items (name, quantity) VALUES ($1, $2)", item.Name, item.Quantity)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString(err.Error())
		}

		messageChan <- "New item added"

		return c.Status(http.StatusCreated).JSON(item)
	})

	app.Get("/item/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		row := dbConn.QueryRow("SELECT id, name, quantity FROM items WHERE id = $1", id)

		item := new(Item)
		err := row.Scan(&item.ID, &item.Name)
		if err != nil {
			return c.Status(http.StatusNotFound).SendString(err.Error())
		}

		return c.Status(http.StatusOK).JSON(item)
	})

	app.Get("/items", func(c *fiber.Ctx) error {
		rows, err := dbConn.Query("SELECT id, name, quantity FROM items ORDER BY id DESC LIMIT 10")
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString(err.Error())
		}
		defer rows.Close()

		items := make([]*Item, 0)
		for rows.Next() {
			item := new(Item)
			err := rows.Scan(&item.ID, &item.Name, &item.Quantity)
			if err != nil {
				return c.Status(http.StatusInternalServerError).SendString(err.Error())
			}
			items = append(items, item)
		}

		return c.Status(http.StatusOK).JSON(items)
	})

	app.Get("/ws/last-item", websocket.New(func(c *websocket.Conn) {

		defer func() {
			unregister <- c
			c.Close()
		}()

		register <- c

		for {

			message := <-messageChan

			fmt.Println("message received:", message)

			row := dbConn.QueryRow("SELECT id, name, quantity FROM items ORDER BY id DESC LIMIT 1")
			item := new(Item)
			err := row.Scan(&item.ID, &item.Name, &item.Quantity)
			if err != nil {
				c.WriteMessage(websocket.TextMessage, []byte(err.Error()))
				return
			}

			itemJson, _ := json.Marshal(item)

			broadcast <- string(itemJson)

			fmt.Println("message sent:", item)

		}

	}))

	app.Listen("127.0.0.1:3000")
}
