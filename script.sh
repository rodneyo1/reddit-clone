#!/bin/bash

# Variables
IMAGE_NAME="forum-app"
CONTAINER_NAME="forum-container"
PORT=8081

# Google OAuth credentials
# To get these credentials:
# 1. Go to https://console.cloud.google.com/
# 2. Create a new project or select an existing one
# 3. Enable the Google+ API and OAuth consent screen
# 4. Go to Credentials
# 5. Create OAuth 2.0 Client ID
# 6. Add redirect URI: http://localhost:8081/auth/google/callback
# 7. Replace the placeholders below with your actual credentials
GOOGLE_CLIENT_ID="1082051974368-4vrk6abov8eeubo0vrmlrutlho0iupvv.apps.googleusercontent.com"
GOOGLE_CLIENT_SECRET="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"
GITHUB_CLIENT_ID="Ov23liyObIU7dL0P2sI4"
GITHUB_CLIENT_SECRET="2bf1f84645f97e29c2b397ae6b00d0c83384cb9a"

# Export environment variables
export GOOGLE_CLIENT_ID
export GOOGLE_CLIENT_SECRET
export GITHUB_CLIENT_ID
export GITHUB_CLIENT_SECRET

# # Stop and remove any existing container
# docker stop $CONTAINER_NAME 2>/dev/null && docker rm $CONTAINER_NAME 2>/dev/null

# # Prune only unused images and stopped containers (safer approach)
# docker image prune -f
# docker container prune -f

# # Build the Docker image
# docker build -t $IMAGE_NAME . && \
# echo "Docker image built successfully." || \
# { echo "Failed to build Docker image."; exit 1; }

# # Run the Docker container
# docker run -d \
#     --name $CONTAINER_NAME \
#     -p $PORT:8081 \
#     -e GOOGLE_CLIENT_ID=$GOOGLE_CLIENT_ID \
#     -e GOOGLE_CLIENT_SECRET=$GOOGLE_CLIENT_SECRET \
#     $IMAGE_NAME && \
# echo "Docker container is running on port $PORT with Google OAuth configured." || \
# { echo "Failed to run Docker container."; exit 1; }

go run main.go