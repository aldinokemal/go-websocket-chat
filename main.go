package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/olahol/melody.v1"
	"net/http"
	"sync"
)

type ResponseWs struct {
	Channel string      `json:"channel"`
	Event   string      `json:"event"`
	Message interface{} `json:"message"`
}

type UserInfo struct {
	Name string `json:"name"`
}

func main() {
	r := gin.Default()
	r.Static("/statics", "./statics")
	r.LoadHTMLGlob("./templates/*.html") // load html template

	m := melody.New()
	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	dataUsers := make(map[*melody.Session]*UserInfo)
	lock := new(sync.Mutex)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "chat.html", nil)
	})

	r.GET("/ws", func(c *gin.Context) {
		_ = m.HandleRequest(c.Writer, c.Request)
	})

	r.GET("/uuid", func(c *gin.Context) {
		token, _ := uuid.NewUUID()
		result := map[string]interface{}{
			"uuid": token.String(),
		}
		c.JSON(http.StatusOK, result)
	})

	m.HandleConnect(func(s *melody.Session) {
		lock.Lock()
		name := s.Request.URL.Query().Get("name")
		newUser := UserInfo{
			Name: name,
		}
		s.Set("data", map[string]interface{}{"name": name})
		dataUsers[s] = &newUser
		var namaUsers []string

		for _, key := range dataUsers {
			namaUsers = append(namaUsers, key.Name)
		}

		message := map[string]interface{}{
			"info":  name + " join chatroom",
			"total": m.Len() + 1,
			"users": namaUsers,
		}
		lock.Unlock()

		data := ResponseWs{
			Channel: "chatroom",
			Event:   "status",
			Message: message,
		}
		b, _ := json.Marshal(data)
		_ = m.Broadcast(b)
	})

	m.HandleDisconnect(func(s *melody.Session) {
		lock.Lock()
		delete(dataUsers, s)

		var namaUsers []string

		for _, key := range dataUsers {
			namaUsers = append(namaUsers, key.Name)
		}

		dataSession := s.Keys["data"].(map[string]interface{})
		name := dataSession["name"].(string)
		message := map[string]interface{}{
			"info":  name + " left chatroom",
			"total": m.Len() - 1,
			"users": namaUsers,
		}
		lock.Unlock()

		data := ResponseWs{
			Channel: "chatroom",
			Event:   "status",
			Message: message,
		}
		b, _ := json.Marshal(data)
		_ = m.Broadcast(b)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		data := ResponseWs{
			Channel: "chatroom",
			Event:   "message",
			Message: string(msg),
		}
		b, _ := json.Marshal(data)
		_ = m.Broadcast(b)
	})

	_ = r.Run(":2000")
}
