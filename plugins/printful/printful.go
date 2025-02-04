package printful

import "github.com/stripe/stripe-go/v81"

// This is a plugin to handle physical apparel printing requests to Printful

// func Validate()
// func Price()
func Handle(event *stripe.Event) {
	switch event.Type {
	case "checkout.session.completed":
		// get product name and quantity from metadata
		// send request to print whatever product
		return
	case "another event type":
		return
	default:
		return
	}
}
