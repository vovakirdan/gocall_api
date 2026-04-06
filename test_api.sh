#!/bin/bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080/api}"
ORIGIN="${ORIGIN:-http://127.0.0.1:1420}"

node <<'NODE'
const base = process.env.BASE_URL || "http://127.0.0.1:8080/api";
const origin = process.env.ORIGIN || "http://127.0.0.1:1420";
const password = "testpass123";
const now = Date.now();
const suffix = String(now).slice(-8);
const userA = `api${suffix}a`;
const userB = `api${suffix}b`;

async function parseResponse(response) {
  const text = await response.text();
  try {
    return { status: response.status, body: JSON.parse(text) };
  } catch {
    return { status: response.status, body: text };
  }
}

async function request(path, { method = "GET", token, body } = {}) {
  const headers = { Origin: origin };
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(base + path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  const parsed = await parseResponse(response);
  if (parsed.status >= 400) {
    throw new Error(`${method} ${path} failed: ${parsed.status} ${JSON.stringify(parsed.body)}`);
  }
  return parsed.body;
}

async function main() {
  console.log("Registering users...");
  await request("/auth/register", { method: "POST", body: { username: userA, password } });
  await request("/auth/register", { method: "POST", body: { username: userB, password } });

  console.log("Logging in...");
  const loginA = await request("/auth/login", { method: "POST", body: { username: userA, password } });
  const loginB = await request("/auth/login", { method: "POST", body: { username: userB, password } });
  const tokenA = loginA.token;
  const tokenB = loginB.token;

  console.log("Fetching profiles...");
  const meA = await request("/user/me", { token: tokenA });
  const meB = await request("/user/me", { token: tokenB });

  console.log("Sending and accepting friend request...");
  await request("/friends/request", {
    method: "POST",
    token: tokenA,
    body: { to_username: userB },
  });
  const requests = await request("/friends/requests", { token: tokenB });
  const requestId = requests.friend_requests?.[0]?.id;
  if (!requestId) {
    throw new Error("No incoming friend request found for second user");
  }
  await request("/friends/accept", {
    method: "POST",
    token: tokenB,
    body: { request_id: requestId },
  });

  console.log("Creating room and joining voice...");
  const room = await request("/rooms/create", {
    method: "POST",
    token: tokenA,
    body: { name: "API Smoke Room", type: "public" },
  });
  const roomId = room.roomID;

  await request(`/rooms/${roomId}/join`, { method: "POST", token: tokenB });
  await request(`/rooms/${roomId}/voice/join`, { method: "POST", token: tokenA });
  const credentials = await request(`/rooms/${roomId}/voice/credentials`, { token: tokenA });

  console.log("Fetching chat history and conversations...");
  await request(`/chat/history?with_user=${encodeURIComponent(meB.user_id)}`, { token: tokenA });
  await request("/chat/conversations", { token: tokenA });

  console.log(JSON.stringify({
    users: [meA.username, meB.username],
    room_id: roomId,
    livekit_url: credentials.url,
    livekit_room: credentials.room_name,
  }, null, 2));
}

main().catch((error) => {
  console.error(error.message);
  process.exit(1);
});
NODE
