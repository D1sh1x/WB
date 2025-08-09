package handler

import (
	"net/http"

	"WB2/internal/dto/request"
	"WB2/internal/dto/response"
	"WB2/internal/models"

	"github.com/labstack/echo/v4"
)

func (h *Handler) CreateOrder(c echo.Context) error {
	var req request.CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bind_error",
			Message: err.Error(),
			Code:    http.StatusBadRequest})
	}

	order := req.ToOrderModel()
	if err := h.storage.Db.Create(order).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "create_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError})
	}

	return c.JSON(http.StatusCreated, response.SuccessResponse{
		Success: true,
		Message: "order created",
		Data:    response.ToOrderResponse(order)})
}

func (h *Handler) GetAllOrdres(c echo.Context) error {
	// Сначала пробуем из кэша
	cached := h.kafka.GetAllFromCache()
	if len(cached) > 0 {
		resp := response.ToOrderResponseList(cached, len(cached), 1, len(cached))
		return c.JSON(http.StatusOK, resp)
	}
	// Фолбек в БД
	var orders []models.Order
	if err := h.storage.Db.Preload("Delivery").Preload("Payment").Preload("Items").Find(&orders).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "list_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError})
	}
	resp := response.ToOrderResponseList(orders, len(orders), 1, len(orders))
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetOrderByID(c echo.Context) error {
	orderUID := c.Param("id")
	if o, ok := h.kafka.GetOrderFromCache(orderUID); ok {
		return c.JSON(http.StatusOK, response.ToOrderResponse(o))
	}
	var order models.Order
	if err := h.storage.Db.Preload("Delivery").Preload("Payment").Preload("Items").Where("order_uid = ?", orderUID).First(&order).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "not_found",
			Message: "order not found",
			Code:    http.StatusNotFound})
	}
	return c.JSON(http.StatusOK, response.ToOrderResponse(&order))
}

func (h *Handler) UpdateOrder(c echo.Context) error {
	var req request.UpdateOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bind_error",
			Message: err.Error(),
			Code:    http.StatusBadRequest})
	}

	var order models.Order
	if err := h.storage.Db.Preload("Delivery").Preload("Payment").Preload("Items").Where("order_uid = ?", req.OrderUID).First(&order).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "not_found",
			Message: "order not found",
			Code:    http.StatusNotFound})
	}

	req.UpdateOrderModel(&order)

	// Сохраняем основные поля заказа
	if err := h.storage.Db.Save(&order).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError})
	}
	// Сохраняем связанные сущности
	if err := h.storage.Db.Save(&order.Delivery).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_delivery_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError})
	}
	if err := h.storage.Db.Save(&order.Payment).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_payment_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError})
	}
	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		if err := h.storage.Db.Save(&order.Items[i]).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error:   "update_item_error",
				Message: err.Error(),
				Code:    http.StatusInternalServerError})
		}
	}

	return c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "order updated",
		Data:    response.ToOrderResponse(&order)})
}

func (h *Handler) DeleteOrder(c echo.Context) error {
	var req response.DeleteOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "bind_error",
			Message: err.Error(),
			Code:    http.StatusBadRequest})
	}

	uid, err := h.storage.DeleteOrder(req.OrderUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "delete_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
	}

	return c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "order: " + uid + " deleted",
	})
}
