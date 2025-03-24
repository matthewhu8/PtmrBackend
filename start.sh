#!/bin/sh

if [ "$SERVICE" = "UserService" ]; then
  exec /app/UserService/main
elif [ "$SERVICE" = "ApplicationService" ]; then
  exec /app/ApplicationService/main
elif [ "$SERVICE" = "MatchingService" ]; then
  exec /app/MatchingService/main
elif [ "$SERVICE" = "JobWriter" ]; then
  exec /app/JobWriter/main
else
  echo "Unknown service: $SERVICE"
  exit 1
fi
