package ports

//go:generate go run go.uber.org/mock/mockgen -destination=../mocks/mock_store.go -package=mocks github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports Store
