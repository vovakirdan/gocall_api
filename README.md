# GoCall API Documentation

## 1. Overview
GoCall is a backend API written in Go (Golang) for managing users, friendships, and rooms (with various room types and invitation flows). It uses SQLite for data storage and JWT for authentication.

## 2. Key Features
1. **User Management**: Register, login, refresh tokens, and basic user searches.  
2. **Friends**:  
   - Add/remove friends directly or via a friend request workflow.  
   - Fetch your friend list with basic status.  
3. **Rooms**:  
   - Public or non-public (private/secret) rooms, which may or may not require a password.  
   - Creator or appointed admins can manage room settings.  
   - Invite registered users to a room with an invite workflow (pending/accepted/declined).  
4. **JWT-Based Authentication**: Secured routes require a valid token.

## 3. Database Structure
All tables are managed by GORM with SQLite as the underlying database.

### 3.1 `users` Table
| Field         | Type      | Description                               |
|---------------|-----------|-------------------------------------------|
| **ID**        | `uint`    | Auto-increment primary key (internal)     |
| **UserID**    | `string`  | UUID (unique) for external references     |
| **Username**  | `string`  | Unique username                           |
| **PasswordHash** | `string`| Hashed password                          |
| **Name**      | `text`    | Optional display name                     |
| **Email**     | `text`    | Optional email address                    |
| **IsOnline**  | `bool`    | Stub to indicate if user is online        |
| **CreatedAt** | `time.Time`| Auto-created timestamp                   |

### 3.2 `friends` Table
| Field         | Type      | Description                                     |
|---------------|-----------|-------------------------------------------------|
| **ID**        | `uint`    | Auto-increment primary key                      |
| **UserID**    | `string`  | UUID of one user                               |
| **FriendID**  | `string`  | UUID of the other user                          |
| **CreatedAt** | `time.Time` | Timestamp of when they became friends        |

\> Each friendship is stored in two rows (A→B and B→A).

### 3.3 `friend_requests` Table
| Field         | Type        | Description                                 |
|---------------|-------------|---------------------------------------------|
| **ID**        | `uint`      | Auto-increment primary key                  |
| **FromUserID**| `string`    | UUID of the user sending the request       |
| **ToUserID**  | `string`    | UUID of the user receiving the request      |
| **Status**    | `string`    | `"pending"`, `"accepted"`, or `"declined"`  |
| **CreatedAt** | `time.Time` | Timestamp                                   |

### 3.4 `rooms` Table
| Field         | Type        | Description                                                                  |
|---------------|-------------|------------------------------------------------------------------------------|
| **ID**        | `uint`      | Auto-increment primary key                                                   |
| **RoomID**    | `string`    | UUID (unique) identifying the room                                           |
| **UserID**    | `string`    | UUID of the creator                                                          |
| **Name**      | `string`    | Name/title of the room                                                       |
| **Type**      | `string`    | `"public"`, `"private"`, or `"secret"`                                       |
| **Password**  | `string`    | Optional password (if the room is password-protected)                        |
| **CreatedAt** | `time.Time` | Timestamp of creation                                                        |

### 3.5 `room_members` Table
| Field         | Type        | Description                                               |
|---------------|-------------|-----------------------------------------------------------|
| **ID**        | `uint`      | Auto-increment primary key                                |
| **RoomID**    | `string`    | UUID of the room                                         |
| **UserID**    | `string`    | UUID of the user                                         |
| **Role**      | `string`    | `"creator"`, `"admin"`, or `"member"`                    |
| **JoinedAt**  | `time.Time` | Timestamp of when the user joined                        |

### 3.6 `room_invites` Table
| Field             | Type        | Description                                                         |
|-------------------|-------------|---------------------------------------------------------------------|
| **ID**            | `uint`      | Auto-increment primary key                                          |
| **RoomID**        | `string`    | UUID of the room being invited to                                   |
| **InviterUserID** | `string`    | UUID of the user sending the room invite                            |
| **InvitedUserID** | `string`    | UUID of the user receiving the room invite                          |
| **Status**        | `string`    | `"pending"`, `"accepted"`, or `"declined"`                          |
| **CreatedAt**     | `time.Time` | Timestamp of when the invitation was created                        |

## 4. Endpoints Reference

All endpoints are grouped under `/api/`.  
Public endpoints do **not** require authorization.  
Protected endpoints require a valid JWT in the `Authorization: Bearer <token>` header.

### 4.1 Authentication
- **POST /api/auth/register**  
  Register a new user with `username` and `password`.  
  ```json
  {
    "username": "john",
    "password": "secret123"
  }
  ```
- **POST /api/auth/login**  
  Login and receive a JWT.  
- **POST /api/auth/refresh**  
  Refresh the JWT token if near expiration (requires existing valid token).

### 4.2 Users
- **GET /api/user/id** (Protected)  
  Returns the authenticated user’s UUID.  
  Example response:
  ```json
  { "userID": "UUID-OF-USER" }
  ```
- **GET /api/friends/search** (Protected)  
  Query users by `q` param.  
  Example: `/api/friends/search?q=jo`  
  Returns array of matching users.

### 4.3 Friends
- **GET /api/friends** (Protected)  
  Get a list of your friends with a basic online status.  
- **POST /api/friends/add** (Protected)  
  Directly add a friend by username (bypasses friend request flow).  
  ```json
  { "friend_username": "jane" }
  ```
- **DELETE /api/friends/remove** (Protected)  
  Remove (unfriend) by username.  
  ```json
  { "friend_username": "jane" }
  ```
- **POST /api/friends/request** (Protected)  
  Send a friend request.  
  ```json
  { "to_username": "jane" }
  ```
- **POST /api/friends/accept** (Protected)  
  Accept a friend request.  
  ```json
  { "request_id": 1 }
  ```
- **POST /api/friends/decline** (Protected)  
  Decline a friend request.  
  ```json
  { "request_id": 1 }
  ```

### 4.4 Rooms
- **GET /api/rooms/:id/exists** (Public)  
  Check if a room with the given UUID exists. Used by SFU to verify the room.  
- **GET /api/rooms/public** (Public)  
  List all public rooms.  
- **GET /api/rooms/:id** (Public)  
  - If room is `"public"`, returns room data without auth.  
  - If room is non-public, user must be a member (requires JWT).  

#### Protected (JWT Required)
- **GET /api/rooms/mine**  
  Returns rooms where you are the creator.  
- **POST /api/rooms/create**  
  Create a new room.  
  ```json
  {
    "name": "My Room",
    "type": "public",    // or private or secret
    "password": ""       // optional
  }
  ```
- **PUT /api/rooms/:id**  
  Update the room’s `name`, `type`, and optional `password`.  
  Only creator/admin can do this.  
- **DELETE /api/rooms/:id**  
  Delete the room entirely (only creator).  
- **POST /api/rooms/:id/make-admin**  
  Creator can appoint an existing member as admin.  
  ```json
  { "user_id": "<USER-UUID>" }
  ```
  
### 4.5 Room Invites
- **POST /api/rooms/invite**  
  Invite a registered user (by username) to a room. Only creator/admin can invite.  
  ```json
  {
    "roomID": "<ROOM-UUID>",
    "username": "jane"
  }
  ```
- **GET /api/rooms/invites**  
  Returns pending/accepted invites for the authenticated user.  
- **POST /api/rooms/invite/accept**  
  Accept a room invitation.  
  ```json
  { "invite_id": 123 }
  ```
- **POST /api/rooms/invite/decline**  
  Decline a room invitation.  
  ```json
  { "invite_id": 123 }
  ```

## 5. Usage Examples

1. **Register** → **Login** → **Get JWT**:
   - `POST /api/auth/register` with `{ "username": "john", "password": "test123" }`
   - `POST /api/auth/login` with same credentials → returns `{ "token": "...jwt..." }`
   - Use the returned token in all protected routes:  
     `Authorization: Bearer <jwt>`

2. **Add a Friend**:
   - `POST /api/friends/add` with `{ "friend_username": "jane" }` in JSON body.

3. **Create a Room**:
   - `POST /api/rooms/create` with body:
     ```json
     {
       "name": "Chill Zone",
       "type": "public",
       "password": ""
     }
     ```
   - Returns `roomID`, which is a UUID.

4. **Invite a User to a Room**:
   - `POST /api/rooms/invite` with:
     ```json
     {
       "roomID": "<ROOM-UUID>",
       "username": "jane"
     }
     ```

5. **Accept a Room Invite**:
   - `POST /api/rooms/invite/accept` with `{ "invite_id": 123 }`

## 6. Relationship Explanation
- **Users**: Uniquely identified by a UUID (`UserID`).  
- **Friends**: A two-way relationship stored in the `friends` table. Optionally established by a friend request.  
- **Rooms**:  
  - Each has a UUID (`RoomID`).  
  - Creator: The user who made the room (`UserID`).  
  - `RoomMember` table tracks membership and role (`creator`, `admin`, or `member`).  
  - Public vs. private/secret determines whether non-members can view/join.  
- **Room Invites**: Mirror the friend request flow. Pending, accepted, or declined.

## 7. Secret key
You can also generate secret key manually (via python):
```python
import secrets
with open(".env", "a") as f:
    f.write(f"SECRET_KEY={secrets.token_hex(32)}")
```
Or just run the [script](generate_secret_key.py)
```bash
python3 generate_secret_key.py
```

## 8. Run test api
> Don't forget to make it executable
```bash
chmod +x test_api.sh
```
Make sure you have installed [jq](https://stedolan.github.io/jq/)
```bash
sudo apt install jq
```
Run tests
```bash
./test_api.sh
```
