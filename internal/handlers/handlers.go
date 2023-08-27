package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var wsChan = make(chan WsPayload)

var clients = make(map[WebSocketConnection]string)

// views is the jet view set
var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

// upgradeConnection is the websocket upgrader from gorilla/websockets
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Home renders the home page
func Home(w http.ResponseWriter, r *http.Request) {
	log.Println("Render Home")
	err := renderPage(w, "home.hbs", nil)
	if err != nil {
		log.Printf("Error in rendering Home page: %v \n", err)
	}
}

// WebSocketConnection is a wrapper for our websocket connection, in case
// we ever need to put more data into the struct
type WebSocketConnection struct {
	*websocket.Conn
}

// WsJsonResponse defines the response sent back from websocket
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Username       string   `json:"username"`
	Person         string   `json:"person"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

// WsPayload defines the websocket request from the client
type WsPayload struct {
	Action   string              `json:"action"`
	Username string              `json:"username"`
	Person   string              `json:"person"`
	Message  string              `json:"message"`
	Conn     WebSocketConnection `json:"-"`
}

// WsEndpoint upgrades connection to websocket
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error in connection upgrade: %v \n", err)
	}

	log.Println("Client connected to endpoint")

	var response WsJsonResponse
	response.Message = `<em><small>Connected to server</small></em>`

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	go ListenForWs(&conn)
}

// ListenForWs is a goroutine that handles communication between server and client, and
// feeds data into the wsChan
func ListenForWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			log.Println(err)
		} else {
			username := payload.Username
			message := payload.Message
			person := payload.Person

			// Insert data into PostgreSQL

			// connectionString := "postgresql://postgres:hBcbrWa3PQhWgQ9wPpnc@containers-us-west-99.railway.app:6589/railway"
			connectionString := "user=postgres password=hBcbrWa3PQhWgQ9wPpnc host=containers-us-west-99.railway.app port=6589 dbname=railway sslmode=disable"
			db, err := sql.Open("postgres", connectionString)
			if err != nil {
				log.Printf("Error in connecting to the database: %v \n", err)
			}
			defer db.Close()

			_, err = db.Exec("INSERT INTO purple_data (username,person,purple) VALUES ($1, $2, $3)", username, person, message)
			if err != nil {
				log.Println(err)
				return
			}

			log.Printf("Inserted: %s - %s\n", username, message)

			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

// ListenToWsChannel is a goroutine that waits for an entry on the wsChan, and handles it according to the
// specified action
func ListenToWsChannel() {
	log.Println("Entered in ListenToWsChannel")
	var response WsJsonResponse

	for {
		e := <-wsChan

		switch e.Action {
		case "username":
			// get a list of all users and send it back via broadcast
			clients[e.Conn] = e.Username
			users := getUserList()
			response.Action = "list_users"
			response.ConnectedUsers = users
			broadcastToAll(response)

		case "left":
			// handle the situation where a user leaves the page
			response.Action = "list_users"
			delete(clients, e.Conn)
			users := getUserList()
			response.ConnectedUsers = users
			broadcastToAll(response)

		case "broadcast":
			response.Action = "broadcast"
			response.Person = e.Person
			response.Username = e.Username
			response.Message = e.Message
			broadcastToAll(response)
		}
	}
}

// getUserList returns a slice of strings containing all usernames who are currently online
func getUserList() []string {
	var userList []string
	for _, x := range clients {
		if x != "" {
			userList = append(userList, x)
		}
	}
	return userList
}

// broadcastToAll sends ws response to all connected clients
func broadcastToAll(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			// the user probably left the page, or their connection dropped
			log.Println("websocket err")
			_ = client.Close()
			delete(clients, client)
		}
	}
}

// renderPage renders a jet template
func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
