package types

import "time"

// Auth
type User struct {
	Id       int
	Customer string
	Username string
	Email    string
	Password string
	Verified bool
	Updated  time.Time
	Created  time.Time
}

type UserSignInPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12,max=128,containsany=!£$%&?^*@#"`
}

type UserSignUpPayload struct {
	Username string `json:"username" validate:"required,min=3,max=16,alphanum,startsnotwith=number"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12,max=128,containsany=!£$%&?^*@#"`
}

type SendVerificationEmailPayload struct {
	Email string `json:"email" validate:"required,email"`
}

type CompleteVerificationPayload struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6,ascii"`
}

// Users
type UpdateUserNamePayload struct {
	New string `json:"new" validate:"required,min=3,max=16,alphanum,startsnotwith=number"`
	Old string `json:"old" validate:"required,min=3,max=16,alphanum,startsnotwith=number"`
}

type UpdateUserEmailPayload struct {
	New string `json:"new" validate:"required,email"`
	Old string `json:"old" validate:"required,email"`
}

type UpdateUserPasswordPayload struct {
	Password string `json:"password" validate:"required,min=12,max=128,containsany=!£$%&?^*@#"`
}

type DeleteAccountPayload struct {
	Password string `json:"password" validate:"required,min=12,max=128,containsany=!£$%&?^*@#"`
}

// Stripe
type NewStripeCheckout struct {
	Product  string `json:"product" validate:"required"`
	Quantity string `json:"quantity" validate:"required"`
}
