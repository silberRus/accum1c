package service

import (
	"inventory/internal/config"
	"inventory/internal/repository"
)

type UpdateInventoryRequest struct {
	ProductID string                 `json:"product_id"`
	Fields    map[string]interface{} `json:"fields"`
}

type Service struct {
	repo *repository.Repository
	cfg  *config.Config
}

func NewService(repo *repository.Repository, cfg *config.Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

func (s *Service) UpdateInventory(request UpdateInventoryRequest) error {
	// Передаем все поля из запроса для обновления в Redis
	return s.repo.UpdateInventoryInRedis(request.ProductID, request.Fields)
}

func (s *Service) GetInventory(productID string) (int, error) {
	key := s.cfg.DatabaseStructure[0].Fields[3].Name
	return s.repo.GetInventoryFromRedis(productID, key)
}
