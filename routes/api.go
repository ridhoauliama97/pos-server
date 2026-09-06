package routes

import (
	"github.com/goravel/framework/contracts/route"

	"github.com/ridhoauliama97/pos-server/app/facades"
	"github.com/ridhoauliama97/pos-server/app/http/controllers/api/v1"
	"github.com/ridhoauliama97/pos-server/app/http/middleware"
	"github.com/ridhoauliama97/pos-server/app/models"
)

func Api() {
	authController := v1.NewAuthController()
	userController := v1.NewUserController()
	customerController := v1.NewCustomerController()
	categoryController := v1.NewCategoryController()
	productController := v1.NewProductController()
	stockController := v1.NewStockController()
	transactionController := v1.NewTransactionController()
	receiptController := v1.NewReceiptController()
	reportController := v1.NewReportController()

	facades.Route().Group(func(router route.Router) {
		api := router.Prefix("/api/v1")

		api.Post("/auth/login", authController.Login)
		api.Middleware(middleware.NewAuth()).Post("/auth/logout", authController.Logout)
		api.Middleware(middleware.NewAuth()).Post("/auth/refresh", authController.Refresh)

		userRouter := api.Middleware(middleware.NewAuth(), middleware.NewRole(models.RoleOwner, models.RoleAdmin)).Prefix("/users")

		userRouter.Get("", userController.Index)
		userRouter.Post("", userController.Store)
		userRouter.Put("/{id}", userController.Update)
		userRouter.Delete("/{id}", userController.Destroy)

		categoryRouter := api.Middleware(middleware.NewAuth()).Prefix("/categories")

		categoryRouter.Get("", categoryController.Index)
		categoryRouter.Middleware(middleware.NewRole(models.RoleOwner, models.RoleAdmin)).Post("", categoryController.Store)
		categoryRouter.Middleware(middleware.NewRole(models.RoleOwner, models.RoleAdmin)).Put("/{id}", categoryController.Update)
		categoryRouter.Middleware(middleware.NewRole(models.RoleOwner, models.RoleAdmin)).Delete("/{id}", categoryController.Destroy)

		customerRouter := api.Middleware(middleware.NewAuth()).Prefix("/customers")

		customerRouter.Get("", customerController.Index)
		customerRouter.Get("/{id}", customerController.Show)
		customerRouter.Get("/{id}/transactions", customerController.Transactions)
		customerRouter.Post("", customerController.Store)
		customerRouter.Middleware(middleware.NewRole(models.RoleOwner, models.RoleAdmin)).Put("/{id}", customerController.Update)
		customerRouter.Middleware(middleware.NewRole(models.RoleOwner, models.RoleAdmin)).Delete("/{id}", customerController.Destroy)

		productRouter := api.Middleware(middleware.NewAuth()).Prefix("/products")

		productRouter.Get("", productController.Index)
		productRouter.Middleware(middleware.NewRole(models.RoleOwner, models.RoleAdmin)).Post("", productController.Store)
		productRouter.Middleware(middleware.NewRole(models.RoleOwner, models.RoleAdmin)).Put("/{id}", productController.Update)
		productRouter.Middleware(middleware.NewRole(models.RoleOwner, models.RoleAdmin)).Delete("/{id}", productController.Destroy)

		stockRouter := api.Middleware(middleware.NewAuth()).Prefix("/stocks")

		stockRouter.Get("", stockController.Index)
		stockRouter.Get("/{productVariantId}", stockController.Show)
		stockRouter.Middleware(middleware.NewRole(models.RoleOwner, models.RoleAdmin)).Post("/adjust", stockController.Adjust)

		transactionRouter := api.Middleware(middleware.NewAuth()).Prefix("/transactions")

		transactionRouter.Post("", transactionController.Store)
		transactionRouter.Middleware(middleware.NewAuth()).Post("/bulk-sync", transactionController.BulkSync)
		transactionRouter.Get("/{id}/receipt", receiptController.Show)
		transactionRouter.Middleware(middleware.NewRole(models.RoleOwner, models.RoleAdmin)).Post("/{id}/void", transactionController.Void)
		transactionRouter.Middleware(middleware.NewRole(models.RoleOwner, models.RoleAdmin)).Post("/{id}/refund", transactionController.Refund)

		reportRouter := api.Middleware(middleware.NewAuth()).Prefix("/reports")

		reportRouter.Get("/daily", reportController.Daily)
		reportRouter.Get("/kasir", reportController.ByKasir)
		reportRouter.Get("/shift", reportController.Shift)
		reportRouter.Get("/products", reportController.Products)
	})
}
