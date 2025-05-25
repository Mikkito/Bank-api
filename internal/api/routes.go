package api

import (
	"bank-api/internal/cbr"
	"bank-api/internal/config"
	"bank-api/internal/handler"
	"bank-api/internal/middleware"
	"bank-api/internal/payment"
	"bank-api/internal/payment/providers"
	"bank-api/internal/repositories"
	"bank-api/internal/service"
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router, db *sqlx.DB) {

	cfg := config.AppConfig

	// dependency initialization
	userRepo := &repositories.UserRepository{DB: db}
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	accountRepo := &repositories.AccountRepository{DB: db}
	accountService := service.NewAccountService(accountRepo)
	accountHandler := handler.NewAccountHandler(accountService)

	cardRepo := &repositories.CardRepository{DB: db}
	encryptionKey := []byte(cfg.Encryption.Secret)
	hmacKey := []byte(cfg.Encryption.HMACKey)

	cardService := service.NewCardService(cardRepo, encryptionKey, hmacKey)
	cardHandler := handler.NewCardHandler(cardService)

	transactionRepo := &repositories.TransactionRepository{DB: db}
	transactionService := service.NewTransactionService(*transactionRepo, *accountRepo)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	httpClient := &http.Client{}
	cbrService := cbr.NewCBRService(httpClient)

	loanRepo := &repositories.LoanRepository{DB: db}
	loanService := service.NewLoanService(loanRepo, accountRepo, cbrService, transactionService)
	loanHandler := handler.NewLoanHandler(loanService)

	paymentMethodRepo := &repositories.PaymentMethodRepository{DB: db}
	paymentMethodService := service.NewPaymentService(paymentMethodRepo)
	paymentMethodHandler := handler.NewPaymentHandler(paymentMethodService)

	stripeProvider := providers.NewStripeProvider("your-stripe-secret-key")
	yoomoneyProvider := providers.NewYooMoneyProvider("your-yoomoney-token")

	providerList := []providers.PaymentProvider{
		stripeProvider,
		yoomoneyProvider,
	}

	paymentRepo := payment.NewPaymentRepository(db)
	paymentService := payment.NewPaymentService(
		paymentRepo,
		providerList,
		transactionService,
	)
	paymentHandler := payment.NewPaymentHandler(paymentService)

	// Public route
	router.HandleFunc("/register", userHandler.Register).Methods(http.MethodPost)
	router.HandleFunc("/login", userHandler.Login).Methods(http.MethodPost)

	// Здесь позже добавим middleware и защищённые маршруты
	secured := router.PathPrefix("/").Subrouter()
	secured.Use(middleware.JWTMiddleware)

	secured.HandleFunc("/me", userHandler.GetProfile).Methods(http.MethodGet)
	secured.HandleFunc("/cards", cardHandler.CreateCard).Methods(http.MethodPost)
	secured.HandleFunc("/accounts", accountHandler.CreateAccount).Methods(http.MethodPost)
	secured.HandleFunc("/cards/{id:[0-9]+}", cardHandler.GetCardByID).Methods(http.MethodGet)

	// accounts
	auth := router.PathPrefix("/").Subrouter()
	auth.Use(middleware.JWTAuth)
	auth.HandleFunc("/accounts", accountHandler.CreateAccount).Methods("POST")

	// cards
	authCard := router.PathPrefix("/").Subrouter()
	authCard.Use(middleware.JWTAuth)
	authCard.HandleFunc("/cards", cardHandler.CreateCard).Methods("POST")
	authCard.HandleFunc("/cards", cardHandler.GetAllCards).Methods("GET")
	authCard.HandleFunc("/cards/{id}/block", cardHandler.BlockCard).Methods("PATCH")
	authCard.HandleFunc("/cards/{id}", cardHandler.DeleteCard).Methods("DELETE")

	// transactions
	securedTransaction := router.PathPrefix("/transactions").Subrouter()
	securedTransaction.Use(middleware.JWTAuth)

	securedTransaction.HandleFunc("/deposit", transactionHandler.Deposit).Methods("POST")
	securedTransaction.HandleFunc("/withdraw", transactionHandler.Withdraw).Methods("POST")
	securedTransaction.HandleFunc("/transfer", transactionHandler.Transfer).Methods("POST")
	securedTransaction.HandleFunc("/credit", transactionHandler.CreditPayment).Methods("POST")
	securedTransaction.HandleFunc("/reverse", transactionHandler.ReverseTransaction).Methods("POST")
	securedTransaction.HandleFunc("/history/{accountID:[0-9]+}", transactionHandler.GetHistory).Methods("GET")

	// Loan
	securedLoans := router.PathPrefix("/loans").Subrouter()
	securedLoans.Use(middleware.JWTAuth)

	securedLoans.HandleFunc("/take", loanHandler.TakeLoan).Methods("POST")
	securedLoans.HandleFunc("", loanHandler.GetUserLoans).Methods("GET")
	securedLoans.HandleFunc("/{id:[0-9]+}/repay", loanHandler.RepayLoan).Methods("POST")
	securedLoans.HandleFunc("/{id:[0-9]+}/mark-repaid", loanHandler.MarkAsRepaid).Methods("POST")
	securedLoans.HandleFunc("/{id:[0-9]+}/debt", loanHandler.GetOutstandingDebt).Methods("GET")
	securedLoans.HandleFunc("/{id:[0-9]+}/repay-partial", loanHandler.RepayPartial).Methods("POST")

	// payment methods
	securedPaymentsMethod := router.PathPrefix("/payments").Subrouter()
	securedPaymentsMethod.Use(middleware.JWTAuth)
	securedPaymentsMethod.HandleFunc("", paymentMethodHandler.AddPaymentMethod).Methods("POST")
	securedPaymentsMethod.HandleFunc("", paymentMethodHandler.GetUserMethods).Methods("GET")
	securedPaymentsMethod.HandleFunc("/{id}", paymentMethodHandler.DeactivateMethod).Methods("DELETE")

	// Payment
	securedPayments := router.PathPrefix("/api").Subrouter()
	securedPayments.Use(middleware.JWTMiddleware)
	securedPayments.HandleFunc("/payments", paymentHandler.ProcessPayment).Methods("POST")
	securedPayments.HandleFunc("/payments/{id:[0-9]+}/refund", paymentHandler.RefundPayment).Methods("POST")

	securedPayments.HandleFunc("/payment-methods", paymentHandler.AddPaymentMethod).Methods("POST")
	securedPayments.HandleFunc("/payment-methods", paymentHandler.GetPaymentMethods).Methods("GET")
	securedPayments.HandleFunc("/payment-methods/{id:[0-9]+}", paymentHandler.DeletePaymentMethod).Methods("DELETE")
}
