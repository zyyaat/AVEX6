// Package service create_order: CreateOrder use case with idempotency.
package service

import (
        "context"
        "errors"
        "fmt"

        "avex-backend/internal/modules/orders/domain"
        "avex-backend/internal/modules/orders/events"
        "avex-backend/internal/modules/orders/port"
)

func (s *Service) CreateOrder(ctx context.Context, input port.CreateOrderInput) (*port.OrderDTO, error) {
        // 1. Idempotency check: if key exists, return the existing order.
        if input.IdempotencyKey != "" {
                existing, err := s.deps.Repos.Orders.GetByIdempotencyKey(ctx, s.pool, input.IdempotencyKey)
                if err == nil && existing != nil {
                        items, _ := s.deps.Repos.Items.ListByOrder(ctx, s.pool, existing.ID())
                        dto := toOrderDTO(*existing, items)
                        return &dto, nil
                }
                if err != nil && !errors.Is(err, domain.ErrOrderNotFound) {
                        return nil, fmt.Errorf("idempotency check: %w", err)
                }
        }

        // 2. Build domain value objects from input.
        currency := input.Currency
        if currency == "" {
                currency = "EGP"
        }
        subtotal, err := domain.NewMoney(input.SubtotalCents, currency)
        if err != nil {
                return nil, err
        }
        deliveryFee, err := domain.NewMoney(input.DeliveryFeeCents, currency)
        if err != nil {
                return nil, err
        }
        discount, err := domain.NewMoney(input.DiscountCents, currency)
        if err != nil {
                return nil, err
        }
        tax, err := domain.NewMoney(input.TaxCents, currency)
        if err != nil {
                return nil, err
        }
        total, err := domain.NewMoney(input.TotalCents, currency)
        if err != nil {
                return nil, err
        }
        pm, err := domain.ParsePaymentMethod(input.PaymentMethod)
        if err != nil {
                return nil, err
        }
        deliveryInfo, err := domain.NewDeliveryInfo(input.DeliveryLat, input.DeliveryLng, input.DeliveryAddress, input.DeliveryNotes)
        if err != nil {
                return nil, err
        }

        // 3. Build order items.
        items := make([]domain.OrderItem, 0, len(input.Items))
        for _, itemInput := range input.Items {
                price, err := domain.NewMoney(itemInput.PriceCents, currency)
                if err != nil {
                        return nil, err
                }
                item, err := domain.NewOrderItem(itemInput.MenuItemID, itemInput.Name, itemInput.NameAr, price, itemInput.Quantity)
                if err != nil {
                        return nil, err
                }
                items = append(items, item)
        }

        // 4. Create Order entity.
        orderID := s.deps.IDGenerator.NewID()
        orderNumber := s.deps.OrderNumberGenerator.Generate()
        now := s.deps.Clock.Now()

        order, err := domain.NewOrder(domain.OrderParams{
                ID:             orderID,
                OrderNumber:    orderNumber,
                UserID:         input.UserID,
                RestaurantID:   input.RestaurantID,
                CustomerName:   input.CustomerName,
                CustomerPhone:  input.CustomerPhone,
                DeliveryInfo:   deliveryInfo,
                Items:          items,
                Subtotal:       subtotal,
                DeliveryFee:    deliveryFee,
                Discount:       discount,
                Tax:            tax,
                Total:          total,
                PaymentMethod:  pm,
                CouponCode:     input.CouponCode,
                ZoneID:         input.ZoneID,
                DeliveryDistM:  input.DeliveryDistM,
                IdempotencyKey: input.IdempotencyKey,
                Now:            now,
        })
        if err != nil {
                return nil, err
        }

        // 5. Event context.
        ec := s.eventContext(ctx, port.ActorContext{Type: "user", ID: input.UserID})

        // 6. Transaction: persist order + items + history + outbox event.
        // Race condition handling: if two requests with the same idempotency key
        // arrive simultaneously, both pass the pre-check, but the DB UNIQUE constraint
        // on idempotency_key causes one to fail with 23505. We catch that error and
        // re-read the existing order.
        var createdOrder *domain.Order
        var createdItems []domain.OrderItem
        err = s.deps.TxRunner.WithinTx(ctx, func(ctx context.Context, exec port.Executor) error {
                if err := s.deps.Repos.Orders.Create(ctx, exec, order); err != nil {
                        // If duplicate idempotency key, re-read the existing order.
                        if errors.Is(err, domain.ErrOrderAlreadyExists) && input.IdempotencyKey != "" {
                                existing, readErr := s.deps.Repos.Orders.GetByIdempotencyKey(ctx, exec, input.IdempotencyKey)
                                if readErr != nil {
                                        return fmt.Errorf("idempotency conflict re-read failed: %w", readErr)
                                }
                                existingItems, _ := s.deps.Repos.Items.ListByOrder(ctx, exec, existing.ID())
                                createdOrder = existing
                                createdItems = existingItems
                                return nil // Return the existing order — no new event published.
                        }
                        return err
                }
                if err := s.deps.Repos.Items.CreateBatch(ctx, exec, items, order.ID()); err != nil {
                        return err
                }
                if err := s.deps.Repos.History.AddEntry(ctx, exec, order.ID(), domain.StatusPending, input.UserID, "", nil, now); err != nil {
                        return err
                }

                // Publish OrderCreated event (consumed by: notifications, audit).
                payload := port.OrderCreatedPayload{
                        OrderID:       order.ID(),
                        OrderNumber:   order.OrderNumber(),
                        UserID:        order.UserID(),
                        RestaurantID:  order.RestaurantID(),
                        ZoneID:        order.Dispatch().ZoneID(),
                        TotalCents:    order.Total().Amount(),
                        Currency:      order.Total().Currency(),
                        PaymentMethod: order.PaymentMethod().String(),
                        ItemsCount:    order.ItemsCount(),
                        DeliveryLat:   order.DeliveryInfo().Lat(),
                        DeliveryLng:   order.DeliveryInfo().Lng(),
                        RestaurantLat: input.RestaurantLat,
                        RestaurantLng: input.RestaurantLng,
                }
                envelope, err := events.OrderCreatedEnvelope(payload, ec)
                if err != nil {
                        return err
                }
                if err := s.deps.EventPublisher.Publish(ctx, exec, envelope); err != nil {
                        return err
                }

                // IMMEDIATELY publish OrderAssignmentRequested so the dispatch engine
                // starts looking for a driver IN PARALLEL with merchant confirmation
                // and food preparation. The driver travels to the restaurant while
                // the food is being cooked — minimizing total delivery time.
                //
                // If the order is cancelled before a driver accepts, the dispatch
                // module cancels any pending offer via the OrderCancelled event.
                assignmentPayload := port.OrderAssignmentRequestedPayload{
                        OrderID:       order.ID(),
                        RestaurantID:  order.RestaurantID(),
                        ZoneID:        order.Dispatch().ZoneID(),
                        DeliveryLat:   order.DeliveryInfo().Lat(),
                        DeliveryLng:   order.DeliveryInfo().Lng(),
                        RestaurantLat: input.RestaurantLat,
                        RestaurantLng: input.RestaurantLng,
                }
                assignmentEnvelope, err := events.OrderAssignmentRequestedEnvelope(assignmentPayload, ec)
                if err != nil {
                        return err
                }
                if err := s.deps.EventPublisher.Publish(ctx, exec, assignmentEnvelope); err != nil {
                        return err
                }

                createdOrder = &order
                createdItems = items
                return nil
        })
        if err != nil {
                return nil, err
        }

        dto := toOrderDTO(*createdOrder, createdItems)
        return &dto, nil
}
