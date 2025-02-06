package handlers

import (
	"log"
	"net/http"
	"sync"

	"GoCall_api/db"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// WebSocket-апгрейдер
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { // todo : check origin
		return true
	},
}

// Хранилище для подключений
// key - userID (UUID), value - канал для отправки сообщений
var chatClients = struct {
	sync.RWMutex
	clients map[string]*websocket.Conn
}{
	clients: make(map[string]*websocket.Conn),
}

// HandleChatWebSocket обрабатывает WebSocket-соединение
func HandleChatWebSocket(c *gin.Context) {
	// Из middleware мы получаем int (ID в БД).
	// Найдём пользователя, чтобы получить его UUID (UserID).
	dbUserID := c.MustGet("user_id").(uint)

	// Ищем пользователя в БД
	var user db.User
	if err := db.DB.First(&user, dbUserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in DB"})
		return
	}

	// Апгрейд соединения до WebSocket
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer wsConn.Close()

	// Сохраняем подключение пользователя в памяти
	chatClients.Lock()
	chatClients.clients[user.UserID] = wsConn
	chatClients.Unlock()
	log.Printf("User %s connected to chat\n", user.UserID)

	for {
		// Ожидаем JSON-объект вида: {"to": "...", "message": "..."}
		var incoming struct {
			To      string `json:"to"`
			Message string `json:"message"`
		}

		// Считываем JSON
		if err := wsConn.ReadJSON(&incoming); err != nil {
			log.Println("ReadJSON error:", err)
			break
		}

		// Проверяем, что есть получатель
		if incoming.To == "" {
			log.Println("No recipient specified")
			continue
		}

		// (Опционально) проверяем, являются ли пользователи друзьями
		// Если надо ограничить общение только друзьям — раскомментируйте:
		if !areFriends(user.UserID, incoming.To) {
			log.Printf("Users %s and %s are not friends. Message blocked.\n", user.UserID, incoming.To)
			continue
		}

		// Сохраняем сообщение в БД
		newMsg := db.Message{
			SenderID:   user.UserID,
			ReceiverID: incoming.To,
			Text:       incoming.Message,
		}
		if err := db.DB.Create(&newMsg).Error; err != nil {
			log.Println("DB Create error:", err)
			continue
		}

		// Пытаемся найти подключение получателя
		chatClients.RLock()
		receiverConn, ok := chatClients.clients[incoming.To]
		chatClients.RUnlock()

		// Если получатель в сети, отправляем сообщение
		if ok && receiverConn != nil {
			var outgoing = struct {
				From    string `json:"from"`
				To      string `json:"to"`
				Message string `json:"message"`
			}{
				From:    user.UserID,
				To:      incoming.To,
				Message: incoming.Message,
			}
			if err := receiverConn.WriteJSON(outgoing); err != nil {
				log.Println("WriteJSON error:", err)
			}
		} else {
			// Иначе пользователь офлайн — сообщение уже сохранено в БД
			log.Printf("User %s is offline. Message stored.\n", incoming.To)
		}
	}

	// Удаляем подключение при разрыве
	chatClients.Lock()
	delete(chatClients.clients, user.UserID)
	chatClients.Unlock()
	log.Printf("User %s disconnected\n", user.UserID)
}

// Пример функции проверки дружбы
func areFriends(user1UUID, user2UUID string) bool {
	var count int64
	// Ищем пару записей Friend в БД
	// Friend.UserID, Friend.FriendID — это userUUID (string), а не ID (uint)
	if err := db.DB.
		Model(&db.Friend{}).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			user1UUID, user2UUID, user2UUID, user1UUID).
		Count(&count).Error; err != nil {
		if err != gorm.ErrRecordNotFound && err != nil {
			log.Println("DB error while checking friendship:", err)
			return false
		}
	}
	return count > 0
}
