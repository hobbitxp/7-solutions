package model

import "time"

// ExternalUser represents a user in the external system (imported from API)
type ExternalUser struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	FirstName  string    `json:"firstName" bson:"first_name"`
	LastName   string    `json:"lastName" bson:"last_name"`
	MaidenName string    `json:"maidenName" bson:"maiden_name"`
	Age        int       `json:"age" bson:"age"`
	Gender     string    `json:"gender" bson:"gender"`
	Email      string    `json:"email" bson:"email"`
	Phone      string    `json:"phone" bson:"phone"`
	Username   string    `json:"username" bson:"username"`
	Password   string    `json:"password" bson:"password,omitempty"`
	BirthDate  string    `json:"birthDate" bson:"birth_date"`
	Image      string    `json:"image" bson:"image"`
	Height     float64   `json:"height" bson:"height"`
	Weight     float64   `json:"weight" bson:"weight"`
	Hair       HairData  `json:"hair" bson:"hair"`
	Address    Address   `json:"address" bson:"address"`
	Bank       BankData  `json:"bank" bson:"bank"`
	Company    Company   `json:"company" bson:"company"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
}

// HairData represents hair information
type HairData struct {
	Color string `json:"color" bson:"color"`
	Type  string `json:"type" bson:"type"`
}

// Address represents an address
type Address struct {
	Address    string `json:"address" bson:"address"`
	City       string `json:"city" bson:"city"`
	PostalCode string `json:"postalCode" bson:"postal_code"`
	State      string `json:"state" bson:"state"`
}

// BankData represents bank information
type BankData struct {
	CardExpire string `json:"cardExpire" bson:"card_expire"`
	CardNumber string `json:"cardNumber" bson:"card_number"`
	CardType   string `json:"cardType" bson:"card_type"`
	Currency   string `json:"currency" bson:"currency"`
	IBAN       string `json:"iban" bson:"iban"`
}

// Company represents company information
type Company struct {
	Address    Address `json:"address" bson:"address"`
	Department string  `json:"department" bson:"department"`
	Name       string  `json:"name" bson:"name"`
	Title      string  `json:"title" bson:"title"`
}

// ImportExternalUserInput represents the input for importing external user data
type ImportExternalUserInput struct {
	APIURL string `json:"apiUrl"`
}

// To make the connection with the previous ExternalUserData structure
func (e *ExternalUser) ToExternalUserData() ExternalUserData {
	return ExternalUserData{
		ID:         0, // Not used for transformation
		FirstName:  e.FirstName,
		LastName:   e.LastName,
		MaidenName: e.MaidenName,
		Age:        e.Age,
		Gender:     e.Gender,
		Email:      e.Email,
		Phone:      e.Phone,
		Username:   e.Username,
		Password:   e.Password,
		BirthDate:  e.BirthDate,
		Image:      e.Image,
		Height:     e.Height,
		Weight:     e.Weight,
		Hair: struct {
			Color string `json:"color"`
			Type  string `json:"type"`
		}{
			Color: e.Hair.Color,
			Type:  e.Hair.Type,
		},
		Address: struct {
			Address    string `json:"address"`
			City       string `json:"city"`
			PostalCode string `json:"postalCode"`
			State      string `json:"state"`
		}{
			Address:    e.Address.Address,
			City:       e.Address.City,
			PostalCode: e.Address.PostalCode,
			State:      e.Address.State,
		},
		Company: struct {
			Address struct {
				Address    string `json:"address"`
				City       string `json:"city"`
				PostalCode string `json:"postalCode"`
				State      string `json:"state"`
			} `json:"address"`
			Department string `json:"department"`
			Name       string `json:"name"`
			Title      string `json:"title"`
		}{
			Address: struct {
				Address    string `json:"address"`
				City       string `json:"city"`
				PostalCode string `json:"postalCode"`
				State      string `json:"state"`
			}{
				Address:    e.Company.Address.Address,
				City:       e.Company.Address.City,
				PostalCode: e.Company.Address.PostalCode,
				State:      e.Company.Address.State,
			},
			Department: e.Company.Department,
			Name:       e.Company.Name,
			Title:      e.Company.Title,
		},
	}
}