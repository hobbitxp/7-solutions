package model

import (
	"fmt"
	"math"
)

// ExternalUserData represents data structure from external APIs like dummyjson
type ExternalUserData struct {
	ID         int    `json:"id"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	MaidenName string `json:"maidenName"`
	Age        int    `json:"age"`
	Gender     string `json:"gender"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	BirthDate  string `json:"birthDate"`
	Image      string `json:"image"`
	Height     float64 `json:"height"`
	Weight     float64 `json:"weight"`
	Hair       struct {
		Color string `json:"color"`
		Type  string `json:"type"`
	} `json:"hair"`
	Address struct {
		Address    string `json:"address"`
		City       string `json:"city"`
		PostalCode string `json:"postalCode"`
		State      string `json:"state"`
	} `json:"address"`
	Bank struct {
		CardExpire string `json:"cardExpire"`
		CardNumber string `json:"cardNumber"`
		CardType   string `json:"cardType"`
		Currency   string `json:"currency"`
		IBAN       string `json:"iban"`
	} `json:"bank"`
	Company struct {
		Address struct {
			Address    string `json:"address"`
			City       string `json:"city"`
			PostalCode string `json:"postalCode"`
			State      string `json:"state"`
		} `json:"address"`
		Department string `json:"department"`
		Name       string `json:"name"`
		Title      string `json:"title"`
	} `json:"company"`
}

// DepartmentData represents transformed data for a department
type DepartmentData struct {
	Male        int               `json:"male"`
	Female      int               `json:"female"`
	AgeRange    string            `json:"ageRange"`
	Hair        map[string]int    `json:"hair"`
	AddressUser map[string]string `json:"addressUser"`
	MinAge      int               `json:"-"` // Internal use for calculating age range
	MaxAge      int               `json:"-"` // Internal use for calculating age range
}

// DepartmentGroupedData represents the final transformed data grouped by department
type DepartmentGroupedData map[string]*DepartmentData

// FetchAndTransformInput represents the input for fetching and transforming external user data
type FetchAndTransformInput struct {
	APIURL string `json:"apiUrl"`
}

// GroupUsersByDepartmentInput represents the input for grouping users by department
type GroupUsersByDepartmentInput struct {
	Users []ExternalUserData `json:"users"`
}

// GroupUsersByDepartment transforms a list of users into department-grouped data
func GroupUsersByDepartment(users []ExternalUserData) DepartmentGroupedData {
	result := make(DepartmentGroupedData)

	for _, user := range users {
		department := user.Company.Department
		
		// Initialize department data if not exists
		if _, exists := result[department]; !exists {
			result[department] = &DepartmentData{
				Hair:        make(map[string]int),
				AddressUser: make(map[string]string),
				MinAge:      math.MaxInt32, // Start with max possible value
				MaxAge:      0,             // Start with min possible value
			}
		}

		// Increment gender count
		if user.Gender == "male" {
			result[department].Male++
		} else if user.Gender == "female" {
			result[department].Female++
		}

		// Update age range
		if user.Age < result[department].MinAge {
			result[department].MinAge = user.Age
		}
		if user.Age > result[department].MaxAge {
			result[department].MaxAge = user.Age
		}

		// Update hair color count
		hairColor := user.Hair.Color
		result[department].Hair[hairColor]++

		// Add user address
		fullName := user.FirstName + user.LastName
		result[department].AddressUser[fullName] = user.Address.PostalCode
	}

	// Calculate age range string for each department
	for _, deptData := range result {
		deptData.AgeRange = fmt.Sprintf("%d-%d", deptData.MinAge, deptData.MaxAge)
	}

	return result
}