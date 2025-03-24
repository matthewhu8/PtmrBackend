DB_URL=postgresql://postgres:password@localhost:5432/ptmrpostgres?sslmode=disable
RDS_URL=postgresql://postgres:lhdzS3Q95R5AyUYJAR30@ptmr-test-postgres.c5oikaswq0wm.us-east-2.rds.amazonaws.com:5432/ptmrpostgres
dockerdown:
	docker compose -f docker-compose.db.yml down

createdb:
	docker exec -it postgres createdb --username=postgres ptmrpostgres

dropdb:
	docker exec -it postgres dropdb --username=postgres ptmrpostgres

mock:
	mockgen -package mockdb -destination pkg/db/mock/store.go github.com/hankimmy/PtmrBackend/pkg/db/sqlc Store
	mockgen -package mockwk -destination pkg/worker/mock/distributor.go github.com/hankimmy/PtmrBackend/pkg/worker TaskDistributor
	mockgen -package mockes -destination pkg/elasticsearch/mock/elasticsearch_mock.go github.com/hankimmy/PtmrBackend/pkg/elasticsearch ESClient
	mockgen -package mockgapi -destination pkg/google/mock/google_mock.go github.com/hankimmy/PtmrBackend/pkg/google GAPI
	mockgen -package mockfb -destination pkg/firebase/mock/firebase_mock.go github.com/hankimmy/PtmrBackend/pkg/firebase AuthClientFirebase

migrateup:
	migrate -path pkg/db/migration -database "$(DB_URL)"  -verbose up

migrateupRDS:
	migrate -path pkg/db/migration -database "$(RDS_URL)"  -verbose up


migratedown:
	migrate -path pkg/db/migration -database "$(DB_URL)" -verbose down

sqlc:
	sqlc generate

new_migration:
	migrate create -ext sql -dir pkg/db/migration -seq $(name)

test:
	SKIP_API_CALLS=true go test -v -cover -short ./...
	cd cmd/UserService; go test -v -cover -short ./...;
	cd cmd/ApplicationService; go test -v -cover -short ./...;
	cd cmd/JobWriter; go test -v -cover -short ./...;
	cd cmd/MatchingService;  go test -v -cover -short ./...;

update:
	@echo "Updating Go modules in UserService..."
	cd cmd/UserService && go get -u ./...
	@echo "Updating Go modules in ApplicationService..."
	cd cmd/ApplicationService && go get -u ./...
	@echo "Updating Go modules in JobWriter..."
	cd cmd/JobWriter && go get -u ./...
	@echo "All modules updated successfully."
	@echo "Updating Go modules in MatchingService..."
	cd cmd/MatchingService && go get -u ./...
	@echo "All modules updated successfully."

.PHONY: migrateup, migratedown, mock, sqlc, createdb, dropdb, new_migration, test, compose, update, dockerdown, migrateupRDS