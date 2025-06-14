package observer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"log"
	"net/http"
	"time"
)

type Observer struct {
	natsConn  *nats.Conn
	jwtSecret string
}

func NewObserver(natsConn *nats.Conn, jwtSecret string) *Observer {
	return &Observer{
		natsConn:  natsConn,
		jwtSecret: jwtSecret,
	}
}

type CustomClaims struct {
	UserId    int32 `json:"UserId"`
	ContestId int32 `json:"ContestId"`
	Role      Role  `json:"Role"`
	jwt.RegisteredClaims
}

type SolutionsListItem struct {
	Id int32 `json:"id"`

	UserId   int32  `json:"user_id"`
	Username string `json:"username"`

	State      int32 `json:"state"`
	Score      int32 `json:"score"`
	Penalty    int32 `json:"penalty"`
	TimeStat   int32 `json:"time_stat"`
	MemoryStat int32 `json:"memory_stat"`
	Language   int32 `json:"language"`

	ProblemId    int32  `json:"problem_id"`
	ProblemTitle string `json:"problem_title"`

	Position int32 `json:"position"`

	ContestId    int32  `json:"contest_id"`
	ContestTitle string `json:"contest_title"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	MessageType string            `json:"message_type"`
	Message     *string           `json:"message,omitempty"`
	Solution    SolutionsListItem `json:"solution"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (o *Observer) ListSolutionsWS(w http.ResponseWriter, r *http.Request) {
	claims, err := parseToken(r.URL.Query().Get("token"), o.jwtSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	subject := fmt.Sprintf("contest-%d-solutions", claims.ContestId)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	defer ws.Close()

	_, err = o.natsConn.Subscribe(subject, func(natsMsg *nats.Msg) {
		var msg Message

		if err := json.Unmarshal(natsMsg.Data, &msg); err != nil {
			log.Printf("unmarshal error: %v", err)
			return
		}

		if !hasAccess(claims, &msg) {
			log.Printf("user has no access")
			return
		}

		err := ws.WriteJSON(msg)
		if err != nil {
			log.Printf("write error: %v", err)
			return
		}
	})

	if err != nil {
		log.Printf("subscribe error: %v", err)
	}

	for {
		_, _, err = ws.ReadMessage()
		if err != nil {
			log.Printf("read error: %v", err)
			break
		}
	}
}

func hasAccess(claims *CustomClaims, msg *Message) bool {
	if claims.ContestId != msg.Solution.ContestId {
		return false
	}

	if claims.Role == RoleStudent {
		return claims.UserId == msg.Solution.UserId
	}

	if claims.Role == RoleAdmin || claims.Role == RoleTeacher {
		return true
	}

	return false
}

func parseToken(tokenStr, secret string) (*CustomClaims, error) {
	if tokenStr == "" {
		return nil, errors.New("missing token")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
