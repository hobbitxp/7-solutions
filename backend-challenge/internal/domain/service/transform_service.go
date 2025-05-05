package service

import (
	"context"
	"errors"

	"backend-challenge/internal/domain/model"
	"backend-challenge/internal/domain/repository"
)

// Common errors
var (
	ErrExternalAPIFailed = errors.New("failed to fetch data from external API")
	ErrDataTransformation = errors.New("failed to transform data")
)

// TransformService defines the transformation service interface
type TransformService interface {
	// GroupUsersByDepartment groups users by department
	GroupUsersByDepartment(ctx context.Context, input *model.GroupUsersByDepartmentInput) (model.DepartmentGroupedData, error)
	
	// FetchAndTransform fetches users from external API and transforms the data
	FetchAndTransform(ctx context.Context, input *model.FetchAndTransformInput) (model.DepartmentGroupedData, error)
	
	// ImportFromExternalAPI imports users from external API to database
	ImportFromExternalAPI(ctx context.Context, apiURL string) (int, error)
	
	// TransformFromDatabase transforms user data from database
	TransformFromDatabase(ctx context.Context) (model.DepartmentGroupedData, error)
	
	// TransformDepartmentFromDatabase transforms user data for a specific department
	TransformDepartmentFromDatabase(ctx context.Context, department string) (model.DepartmentGroupedData, error)
}

// transformService implements TransformService
type transformService struct {
	externalRepo repository.ExternalUserRepository
}

// NewTransformService creates a new transform service
func NewTransformService(externalRepo repository.ExternalUserRepository) TransformService {
	return &transformService{
		externalRepo: externalRepo,
	}
}

// GroupUsersByDepartment groups users by department
func (s *transformService) GroupUsersByDepartment(ctx context.Context, input *model.GroupUsersByDepartmentInput) (model.DepartmentGroupedData, error) {
	if input == nil || len(input.Users) == 0 {
		return model.DepartmentGroupedData{}, nil
	}

	return model.GroupUsersByDepartment(input.Users), nil
}

// FetchAndTransform fetches users from external API and transforms the data
func (s *transformService) FetchAndTransform(ctx context.Context, input *model.FetchAndTransformInput) (model.DepartmentGroupedData, error) {
	// First try to get data from database
	result, err := s.TransformFromDatabase(ctx)
	if err == nil && len(result) > 0 {
		return result, nil
	}
	
	// If no data in database or error occurred, import from API
	apiURL := ""
	if input != nil {
		apiURL = input.APIURL
	}
	
	// Import data from API to database
	_, err = s.ImportFromExternalAPI(ctx, apiURL)
	if err != nil {
		return nil, err
	}
	
	// Now read from database and transform
	return s.TransformFromDatabase(ctx)
}

// ImportFromExternalAPI imports users from external API to database
func (s *transformService) ImportFromExternalAPI(ctx context.Context, apiURL string) (int, error) {
	// Use repository to import data
	return s.externalRepo.ImportFromAPI(ctx, apiURL)
}

// TransformFromDatabase transforms user data from database
func (s *transformService) TransformFromDatabase(ctx context.Context) (model.DepartmentGroupedData, error) {
	// Get all users from database
	users, err := s.externalRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	
	// Convert to ExternalUserData for transformation
	externalUsers := make([]model.ExternalUserData, 0, len(users))
	for _, user := range users {
		externalUsers = append(externalUsers, user.ToExternalUserData())
	}
	
	// Transform the data
	return model.GroupUsersByDepartment(externalUsers), nil
}

// TransformDepartmentFromDatabase transforms user data for a specific department
func (s *transformService) TransformDepartmentFromDatabase(ctx context.Context, department string) (model.DepartmentGroupedData, error) {
	// Get users from specific department
	users, err := s.externalRepo.ListByDepartment(ctx, department)
	if err != nil {
		return nil, err
	}
	
	// Convert to ExternalUserData for transformation
	externalUsers := make([]model.ExternalUserData, 0, len(users))
	for _, user := range users {
		externalUsers = append(externalUsers, user.ToExternalUserData())
	}
	
	// Transform the data
	return model.GroupUsersByDepartment(externalUsers), nil
}