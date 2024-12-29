#!/bin/bash

# vars
BASE_URL="http://localhost:8080/api"
TOKEN=""
ROOM_ID=""
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

    USER_ID=$(echo "$RESPONSE" | jq -r '.userID')
    if [[ -z "$USER_ID" || "$USER_ID" == "null" ]]; then
        echo "Failed to register. USER_ID is empty."
        exit 1
    fi
    echo "Extracted USER_ID: $USER_ID"
}

# Test auth
test_login() {
    divider
    echo "Testing: User Login"
    RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"testuser","password":"testpass"}')
    echo "$RESPONSE"

    TOKEN=$(echo "$RESPONSE" | jq -r '.token')
    if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
        echo "Failed to login. Token is empty."
        exit 1
    fi
    echo "Extracted Token: $TOKEN"
}

# Test room creation
test_create_room() {
    divider
    echo "Testing: Create Room"
    RESPONSE=$(curl -s -X POST "$BASE_URL/rooms/create" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"name":"Test Room", "type":"public", "password":""}')
    echo "$RESPONSE"

    ROOM_ID=$(echo "$RESPONSE" | jq -r '.roomID')
    if [[ -z "$ROOM_ID" || "$ROOM_ID" == "null" ]]; then
        echo "Failed to create room. Room ID is empty."
        exit 1
    fi
    echo "Created Room ID: $ROOM_ID"
}

# Test list rooms
test_get_rooms() {
    divider
    echo "Testing: Get Rooms"
    RESPONSE=$(curl -s -X GET "$BASE_URL/rooms" \
        -H "Authorization: Bearer $TOKEN")
    echo "$RESPONSE"
}

# Test adding user to room
test_add_user_to_room() {
    divider
    echo "Testing: Add User to Room"
    RESPONSE=$(curl -s -X POST "$BASE_URL/rooms/add-user" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"roomID":"'$ROOM_ID'", "userID":"'$USER_ID'", "role":"member"}')
    echo "$RESPONSE"
}

# Test get room members
test_get_room_members() {
    divider
    echo "Testing: Get Room Members"
    RESPONSE=$(curl -s -X GET "$BASE_URL/rooms/$ROOM_ID/members" \
        -H "Authorization: Bearer $TOKEN")
    echo "$RESPONSE"
}

# Test removing room
test_delete_room() {
    divider
    echo "Testing: Delete Room"
    RESPONSE=$(curl -s -X DELETE "$BASE_URL/rooms/$ROOM_ID" \
        -H "Authorization: Bearer $TOKEN")
    echo "$RESPONSE"
}

# run tests
test_register
test_login
test_create_room
test_get_rooms
test_add_user_to_room
test_get_room_members
test_delete_room
