package handlers

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"zero/config"
	"zero/services"
	"zero/types"
	"zero/utils"

	"github.com/charmbracelet/log"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/webhook"
)

type StripeHandler struct {
	US *services.UsersService
}

// Creates a new Stripe checkout link
func (h *StripeHandler) StripeCheckout(w http.ResponseWriter, r *http.Request) {
	logger := config.GetLogger()
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	domain := os.Getenv("STRIPE_DOMAIN")
	if stripe.Key == "" || domain == "" {
		log.Fatal("Stripe enviroment variables are not set")
	}
	payload := new(types.NewStripeCheckout)
	if err := utils.Validate(w, r, payload); err != nil {
		return
	}

	// Extra validations that are necessary
	// plugins/name.Validate()

	// Getting user id from request context
	id, err := services.GetUserIdFromContext(r)
	if err != nil {
		utils.Response(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if id == 0 {
		log.Warn("failed to get user from context")
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		logger.Error("failed to get user from context", "id", id, "err", err)
		return
	}
	// Getting user by id
	user, err := h.US.GetUserById(id)
	if err != nil {
		utils.Response(w, http.StatusUnauthorized, "user not found")
		return
	}
	// Getting a new Stripe customer id if he/she doesn't have one
	if user.Customer == "" {
		user.Customer = CreateCustomer(user.Username, user.Email)
	}

	// Let plugins handle price calculations with Stripe
	// plugins/name.Price()

	checkoutPrice := float64(99)
	user_id := strconv.Itoa(user.Id)
	checkoutQuantity, err := strconv.ParseFloat(payload.Quantity, 64)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "interal server error")
		logger.Error("failed to conver float64", "quantity", payload.Quantity, "err")
		return
	}
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card", "paypal"}),
		Metadata: map[string]string{
			"user":     user_id,
			"product":  payload.Product,
			"quantity": payload.Quantity,
		},
		Customer: &user.Customer,
		LineItems: []*stripe.CheckoutSessionLineItemParams{{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("eur"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(payload.Product),
				},
				UnitAmount: stripe.Int64(int64(checkoutPrice)),
			},
			Quantity: stripe.Int64(int64(checkoutQuantity)),
		}},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(domain + "/success"),
		CancelURL:  stripe.String(domain + "/failure"),
	}
	stripeSession, err := session.New(params)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		logger.Error("failed to create checkout session", "user", user.Id, "metadata", params.Metadata, "customer", params.Customer, "err", err)
		return
	}
	logger.Info("checkout link", "product", payload.Product, "quantity", payload.Quantity)
	utils.Response(w, http.StatusSeeOther, stripeSession.URL)
}

// Handles Stripe events
func (h *StripeHandler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	logger := config.GetLogger()
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("stripe webhook unavailable", "err", err)
		utils.Response(w, http.StatusServiceUnavailable, "request body error")
		return
	}
	sigHeader := r.Header.Get("Stripe-Signature")
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if endpointSecret == "" {
		log.Error("stripe webhook unavailable", "STRIPE_WEBHOOK_SECRET", endpointSecret)
		return
	}
	// Creating a new Stripe event
	event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, "internal server error")
		logger.Error("signature verification failed", "err", err)
		return
	}

	// Plugins will handle the Stripe event
	// plugins/name.Handle()
	// example printful.Handle(&event)

	switch event.Type {
	case "checkout.session.completed":
		return
	case "invoice.payment_failed":
		return
	default:
		return
	}
}

// Creates a Stripe customer
func CreateCustomer(name, email string) string {
	logger := config.GetLogger()
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	params := &stripe.CustomerParams{
		Name:  &name,
		Email: &email,
	}
	customer, err := customer.New(params)
	if err != nil {
		logger.Error("failed to create stripe customer", "err", err)
		return ""
	}
	return string(customer.ID)
}
