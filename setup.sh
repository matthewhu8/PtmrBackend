#!/bin/bash

# Run Docker Compose
docker compose -f docker-compose.db.yml up -d

# Wait for Elasticsearch to start up
echo "Waiting for Elasticsearch to start..."
sleep 10
until curl -s http://localhost:9200/_cluster/health | grep -q '"status":"green"'; do
  echo "Waiting for Elasticsearch to be healthy..."
  sleep 5
done

JOB_INDEX_MAPPING='{
  "mappings": {
    "properties": {
      "date_posted": {
        "type": "date",
        "format": "yyyy-MMMM-dd"
      },
      "job_location": { "type": "keyword" },
      "employment_type": { "type": "keyword" },
      "title": { "type": "text" },
      "industry": { "type": "keyword" },
      "wage": { "type": "half_float" },
      "user_created": { "type": "boolean" },
      "rating": { "type": "half_float" },
      "price_level": { "type": "keyword" },
      "requirements": { "type": "keyword" },
      "opening_hours": {
        "type": "nested",
        "properties": {
          "close": {
            "type": "object",
            "properties": {
              "day": { "type": "integer" },
              "hour": { "type": "integer" },
              "minute": { "type": "integer" }
            }
          },
          "open": {
            "type": "object",
            "properties": {
              "day": { "type": "integer" },
              "hour": { "type": "integer" },
              "minute": { "type": "integer" }
            }
          }
        }
      },
      "precise_location": { "type": "geo_point" }
    }
  }
}'


CANDIDATE_INDEX_MAPPING='
{
  "mappings": {
    "properties": {
      "education": { "type": "keyword" },
      "location": { "type": "text" },
      "skill_set": { "type": "keyword" },
      "certificates": { "type": "keyword" },
      "time_availability": {
        "type": "nested",
        "properties": {
          "morning": { "type": "boolean" },
          "afternoon": { "type": "boolean" },
          "evening": { "type": "boolean" },
          "night": { "type": "boolean" }
        }
      },
      "account_verified": { "type": "boolean" },
      "has_resume": {"type": "boolean" },
      "rating": { "type": "integer" },
      "past_experience": {
        "type": "nested",
        "properties": {
          "industry": { "type": "keyword" },
          "job_title": { "type": "text" },
          "length": { "type": "half_float" }
        }
      }
    }
  }
}
'

APPLICATION_INDEX_MAPPING='
{
  "mappings": {
    "properties": {
      "any_field": {
        "type": "text",
        "index": false  // Disable indexing
      }
    }
  }
}
'
# Delete the index if it exists
JOB_INDEX_EXISTS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:9200/jobs")
CANDIDATE_INDEX_EXISTS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:9200/candidates")
CANDIDATE_APPLICATIONS_INDEX_EXISTS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:9200/candidate_applications")
EMPLOYER_APPLICATIONS_INDEX_EXISTS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:9200/employer_applications")


if [ "$JOB_INDEX_EXISTS" -eq 200 ]; then
  echo "Index 'jobs' exists. Deleting index..."
  curl -X DELETE "http://localhost:9200/jobs"
  echo "Index 'jobs' deleted."
fi
if [ "$CANDIDATE_INDEX_EXISTS" -eq 200 ]; then
  echo "Index 'candidates' exists. Deleting index..."
  curl -X DELETE "http://localhost:9200/candidates"
  echo "Index 'candidates' deleted."
fi

if [ "$CANDIDATE_APPLICATIONS_INDEX_EXISTS" -eq 200 ]; then
  echo "Index 'candidate_applications' exists. Deleting index..."
  curl -X DELETE "http://localhost:9200/candidate_applications"
  echo "Index 'candidate_applications' deleted."
fi

if [ "$EMPLOYER_APPLICATIONS_INDEX_EXISTS" -eq 200 ]; then
  echo "Index 'employer_applications' exists. Deleting index..."
  curl -X DELETE "http://localhost:9200/employer_applications"
  echo "Index 'employer_applications' deleted."
fi


curl -X PUT "http://localhost:9200/jobs" -H 'Content-Type: application/json' -d "$JOB_INDEX_MAPPING"
echo "Index 'jobs' created with specified mappings."

curl -X PUT "http://localhost:9200/candidates" -H 'Content-Type: application/json' -d "$CANDIDATE_INDEX_MAPPING"
echo "Index 'candidates' created with specified mappings."

curl -X PUT "http://localhost:9200/candidate_applications" -H 'Content-Type: application/json' -d "$APPLICATION_INDEX_MAPPING"
echo "Index 'candidate_applications' created with specified mappings."

curl -X PUT "http://localhost:9200/employer_applications" -H 'Content-Type: application/json' -d "$APPLICATION_INDEX_MAPPING"
echo "Index 'employer_applications' created with specified mappings."

# Wait for the PostgreSQL service to be available
until docker exec postgres pg_isready -U postgres; do
  echo "Waiting for PostgreSQL to be available..."
  sleep 2
done

migrate -path pkg/db/migration -database "postgresql://postgres:password@localhost:5432/ptmrpostgres?sslmode=disable" -verbose up
