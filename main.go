package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/olahol/melody.v1"
	"net/http"
)

func main() {
	r := gin.Default()
	r.Static("/statics", "./statics")
	r.LoadHTMLGlob("./templates/*.html") // load html template
	m := melody.New()

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

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		fmt.Println(string(msg))
		_ = m.Broadcast(msg)
	})

	_ = r.Run(":2000")
}
