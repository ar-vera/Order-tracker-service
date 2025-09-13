package service

import (
	"Order-tracker-service/internal/domain"
	"Order-tracker-service/internal/repository"
	"context"
	"log"
	"sync"
)

type OrderService struct {
	repo      repository.OrderRepository
	cache     map[string]*domain.Order
	mu        sync.RWMutex
	CacheSize int
}

func NewOrderService(repo repository.OrderRepository, Size int) *OrderService {
	return &OrderService{
		repo:      repo,
		cache:     make(map[string]*domain.Order),
		CacheSize: Size,
	}
}

func (s *OrderService) CheckCache() bool {
	return s.CacheSize <= 100
}

func (s *OrderService) GetInfo(orderUID string) (*domain.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.cache[orderUID]
	if !ok {
		return order, nil
	}
	s.mu.Unlock()

	order, err := s.repo.GetById(orderUID)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) Create(order *domain.Order) error {

	if err := s.repo.Create(order); err != nil {
		return err
	}
	if s.CheckCache() {
		s.mu.Lock()
		if _, ok := s.cache[order.OrderUID]; ok {
			s.CacheSize++
			return nil
		}
		s.cache[order.OrderUID] = order
		s.mu.Unlock()
	}
	return nil
}

func (s *OrderService) RestoreCache(order *domain.Order) error {
	orders, err := s.repo.GetAll()
	if err != nil {
		return err
	}

	s.mu.Lock()
	for _, order := range orders {
		s.cache[order.OrderUID] = order
	}
	s.mu.Unlock()

	log.Printf("restore cache of order %v", len(orders))

	return nil
}

// HandleOrder обрабатывает заказ, полученный из Kafka
func (s *OrderService) HandleOrder(ctx context.Context, order *domain.Order) error {
	log.Printf("Processing order from Kafka: %s", order.OrderUID)

	// Сохраняем заказ в базу данных
	if err := s.Create(order); err != nil {
		log.Printf("Failed to save order %s to database: %v", order.OrderUID, err)
		return err
	}

	log.Printf("Successfully processed order from Kafka: %s", order.OrderUID)
	return nil
}
