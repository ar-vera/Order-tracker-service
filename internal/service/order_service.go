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
	// Ограничиваем размер кэша простым порогом по количеству ключей
	return len(s.cache) < 100
}

func (s *OrderService) GetInfo(orderUID string) (*domain.Order, error) {
	// 1) Проверяем кэш под RLock
	s.mu.RLock()
	orderFromCache, found := s.cache[orderUID]
	s.mu.RUnlock()
	if found && orderFromCache != nil {
		return orderFromCache, nil
	}

	// 2) Если в кэше нет — читаем из БД
	orderFromDB, err := s.repo.GetById(orderUID)
	if err != nil {
		return nil, err
	}
	if orderFromDB == nil {
		return nil, nil
	}

	// 3) Кладём в кэш для последующих запросов
	if s.CheckCache() {
		s.mu.Lock()
		s.cache[orderUID] = orderFromDB
		s.mu.Unlock()
	}

	return orderFromDB, nil
}

func (s *OrderService) Create(order *domain.Order) error {

	if err := s.repo.Create(order); err != nil {
		return err
	}
	if s.CheckCache() {
		s.mu.Lock()
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
