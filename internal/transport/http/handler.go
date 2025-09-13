package http

import (
	"Order-tracker-service/internal/domain"
	"Order-tracker-service/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler представляет HTTP хэндлер
type Handler struct {
	orderService *service.OrderService
}

// NewHandler создает новый экземпляр HTTP хэндлера
func NewHandler(orderService *service.OrderService) *Handler {
	return &Handler{
		orderService: orderService,
	}
}

// GetOrder обрабатывает GET запрос для получения заказа по ID
func (h *Handler) GetOrder(c *gin.Context) {
	orderUID := c.Param("id")
	if orderUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Order ID is required",
		})
		return
	}

	// Получаем заказ из сервиса
	order, err := h.orderService.GetInfo(orderUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get order",
		})
		return
	}

	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Order not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order": order,
	})
}

// Index обрабатывает GET запрос для главной страницы
func (h *Handler) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":   "Order Tracker Service",
		"message": "Welcome to Order Tracker Service",
	})
}

// HealthCheck обрабатывает GET запрос для проверки здоровья сервиса
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "order-tracker-service",
	})
}

// GetAllOrders обрабатывает GET запрос для получения всех заказов
func (h *Handler) GetAllOrders(c *gin.Context) {
	// Получаем параметры пагинации
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	// В реальном приложении здесь была бы логика пагинации
	// Пока возвращаем простой ответ
	c.JSON(http.StatusOK, gin.H{
		"message": "Get all orders endpoint",
		"page":    page,
		"limit":   limit,
		"orders":  []domain.Order{},
	})
}

// InitRoutes инициализирует маршруты
func (h *Handler) InitRoutes() *gin.Engine {
	// Создаем Gin роутер
	r := gin.Default()

	// Middleware для CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Middleware для логирования
	r.Use(gin.Logger())

	// Middleware для восстановления после паники
	r.Use(gin.Recovery())

	// Статические файлы
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/static/*.html")

	// API маршруты
	api := r.Group("/api/v1")
	{
		// Health check
		api.GET("/health", h.HealthCheck)

		// Заказы
		api.GET("/orders", h.GetAllOrders)
		api.GET("/orders/:id", h.GetOrder)
	}

	// Главная страница
	r.GET("/", h.Index)

	return r
}
