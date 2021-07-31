package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/olahol/melody.v1"
	"net/http"
	"strings"
	"sync"
)

type WebsocketData struct {
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
	MessageID string   `json:"message_id"`
	Name      string   `json:"name"`
	Message   string   `json:"message"`
	Sender    string   `json:"sender"`
	Time      string   `json:"time"`
	File      FileType `json:"file"`
}

func main() {
	r := gin.Default()
	r.Static("/statics", "./statics")
	r.LoadHTMLGlob("./templates/*") // load html template

	m := melody.New()
	m.Config.MaxMessageSize = 2000
	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true } // origni check
	dataUsers := make(map[*melody.Session]*UserInfo)
	lock := new(sync.Mutex)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "chat.html", nil)
	})

	r.GET("sw.js", func(c *gin.Context) {
		c.Header("Content-Type", "application/javascript")
		c.HTML(http.StatusOK, "sw.js", nil)
	})
	r.GET("/ws", func(c *gin.Context) {
		err := m.HandleRequest(c.Writer, c.Request)
		if err != nil {
			fmt.Println(err.Error())
		}
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
		var token string
		name := s.Request.URL.Query().Get("name")
		token = s.Request.URL.Query().Get("uuid")
		if token == "" {
			t, _ := uuid.NewUUID()
			token = t.String()
		}

		newUser := UserInfo{
			Name: name,
			UUID: token,
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

		data := WebsocketData{
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

		data := WebsocketData{
			Channel: "chatroom",
			Event:   "status",
			Message: message,
		}
		b, _ := json.Marshal(data)
		_ = m.Broadcast(b)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		var WsData WebsocketData
		if err := json.Unmarshal(msg, &WsData); err != nil {
			panic(err)
		}

		if WsData.Channel == "chatroom" {
			if WsData.Event == "send_message_text" || WsData.Event == "send_message_image" {
				var message DataMessage
				_ = mapstructure.Decode(WsData.Message, &message)
				token, _ := uuid.NewUUID()
				message.MessageID = "msg-" + token.String()
				message.Message = strings.Trim(message.Message, " ")
				message.Message = strings.Trim(message.Message, "\n")
				data := WebsocketData{
					Channel: WsData.Channel,
					Event:   WsData.Event,
					Message: message,
				}
				b, _ := json.Marshal(data)
				_ = m.Broadcast(b)
			} else if WsData.Event == "unsend_message" {
				data := WebsocketData{
					Channel: WsData.Channel,
					Event:   WsData.Event,
					Message: WsData.Message,
				}
				b, _ := json.Marshal(data)
				_ = m.Broadcast(b)
			}
		}

	})

	_ = r.Run(":2000")
}
