package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/olahol/melody.v1"
	"net/http"
	"strings"
	"sync"
)

type ResponseWs struct {
	Channel string      `json:"channel"`
	Event   string      `json:"event"`
	Message interface{} `json:"message"`
}

type UserInfo struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

type FileType struct {
	FileName string `json:"filename"`
	Type     string `json:"type"`
	Url      string `json:"url"`
}

type DataMessage struct {
	Name    string   `json:"name"`
	Message string   `json:"message"`
	Sender  string   `json:"sender"`
	Time    string   `json:"time"`
	File    FileType `json:"file"`
}

func main() {
	r := gin.Default()
	r.Static("/statics", "./statics")
	r.LoadHTMLGlob("./templates/*.html") // load html template

	m := melody.New()
	m.Config.MaxMessageSize = 2000
	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true } // origni check
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
		token, _ := uuid.NewUUID()
		newUser := UserInfo{
			Name: name,
			UUID: token.String(),
		}
		s.Set("data", newUser)
		dataUsers[s] = &newUser
		var namaUsers []UserInfo

		for _, key := range dataUsers {
			namaUsers = append(namaUsers, UserInfo{
				Name: key.Name,
				UUID: key.UUID,
			})
		}

		message := map[string]interface{}{
			"info":   strings.TrimLeft(name, " ") + " join chatroom",
			"total":  m.Len() + 1,
			"target": newUser,
			"users":  namaUsers,
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

		var namaUsers []UserInfo

		for _, key := range dataUsers {
			namaUsers = append(namaUsers, UserInfo{
				Name: key.Name,
				UUID: key.UUID,
			})
		}

		dataSession := s.Keys["data"].(UserInfo)
		name := dataSession.Name
		message := map[string]interface{}{
			"info":   name + " left chatroom",
			"total":  m.Len() - 1,
			"target": dataSession,
			"users":  namaUsers,
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
		var message DataMessage
		if err := json.Unmarshal(msg, &message); err != nil {
			panic(err)
		}
		message.Message = strings.Trim(message.Message, " ")
		message.Message = strings.Trim(message.Message, "\n")
		fmt.Println(message)
		data := ResponseWs{
			Channel: "chatroom",
			Event:   "message",
			Message: message,
		}
		b, _ := json.Marshal(data)
		_ = m.Broadcast(b)
	})

	_ = r.Run(":2000")
}
