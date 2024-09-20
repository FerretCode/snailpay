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

type shortenPaymentLink struct {
	ID string `json:"id"`
	URL string `json:"url"`
}

type PaymentLinkRequest struct {
	Image string `json:"image"`
	Name string `json:"name"`
	Price float64 `json:"price"`
}

type PaymentLinkRecord struct {
	URL string `json:"url"`
	ID string `json:"key"`
	UserId string `json:"user_id"`
	ProductName string `json:"product_name"`
	PaymentLinkId string `json:"payment_link_id"`
	Subscription bool `json:"subscription"`
}

func PaymentLink(w http.ResponseWriter, r *http.Request, db *base.Base) error {
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

	paymentLinkRequest := PaymentLinkRequest{}

	if err := json.Unmarshal(bytes, &paymentLinkRequest); err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(paymentLinkRequest.Image)

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
		Name: stripe.String(paymentLinkRequest.Name),
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
		UnitAmount: stripe.Int64(int64(paymentLinkRequest.Price * 100)),
	}

	priceParams.SetStripeAccount(user.AccountId)

	price, err := price.New(priceParams)

	if err != nil {
		return err
	}

	paymentLinkId := uuid.NewString() 

	paymentLinkParams := stripe.PaymentLinkParams{
		LineItems: []*stripe.PaymentLinkLineItemParams{
			{
				Price: stripe.String(price.ID),
				Quantity: stripe.Int64(1),
			},
		},
		CustomerCreation: stripe.String("always"),
		ApplicationFeeAmount: stripe.Int64(int64((paymentLinkRequest.Price * 100) * 2/100)),
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

	paymentLinkRecord := PaymentLinkRecord{
		ID: paymentLinkId,
		URL: paymentLink.URL,
		UserId: user.UserId,
		ProductName: product.Name,
		PaymentLinkId: paymentLink.ID,
		Subscription: false,
	}

	_, err = linkDb.Put(paymentLinkRecord)

	if err != nil {
		return err
	}

	client := &http.Client{}

	shortenPaymentLink := shortenPaymentLink{
		ID: paymentLinkId,
		URL: paymentLink.URL,
	}

	shortenBytes, err := json.Marshal(shortenPaymentLink)

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
