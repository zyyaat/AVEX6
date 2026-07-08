// Package service queries: GetOrder, TrackOrder, List* use cases.
package service

import (
	"context"
	"errors"

	"avex-backend/internal/modules/orders/domain"
	"avex-backend/internal/modules/orders/port"
)

// ===== GetOrder =====

func (s *Service) GetOrder(ctx context.Context, orderID string) (*port.OrderDTO, error) {
	order, items, err := s.loadOrder(ctx, s.pool, orderID)
	if err != nil {
		return nil, err
	}
	dto := toOrderDTO(*order, items)
	return &dto, nil
}

// ===== TrackOrder =====

func (s *Service) TrackOrder(ctx context.Context, orderNumber string) (*port.OrderDTO, error) {
	order, err := s.deps.Repos.Orders.GetByOrderNumber(ctx, s.pool, orderNumber)
	if err != nil {
		return nil, err
	}
	items, err := s.deps.Repos.Items.ListByOrder(ctx, s.pool, order.ID())
	if err != nil {
		return nil, err
	}
	dto := toOrderDTO(*order, items)
	return &dto, nil
}

// ===== ListMyOrders =====

func (s *Service) ListMyOrders(ctx context.Context, userID string, page port.PageQuery) (port.Page[port.OrderDTO], error) {
	ordersPage, err := s.deps.Repos.Orders.ListByUser(ctx, s.pool, userID, page)
	if err != nil {
		return port.Page[port.OrderDTO]{}, err
	}
	return s.enrichOrdersPage(ctx, ordersPage)
}

// ===== ListRestaurantOrders =====

func (s *Service) ListRestaurantOrders(ctx context.Context, restaurantID string, page port.PageQuery) (port.Page[port.OrderDTO], error) {
	ordersPage, err := s.deps.Repos.Orders.ListByRestaurant(ctx, s.pool, restaurantID, page)
	if err != nil {
		return port.Page[port.OrderDTO]{}, err
	}
	return s.enrichOrdersPage(ctx, ordersPage)
}

// ===== ListDriverOrders =====

func (s *Service) ListDriverOrders(ctx context.Context, driverID string, page port.PageQuery) (port.Page[port.OrderDTO], error) {
	ordersPage, err := s.deps.Repos.Orders.ListByDriver(ctx, s.pool, driverID, page)
	if err != nil {
		return port.Page[port.OrderDTO]{}, err
	}
	return s.enrichOrdersPage(ctx, ordersPage)
}

// ===== ListOrdersByStatus =====

func (s *Service) ListOrdersByStatus(ctx context.Context, status string, page port.PageQuery) (port.Page[port.OrderDTO], error) {
	domainStatus, err := domain.ParseOrderStatus(status)
	if err != nil {
		return port.Page[port.OrderDTO]{}, errors.New("invalid status")
	}
	ordersPage, err := s.deps.Repos.Orders.ListByStatus(ctx, s.pool, domainStatus, page)
	if err != nil {
		return port.Page[port.OrderDTO]{}, err
	}
	return s.enrichOrdersPage(ctx, ordersPage)
}

// ===== Helper: enrich a page of orders with items and convert to DTOs =====
//
// Uses ListByOrderIDs (batch query) to avoid N+1 queries.
// For a page of 50 orders, this makes exactly 2 queries:
//   1. List orders (with LIMIT/OFFSET)
//   2. List items for all those orders (WHERE order_id IN (...))

func (s *Service) enrichOrdersPage(ctx context.Context, ordersPage port.Page[domain.Order]) (port.Page[port.OrderDTO], error) {
	if len(ordersPage.Items) == 0 {
		return port.Page[port.OrderDTO]{
			Items: []port.OrderDTO{}, Total: ordersPage.Total, Limit: ordersPage.Limit, Offset: ordersPage.Offset,
		}, nil
	}

	// Collect all order IDs for batch query.
	orderIDs := make([]string, len(ordersPage.Items))
	for i, order := range ordersPage.Items {
		orderIDs[i] = order.ID()
	}

	// Single batch query for all items.
	itemsMap, err := s.deps.Repos.Items.ListByOrderIDs(ctx, s.pool, orderIDs)
	if err != nil {
		return port.Page[port.OrderDTO]{}, err
	}

	// Build DTOs.
	dtos := make([]port.OrderDTO, 0, len(ordersPage.Items))
	for _, order := range ordersPage.Items {
		items := itemsMap[order.ID()]
		if items == nil {
			items = []domain.OrderItem{}
		}
		dtos = append(dtos, toOrderDTO(order, items))
	}
	return port.Page[port.OrderDTO]{
		Items:  dtos,
		Total:  ordersPage.Total,
		Limit:  ordersPage.Limit,
		Offset: ordersPage.Offset,
	}, nil
}
