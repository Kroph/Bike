package service_test

import (
	"context"
	"testing"

	"api-gateway/service"
	"proto/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserServiceClient struct {
	mock.Mock
	user.UserServiceClient
}

func (m *MockUserServiceClient) GenerateVerificationCode(ctx context.Context, req *user.GenerateCodeRequest, opts ...interface{}) (*user.GenerateCodeResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*user.GenerateCodeResponse), args.Error(1)
}

func TestGenerateVerificationCode_Success(t *testing.T) {
	mockClient := new(MockUserServiceClient)
	ctx := context.TODO()
	userID := "test-user-id"

	mockClient.On("GenerateVerificationCode", ctx, mock.Anything).
		Return(&user.GenerateCodeResponse{
			Success: true,
			Code:    "123456",
			Message: "OK",
		}, nil)

	clients := &service.GrpcClients{}
	// unsafe hack to inject mock if field is unexported in real code
	// but assume for this test it's assignable or field is exported (userClient)
	clientsReflection := *clients
	clientsReflection.userClient = mockClient

	// Call method
	code, err := clientsReflection.GenerateVerificationCode(ctx, userID)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, "123456", code)
	mockClient.AssertExpectations(t)
}
