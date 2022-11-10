package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChan = make(chan WsPayload)

var clients = make(map[WebSocketConnection]string)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("../../html"),
	jet.InDevelopmentMode(),
)

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Render Home Page
func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

type WebSocketConnection struct {
	*websocket.Conn
}

// Define response send back from web socket
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

type WsPayload struct {
	Action   string              `json:"action"`
	Username string              `json:"username"`
	Message  string              `json:"message"`
	Conn     WebSocketConnection `json:"-"`
}

// Upgrade connection to websocket
func WsEndPoint(w http.ResponseWriter, r *http.Request) {

	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client Connect to endpoint")

	var respose WsJsonResponse
	respose.Message = `<em><small>Connect to server</small></em>`

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""
	err = ws.WriteJSON(respose)
	if err != nil {
		log.Println(err)
	}
	go ListenForWs(&conn)
}

func ListenForWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("error", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload
	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// lakukan sesuatu
			break
		} else {
			payload.Conn = *conn
			wsChan <- payload

		}
	}
}

func ListenToWsChannel() {
	var respose WsJsonResponse
	for {
		e := <-wsChan
		switch e.Action {
		case "username":
			// get a list of all users and send it back via broadcast
			clients[e.Conn] = e.Username
			users := getUserList()
			respose.Action = "list_users"
			respose.ConnectedUsers = users
			broadCastToAll(respose)
		case "left":
			respose.Action = "list_users"
			delete(clients, e.Conn)
			users := getUserList()
			respose.ConnectedUsers = users
			broadCastToAll(respose)

		case "broadcast":
			respose.Action = "broadcast"
			respose.Message = fmt.Sprintf("<strong>%s</strong>:%s", e.Username, e.Message)
			broadCastToAll(respose)
		}

		// respose.Action = "got here"
		// respose.Message = fmt.Sprintf("some message and action %s", e.Action)
		// broadCastToAll(respose)
	}
}

func getUserList() []string {
	var userlist []string
	for _, x := range clients {
		if x != "" {
			userlist = append(userlist, x)
		}
	}
	sort.Strings(userlist)
	return userlist
}

func broadCastToAll(respose WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(respose)
		if err != nil {
			log.Println("web socket error ")
			_ = client.Close()
			delete(clients, client)
		}
	}
}

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
