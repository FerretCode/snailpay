package stripe

import (
	byteReader "bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/deta/deta-go/deta"
	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v74"
	stripeFile "github.com/stripe/stripe-go/v74/file"
	"github.com/stripe/stripe-go/v74/paymentlink"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
)

type shortenSubscriptionLink struct {
	ID string `json:"id"`
	URL string `json:"url"`
}

type SubscriptionLinkRequest struct {
	Image string `json:"image"`
	Name string `json:"name"`
	Price int `json:"price"`
}

type SubscriptionLinkRecord struct {
	URL string `json:"url"`
	ID string `json:"key"`
	UserId string `json:"user_id"`
	ProductName string `json:"product_name"`
	PaymentLinkId string `json:"payment_link_id"`
	Subscription bool `json:"subscription"`
}

//	@title			Subscription Link 
//	@version		1.0
//	@description	Create a new subscription link 

//	@BasePath	/api

//	@accept		json
//	@produce	json

//	@schemes	https

//	@securityDefinitions.apiKey	ApiKeyAuth

//	@Param	image	body	string	false	"the base64 encoded image for the product"	
//	@Param	name	body	string	false	"the name for the product"
//	@Param	price	body	int		false	"the monthly price for the product"

func SubscriptionLink(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	stripe.Key = os.Getenv("STRIPE_TOKEN")

	deta, err := deta.New(deta.WithProjectKey(os.Getenv("PROJECT_KEY")))

	if err != nil {
		return err
	}

	linkDb, err := base.New(deta, "links")

	if err != nil {
		return err
	}

	user := r.Context().Value("user").(auth.SnailUser)

	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	subscriptionLinkRequest := SubscriptionLinkRequest{}

	if err := json.Unmarshal(bytes, &subscriptionLinkRequest); err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(subscriptionLinkRequest.Image)

	if err != nil {
		return err
	}

	fileName := uuid.NewString()

	file, err := os.Create(fmt.Sprintf("./tmp/%s.jpg", fileName))

	if err != nil {
		return err
	}

	if _, err := file.Write(decoded); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	f, err := os.Open(fmt.Sprintf("./tmp/%s.jpg", fileName))

	if err != nil {
		return err
	}

	fileParams := &stripe.FileParams{
		FileReader: f,	
		Filename: stripe.String(fileName + ".jpg"),
		Purpose: stripe.String(string(stripe.FilePurposeBusinessIcon)),
	}

	fileParams.SetStripeAccount(user.AccountId)

	stripeFile, err := stripeFile.New(fileParams)

	if err != nil {
		return err
	}

	err = os.Remove(fmt.Sprintf("./tmp/%s.jpg", fileName))	

	if err != nil {
		return err
	}

	productParams := &stripe.ProductParams{
		Name: stripe.String(subscriptionLinkRequest.Name),
		Images: []*string{stripe.String(stripeFile.URL)},
	}

	productParams.SetStripeAccount(user.AccountId)

	product, err := product.New(productParams)

	if err != nil {
		return err
	}

	priceParams := &stripe.PriceParams{
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		Product: stripe.String(product.ID),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String("month"),
		},
		UnitAmount: stripe.Int64(int64(subscriptionLinkRequest.Price * 100)),
	}

	priceParams.SetStripeAccount(user.AccountId)

	price, err := price.New(priceParams)

	if err != nil {
		return err
	}

	paymentLinkParams := stripe.PaymentLinkParams{
		LineItems: []*stripe.PaymentLinkLineItemParams{
			{
				Price: stripe.String(price.ID),
				Quantity: stripe.Int64(1),
			},
		},
		ApplicationFeePercent: stripe.Float64(5),
		AfterCompletion: &stripe.PaymentLinkAfterCompletionParams{
			Redirect: &stripe.PaymentLinkAfterCompletionRedirectParams{
				URL: stripe.String(
					fmt.Sprintf("%s/redirect?account_id=%s&session_id={CHECKOUT_SESSION_ID}", os.Getenv("HOST"), user.AccountId),
				),
			},
			Type: stripe.String("redirect"),
		},
	}

	paymentLinkParams.SetStripeAccount(user.AccountId)

	paymentLink, err := paymentlink.New(&paymentLinkParams)

	if err != nil {
		return err
	}

	subscriptionLinkId := uuid.NewString() 

	subscriptionLinkRecord := SubscriptionLinkRecord{
		ID: subscriptionLinkId,
		URL: paymentLink.URL,
		UserId: user.UserId,
		ProductName: product.Name,
		PaymentLinkId: paymentLink.ID,
		Subscription: true,
	}

	_, err = linkDb.Put(subscriptionLinkRecord)

	if err != nil {
		return err
	}

	client := &http.Client{}

	shortenSubscriptionLink := shortenSubscriptionLink{
		ID: subscriptionLinkId,
		URL: paymentLink.URL,
	}

	shortenBytes, err := json.Marshal(shortenSubscriptionLink)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/api/shorten-payment-link", 
			os.Getenv("HOST"),
		),
		byteReader.NewReader(shortenBytes),
	)

	cookie, err := r.Cookie("snail")

	if err != nil {
		return err
	}

	req.AddCookie(cookie)

	res, err := client.Do(req)

	if err != nil {
		return err
	}

	resBytes, err := io.ReadAll(res.Body)

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write(resBytes)

	return nil
}
