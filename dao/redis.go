package dao

import (
	"context"
	"errors"
	"sync"

	"github.com/padam-meesho/NotificationService/config"
	"github.com/padam-meesho/NotificationService/internal/utils"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// this file shall have the different CRUD operations
// to the different data access object (DAO) layers.

// Example of a generic Redis configuration struct and usage:

type RedisDao interface {
	AddToRedisSet(ctx context.Context, setname string, val string) error
	CheckInRedisSet(ctx context.Context, setname string, val string) (bool, error)
	GetRedisSetMembers(ctx context.Context, setname string, val string) ([]string, error)
}

var (
	BLACKLISTED_NUMBERS_SET = "blacklisted_numbers_set"
)

type RedisDaoImpl struct {
	redisClient *redis.Client
}

var (
	redisDaoOnce  sync.Once
	redisInstance *RedisDaoImpl
)

func NewRedisDao() *RedisDaoImpl {
	redisDaoOnce.Do(func() {
		redisInstance = &RedisDaoImpl{
			redisClient: config.GetRedisClient().RedisClient,
		}
	})
	return redisInstance
}

// okay we need to have a struct that has the DAOs which we shall need to implement the changes.
// we need to have an interface which has all the methods we need to define in our service.
// Each method shall be defined as a struct method it is a part of.

// func GetNewRedisClient(config RedisConfig) *redis.Client {
// 	client := redis.NewClient(&redis.Options{
// 		Addr:     config.Addr,
// 		Password: config.Password,
// 		DB:       config.DB,
// 	})
// 	return client
// }

// here sync.Once ensures, this function is called only once during the entire go runtime.

// this is just syntactical sugarcasing, can be done at the end.
// func BuildRedisOptionsFromEnv(envPrefix string) *redis.Options {

// }

func setOptionalString(key string, option *string) {
	if viper.IsSet(key) {
		*option = viper.GetString(key)
	}
}

// we shall follow the convention that
// if there is an error we generate and return it,
// but if not we return a nil, so that the accepting
//  function can validate if the operation was done
//   successfully or not.

// function names should be defined such that they are easily understandable.
func (r RedisDaoImpl) AddNumberToBlacklistedSet(ctx context.Context, numberToAdd string) error {
	logger := utils.DatabaseLogger(ctx, "sadd", "blacklisted_numbers", "")

	logger.Info().
		Str("phone_number", numberToAdd).
		Msg("Attempting to add number to blacklist")

	result := r.redisClient.SAdd(ctx, BLACKLISTED_NUMBERS_SET, numberToAdd)
	if result.Err() != nil {
		logger.Error().
			Err(result.Err()).
			Str("phone_number", numberToAdd).
			Msg("Failed to add number to Redis blacklist")
		return errors.New("failed to add number to blacklist")
	}

	logger.Info().
		Str("phone_number", numberToAdd).
		Int64("added_count", result.Val()).
		Msg("Successfully added number to blacklist")
	return nil
}

func (r RedisDaoImpl) CheckNumberInBlacklistedSet(ctx context.Context, numberToCheck string) (bool, error) {
	logger := utils.DatabaseLogger(ctx, "sismember", "blacklisted_numbers", "")

	logger.Debug().
		Str("phone_number", numberToCheck).
		Msg("Checking if number is blacklisted")

	exists, err := r.redisClient.SIsMember(ctx, BLACKLISTED_NUMBERS_SET, numberToCheck).Result()
	if err != nil {
		logger.Error().
			Err(err).
			Str("phone_number", numberToCheck).
			Msg("Failed to check number in Redis blacklist")
		return false, errors.New("failed to check blacklist status")
	}

	logger.Debug().
		Str("phone_number", numberToCheck).
		Bool("is_blacklisted", exists).
		Msg("Blacklist check completed")

	return exists, nil
}

func (r RedisDaoImpl) GetAllBlacklistedNumbers(ctx context.Context) ([]string, error) {
	logger := utils.DatabaseLogger(ctx, "smembers", "blacklisted_numbers", "")

	logger.Info().Msg("Retrieving all blacklisted numbers")

	members, err := r.redisClient.SMembers(ctx, BLACKLISTED_NUMBERS_SET).Result()
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to retrieve blacklisted numbers from Redis")
		return nil, errors.New("failed to retrieve blacklisted numbers")
	}

	logger.Info().
		Int("count", len(members)).
		Msg("Successfully retrieved blacklisted numbers")
	return members, nil
}

func (r RedisDaoImpl) RemoveFromBlacklistedSet(ctx context.Context, number string) (int64, error) {
	logger := utils.DatabaseLogger(ctx, "srem", "blacklisted_numbers", "")

	logger.Info().
		Str("phone_number", number).
		Msg("Attempting to remove number from blacklist")

	removedCount, err := r.redisClient.SRem(ctx, BLACKLISTED_NUMBERS_SET, number).Result()
	if err != nil {
		logger.Error().
			Err(err).
			Str("phone_number", number).
			Msg("Failed to remove number from Redis blacklist")
		return 0, errors.New("failed to remove number from blacklist")
	}

	logger.Info().
		Str("phone_number", number).
		Int64("removed_count", removedCount).
		Msg("Successfully removed number from blacklist")
	return removedCount, nil
}

// in a struct we define a type, and then in the variables we define an ibject of that variable,
// now while accessing, we set the object as
// object_name = &type(
// 	congfiguration to set
// )

// note: it is always better to use sync.Once while declaring the singleton instance of any config
// why so that maybe multiple functions may access the config, it needs to be set once just once right?
// how to use singleton instance?
// declare a variable of type sync.Once
// wrap the code that needs to be executed once, in the .Do() method of the variable
