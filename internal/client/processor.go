package client

import (
	"log"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/gsk148/gophermart/internal/enums"
)

func (a *accrualClient) Run() {
	var errgr errgroup.Group
	errgr.SetLimit(a.poolSize)
	for {
		orderList, err := a.db.GetOrdersForProcessing(a.poolSize)
		if err != nil || len(orderList) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}
		for _, orderID := range orderList {
			orderID := orderID
			errgr.Go(func() error {
				receivedOrder, err := a.GetOrderByNumber(orderID)
				if err != nil {
					return err
				}
				if receivedOrder != nil {
					log.Println("received order status ", *receivedOrder)
					if err = a.db.UpdateOrderStateProcessed(receivedOrder); err != nil {
						return err

					} else if receivedOrder.Status == enums.StatusInvalid {
						if err = a.db.UpdateOrderStateInvalid(receivedOrder); err != nil {
							return err
						}
					}
				}
				return nil
			})
		}
	}
}
