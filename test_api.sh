#!/bin/bash

# vars
BASE_URL="http://localhost:8080/api"
TOKEN=""
# for another tests
USER_ID=""

divider() {
    echo "============================="
}

# Test register user
test_register() {
    divider
    echo "Testing: User Registration"
    RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
        -H "Content-Type: application/json" \
        -d '{"username":"testuser","password":"testpass"}')
    echo "$RESPONSE"
}

# Test auth
test_login() {
    divider
    echo "Testing: User Login"
    RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"testuser","password":"testpass"}')
    echo "$RESPONSE"
    
    # Извлекаем токен из ответа
    TOKEN=$(echo "$RESPONSE" | jq -r '.token')
    echo "Extracted Token: $TOKEN"
}

# Test room creation
test_create_room() {
    divider
    echo "Testing: Create Room"
    RESPONSE=$(curl -s -X POST "$BASE_URL/rooms/create" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"name":"Test Room"}')
    echo "$RESPONSE"
}

# Test list rooms
test_get_rooms() {
    divider
    echo "Testing: Get Rooms"
    RESPONSE=$(curl -s -X GET "$BASE_URL/rooms" \
        -H "Authorization: Bearer $TOKEN")
    echo "$RESPONSE"
}

# Test removing room
test_delete_room() {
    divider
    echo "Testing: Delete Room"
    ROOM_ID=1  # Замените на ID существующей комнаты
    RESPONSE=$(curl -s -X DELETE "$BASE_URL/rooms/$ROOM_ID" \
        -H "Authorization: Bearer $TOKEN")
    echo "$RESPONSE"
}

# run tests
test_register
test_login
test_create_room
test_get_rooms
test_delete_room
