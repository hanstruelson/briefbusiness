#!/bin/bash
# Exit immediately if a command exits with a non-zero status
set -e

echo "Starting Legal Appeals Draft Manager..."

# Start backend Go server
echo "Starting Go API Server on http://localhost:8085..."
cd server
go run main.go &
BACKEND_PID=$!
cd ..

# Start frontend dev server
echo "Starting Vite Dev Server..."
cd client
npm run dev &
FRONTEND_PID=$!
cd ..

# Trap SIGINT, SIGTERM, and EXIT to stop both servers when Ctrl+C is pressed
cleanup() {
  echo "Shutting down servers..."
  kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
}
trap cleanup INT TERM EXIT

# Wait for background processes to finish
wait $BACKEND_PID $FRONTEND_PID
