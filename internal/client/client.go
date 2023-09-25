package client

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gsk148/gophermart/internal/models"
	"github.com/gsk148/gophermart/internal/storage"
)

type accrualClient struct {
	db       *storage.Repository
	host     string
	endpoint string
	poolSize int
	rw       sync.RWMutex
	m        sync.Mutex
}

func NewAccrualClient(db *storage.Repository, host string, poolSize int) *accrualClient {
	proc := &accrualClient{
		db:       db,
		host:     host,
		endpoint: "/api/orders/",
		poolSize: poolSize,
	}
	return proc
}

func (a *accrualClient) GetOrderByNumber(orderNum string) (*models.GetOrderAccrual, error) {
	a.rw.RLock()
	defer a.rw.RUnlock()

	res, err := http.Get(a.host + a.endpoint + orderNum)
	if err != nil {
		log.Println("GetOrderByNumber err http get", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		lock := a.m.TryLock()
		if lock {
			a.rw.Lock()
			duration := returnRetryDuration(res)
			currentTime := time.Now()
			wakeUpTime := currentTime.Add(duration * time.Second)
			go func() {
				time.Sleep(time.Until(wakeUpTime))
				a.rw.Unlock()
				a.m.Unlock()
			}()
		}
		return nil, errors.New("429 error code")
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("non-200 status code")
	}

	var orders models.GetOrderAccrual
	if err := json.NewDecoder(res.Body).Decode(&orders); err != nil {
		log.Print("GetOrderByNumber err", err)
		return nil, err
	}
	return &orders, nil
}

func returnRetryDuration(r *http.Response) time.Duration {
	headerParam := "Retry-After"
	duration := r.Header.Get(headerParam)
	convertedDuration, err := strconv.Atoi(duration)
	if err != nil {
		return 0
	}
	return time.Duration(convertedDuration)
}
