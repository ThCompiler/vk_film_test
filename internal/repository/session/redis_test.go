package session

import (
	"fmt"
	"github.com/go-redis/redismock/v9"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/pkg/errors"
	"testing"
	"time"
	"vk_film/internal/pkg/types"
)

var testError = errors.New("test error")

type RedisRepositorySuite struct {
	suite.Suite
	redisRepository *RedisSession
	mock            redismock.ClientMock
}

func (rrs *RedisRepositorySuite) BeforeEach(t provider.T) {
	db, mock := redismock.NewClientMock()
	rrs.redisRepository = NewRedisSession(db)
	rrs.mock = mock
}

func (rrs *RedisRepositorySuite) AfterEach(t provider.T) {
	t.Assert().NoError(rrs.mock.ExpectationsWereMet())
	t.Require().NoError(rrs.redisRepository.client.Close())
}

func (rrs *RedisRepositorySuite) TestSetFunction(t provider.T) {
	t.Title("Set function of Redis repository")
	t.NewStep("Init test data")
	timeExpired := 2 * time.Second
	sessionId := "id"
	userId := types.Id(1)

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {

		t.NewStep("Init mock")
		rrs.mock.ExpectSet(sessionId, uint64(userId), timeExpired).SetVal(sessionId)

		t.NewStep("Check result")
		t.Require().NoError(rrs.redisRepository.Set(sessionId, userId, timeExpired))
	})

	t.WithNewStep("Redis error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		rrs.mock.ExpectSet(sessionId, uint64(userId), timeExpired).SetErr(testError)

		t.NewStep("Check result")
		t.Require().ErrorIs(rrs.redisRepository.Set(sessionId, userId, timeExpired), testError)
	})
}

func (rrs *RedisRepositorySuite) TestGetFunction(t provider.T) {
	t.Title("GetUserId function of Redis repository")
	t.NewStep("Init test data")
	timeExpired := 2 * time.Second
	sessionId := "id"
	userId := types.Id(1)

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		rrs.mock.ExpectGet(sessionId).SetVal(fmt.Sprintf("%d", userId))
		rrs.mock.ExpectExpire(sessionId, timeExpired).SetVal(true)

		t.NewStep("Check result")
		resUserId, err := rrs.redisRepository.GetUserId(sessionId, timeExpired)
		t.Require().NoError(err)
		t.Require().Equal(userId, resUserId)
	})

	t.WithNewStep("No records execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		rrs.mock.ExpectGet(sessionId).RedisNil()

		t.NewStep("Check result")
		_, err := rrs.redisRepository.GetUserId(sessionId, timeExpired)
		t.Require().ErrorIs(err, ErrorNoSession)
	})

	t.WithNewStep("Redis error execute of redis get", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		rrs.mock.ExpectGet(sessionId).SetErr(testError)

		t.NewStep("Check result")
		_, err := rrs.redisRepository.GetUserId(sessionId, timeExpired)
		t.Require().ErrorIs(err, testError)
	})

	t.WithNewStep("Redis error execute of redis expire", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		rrs.mock.ExpectGet(sessionId).SetVal(fmt.Sprintf("%d", userId))
		rrs.mock.ExpectExpire(sessionId, timeExpired).SetErr(testError)

		t.NewStep("Check result")
		_, err := rrs.redisRepository.GetUserId(sessionId, timeExpired)
		t.Require().ErrorIs(err, testError)
	})
}

func (rrs *RedisRepositorySuite) TestDeleteFunction(t provider.T) {
	t.Title("Get function of Redis repository")
	t.NewStep("Init test data")
	sessionId := "id"

	t.WithNewStep("Correct execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		rrs.mock.ExpectDel(sessionId).SetVal(0)

		t.NewStep("Check result")
		t.Require().NoError(rrs.redisRepository.Del(sessionId))
	})

	t.WithNewStep("Redis error execute", func(t provider.StepCtx) {
		t.NewStep("Init mock")
		rrs.mock.ExpectDel(sessionId).SetErr(testError)

		t.NewStep("Check result")
		t.Require().ErrorIs(rrs.redisRepository.Del(sessionId), testError)
	})
}

func TestRunRedisRepositorySuite(t *testing.T) {
	suite.RunSuite(t, new(RedisRepositorySuite))
}
