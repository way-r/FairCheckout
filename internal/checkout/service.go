package checkout

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CheckoutService struct {
	RedisClusterClient *redis.ClusterClient
}

type CheckoutResult struct {
	EventId int64
}

func (cs *CheckoutService) ProcessCheckout(ctx context.Context, checkoutAddress CheckoutAddress) CheckoutResult {
	lockKey := fmt.Sprintf("address:{%s}%s_%s_%s", checkoutAddress.ZipCode, checkoutAddress.StreetAddress, checkoutAddress.City, checkoutAddress.State)

	user_id := uuid.New().String()
	err := cs.acquireLock(ctx, lockKey, user_id)
	if err != nil {
		return CheckoutResult{
			EventId: 101,
		}
	}
	defer cs.releaseLock(context.Background(), lockKey, user_id)
	return CheckoutResult{
		EventId: 101,
	}

}

func (cs *CheckoutService) acquireLock(ctx context.Context, lockKey string, userID string) error {
	acquired, err := cs.RedisClusterClient.SetNX(ctx, lockKey, userID, 2*time.Second).Result()
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	if !acquired {
		return fmt.Errorf("Key-%s is currently locked by another transaction", lockKey)
	}
	return nil
}

func (cs *CheckoutService) releaseLock(ctx context.Context, lockKey string, userID string) {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	cs.RedisClusterClient.Eval(ctx, script, []string{lockKey}, userID)
}
