// Code generated by MockGen. DO NOT EDIT.
// Source: vk_film/internal/repository/session (interfaces: Repository)
//
// Generated by this command:
//
//	mockgen -destination=mocks/repository.go -package=mr -mock_names=Repository=SessionRepository . Repository
//

// Package mr is a generated GoMock package.
package mr

import (
	reflect "reflect"
	time "time"
	types "vk_film/internal/pkg/types"

	gomock "go.uber.org/mock/gomock"
)

// SessionRepository is a mock of Repository interface.
type SessionRepository struct {
	ctrl     *gomock.Controller
	recorder *SessionRepositoryMockRecorder
}

// SessionRepositoryMockRecorder is the mock recorder for SessionRepository.
type SessionRepositoryMockRecorder struct {
	mock *SessionRepository
}

// NewSessionRepository creates a new mock instance.
func NewSessionRepository(ctrl *gomock.Controller) *SessionRepository {
	mock := &SessionRepository{ctrl: ctrl}
	mock.recorder = &SessionRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *SessionRepository) EXPECT() *SessionRepositoryMockRecorder {
	return m.recorder
}

// Del mocks base method.
func (m *SessionRepository) Del(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Del", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Del indicates an expected call of Del.
func (mr *SessionRepositoryMockRecorder) Del(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Del", reflect.TypeOf((*SessionRepository)(nil).Del), arg0)
}

// GetUserId mocks base method.
func (m *SessionRepository) GetUserId(arg0 string, arg1 time.Duration) (types.Id, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserId", arg0, arg1)
	ret0, _ := ret[0].(types.Id)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserId indicates an expected call of GetUserId.
func (mr *SessionRepositoryMockRecorder) GetUserId(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserId", reflect.TypeOf((*SessionRepository)(nil).GetUserId), arg0, arg1)
}

// Set mocks base method.
func (m *SessionRepository) Set(arg0 string, arg1 types.Id, arg2 time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *SessionRepositoryMockRecorder) Set(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*SessionRepository)(nil).Set), arg0, arg1, arg2)
}