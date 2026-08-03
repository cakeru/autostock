package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	analyticsHandler "github.com/cakeru/autostock/internal/analytics/handler"
	analyticsService "github.com/cakeru/autostock/internal/analytics/service"
	auditHandler "github.com/cakeru/autostock/internal/audit/handler"
	authHandler "github.com/cakeru/autostock/internal/auth/handler"
	authService "github.com/cakeru/autostock/internal/auth/service"
	backupsHandler "github.com/cakeru/autostock/internal/backups/handler"
	backupsService "github.com/cakeru/autostock/internal/backups/service"
	cashshiftHandler "github.com/cakeru/autostock/internal/cashshift/handler"
	"github.com/cakeru/autostock/internal/config"
	customerHandler "github.com/cakeru/autostock/internal/customer/handler"
	vehicleHandler "github.com/cakeru/autostock/internal/vehicle/handler"
	batchinstallHandler "github.com/cakeru/autostock/internal/batchinstall/handler"
	dashboardHandler "github.com/cakeru/autostock/internal/dashboard/handler"
	"github.com/cakeru/autostock/internal/database"
	depositHandler "github.com/cakeru/autostock/internal/deposit/handler"
	employeeHandler "github.com/cakeru/autostock/internal/employee/handler"
	expenseHandler "github.com/cakeru/autostock/internal/expense/handler"
	inventoryHandler "github.com/cakeru/autostock/internal/inventory/handler"
	invoiceHandler "github.com/cakeru/autostock/internal/invoice/handler"
	exportHandler "github.com/cakeru/autostock/internal/export/handler"
	updateHandler "github.com/cakeru/autostock/internal/update/handler"
	"github.com/cakeru/autostock/internal/middleware"
	purchaseorderHandler "github.com/cakeru/autostock/internal/purchaseorder/handler"
	returnsHandler "github.com/cakeru/autostock/internal/returns/handler"
	searchHandler "github.com/cakeru/autostock/internal/search/handler"
	servicejobHandler "github.com/cakeru/autostock/internal/servicejob/handler"
	settingsHandler "github.com/cakeru/autostock/internal/settings/handler"
	settingsService "github.com/cakeru/autostock/internal/settings/service"
	stocktakeHandler "github.com/cakeru/autostock/internal/stocktake/handler"
	"github.com/cakeru/autostock/internal/storage"
	supplierHandler "github.com/cakeru/autostock/internal/supplier/handler"
	telegramHandler "github.com/cakeru/autostock/internal/telegram/handler"
	telegramService "github.com/cakeru/autostock/internal/telegram/service"
	vehicleService "github.com/cakeru/autostock/internal/vehicle/service"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	zerolog.TimeFieldFormat = time.RFC3339
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer pool.Close()
	logger.Info().Msg("Connected to database")

	store, err := storage.New(cfg.Storage)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize storage")
	}
	logger.Info().Str("driver", cfg.Storage.Driver).Msg("Storage ready")

	authSvc := authService.NewService(pool, cfg.JWTSecret, cfg.JWTExpiry)
	settingsSvc := settingsService.NewService(pool)
	telegramSvc := telegramService.NewService(pool, analyticsService.NewService(pool), vehicleService.NewService(pool), cfg.DatabaseURL)

	authH := authHandler.NewHandler(authSvc)
	settingsH := settingsHandler.NewHandler(settingsSvc)
	telegramH := telegramHandler.NewHandler(telegramSvc)
	inventoryH := inventoryHandler.NewHandler(pool, store)
	customerH := customerHandler.NewHandler(pool)
	vehicleH := vehicleHandler.NewHandler(pool, store)
	batchinstallH := batchinstallHandler.NewHandler(pool)
	servicejobH := servicejobHandler.NewHandler(pool)
	invoiceH := invoiceHandler.NewHandler(pool, store)
	exportH := exportHandler.NewHandler(pool, cfg.BackupDir)
	updateH := updateHandler.NewHandler(cfg.UpdaterURL)
	backupSvc := backupsService.NewService(pool, cfg.BackupDir, cfg.DatabaseURL)
	backupH := backupsHandler.NewHandler(backupSvc)
	dashboardH := dashboardHandler.NewHandler(pool)
	analyticsH := analyticsHandler.NewHandler(pool)
	auditH := auditHandler.NewHandler(pool)
	cashshiftH := cashshiftHandler.NewHandler(pool)
	supplierH := supplierHandler.NewHandler(pool)
	stocktakeH := stocktakeHandler.NewHandler(pool)
	purchaseorderH := purchaseorderHandler.NewHandler(pool)
	employeeH := employeeHandler.NewHandler(pool)
	depositH := depositHandler.NewHandler(pool)
	returnsH := returnsHandler.NewHandler(pool)
	expenseH := expenseHandler.NewHandler(pool)
	searchH := searchHandler.NewHandler(pool)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Serve locally-stored uploads (no-op path when using the r2 driver).
	if cfg.Storage.Driver == "" || cfg.Storage.Driver == "local" {
		router.Static(cfg.Storage.PublicURL, cfg.Storage.UploadDir)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		})
	})
	router.GET("/ready", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	v1 := router.Group("/api/v1")
	{
		v1.POST("/auth/login", authH.Login)

		authed := v1.Group("")
		authed.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			authed.GET("/auth/me", authH.GetMe)
			authed.PUT("/auth/password", authH.ChangePassword)
		}

		admin := v1.Group("")
		admin.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		admin.Use(middleware.PermissionMiddleware("user:view"))
		{
			admin.GET("/users", authH.ListUsers)
			admin.POST("/users", middleware.PermissionMiddleware("user:create"), authH.CreateUser)
			admin.PUT("/users/:id", middleware.PermissionMiddleware("user:update"), authH.UpdateUser)
			admin.DELETE("/users/:id", middleware.PermissionMiddleware("user:delete"), authH.DeleteUser)

			admin.GET("/employees", employeeH.List)
			admin.GET("/employees/:id", employeeH.Get)
			admin.POST("/employees", middleware.PermissionMiddleware("user:create"), employeeH.Create)
			admin.PUT("/employees/:id", middleware.PermissionMiddleware("user:update"), employeeH.Update)
			admin.DELETE("/employees/:id", middleware.PermissionMiddleware("user:delete"), employeeH.Delete)
			admin.POST("/employees/:id/create-account", middleware.PermissionMiddleware("user:create"), employeeH.CreateAccount)
		}

		inv := v1.Group("/products")
		inv.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			inv.GET("", middleware.PermissionMiddleware("inventory:view"), inventoryH.List)
			inv.GET("/low-stock", middleware.PermissionMiddleware("inventory:view"), inventoryH.LowStock)
			inv.POST("/import", middleware.PermissionMiddleware("inventory:create"), inventoryH.Import)
			inv.GET("/:id", middleware.PermissionMiddleware("inventory:view"), inventoryH.Get)
			inv.POST("", middleware.PermissionMiddleware("inventory:create"), inventoryH.Create)
			inv.PUT("/:id", middleware.PermissionMiddleware("inventory:update"), inventoryH.Update)
			inv.DELETE("/:id", middleware.PermissionMiddleware("inventory:delete"), inventoryH.Delete)
			inv.POST("/:id/receive", middleware.PermissionMiddleware("inventory:update"), inventoryH.ReceiveStock)
			inv.POST("/:id/adjust", middleware.PermissionMiddleware("inventory:update"), inventoryH.AdjustStock)
			inv.GET("/:id/movements", middleware.PermissionMiddleware("inventory:view"), inventoryH.Movements)
			inv.GET("/:id/batches", middleware.PermissionMiddleware("inventory:view"), inventoryH.Batches)
			inv.POST("/:id/image", middleware.PermissionMiddleware("inventory:update"), inventoryH.UploadImage)
			inv.DELETE("/:id/image", middleware.PermissionMiddleware("inventory:update"), inventoryH.DeleteImage)
		}

		v1.GET("/batches/:batch_id/consumers", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("inventory:view"), inventoryH.BatchConsumers)

		stk := v1.Group("/stocktakes")
		stk.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			stk.GET("", middleware.PermissionMiddleware("inventory:view"), stocktakeH.List)
			stk.GET("/:id", middleware.PermissionMiddleware("inventory:view"), stocktakeH.Get)
			stk.POST("", middleware.PermissionMiddleware("inventory:update"), stocktakeH.Create)
			stk.POST("/:id/items", middleware.PermissionMiddleware("inventory:update"), stocktakeH.AddItem)
			stk.POST("/:id/cancel", middleware.PermissionMiddleware("inventory:update"), stocktakeH.Cancel)
			stk.POST("/:id/finalize", middleware.PermissionMiddleware("inventory:update"), stocktakeH.Finalize)
		}
		v1.PUT("/stocktakes/items/:item_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("inventory:update"), stocktakeH.SetCount)
		v1.DELETE("/stocktakes/items/:item_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("inventory:update"), stocktakeH.RemoveItem)

		po := v1.Group("/purchase-orders")
		po.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			po.GET("", middleware.PermissionMiddleware("inventory:view"), purchaseorderH.List)
			po.GET("/:id", middleware.PermissionMiddleware("inventory:view"), purchaseorderH.Get)
			po.POST("", middleware.PermissionMiddleware("inventory:update"), purchaseorderH.Create)
			po.PUT("/:id", middleware.PermissionMiddleware("inventory:update"), purchaseorderH.Update)
			po.POST("/:id/items", middleware.PermissionMiddleware("inventory:update"), purchaseorderH.AddItem)
			po.POST("/:id/place", middleware.PermissionMiddleware("inventory:update"), purchaseorderH.Place)
			po.POST("/:id/cancel", middleware.PermissionMiddleware("inventory:update"), purchaseorderH.Cancel)
			po.POST("/:id/receive", middleware.PermissionMiddleware("inventory:update"), purchaseorderH.Receive)
		}
		v1.DELETE("/purchase-orders/items/:item_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("inventory:update"), purchaseorderH.RemoveItem)

		cust := v1.Group("/customers")
		cust.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			cust.GET("", middleware.PermissionMiddleware("customer:view"), customerH.List)
			cust.GET("/:id", middleware.PermissionMiddleware("customer:view"), customerH.Get)
			cust.POST("", middleware.PermissionMiddleware("customer:create"), customerH.Create)
			cust.PUT("/:id", middleware.PermissionMiddleware("customer:update"), customerH.Update)
			cust.DELETE("/:id", middleware.PermissionMiddleware("customer:delete"), customerH.Delete)
			cust.GET("/:id/history", middleware.PermissionMiddleware("customer:view"), customerH.History)
			cust.GET("/:id/vehicles", middleware.PermissionMiddleware("customer:view"), customerH.ListVehicles)
			cust.POST("/:id/vehicles", middleware.PermissionMiddleware("customer:create"), customerH.CreateVehicle)
		}

		v1.PUT("/vehicles/:id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), customerH.UpdateVehicle)
		v1.DELETE("/vehicles/:id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:delete"), customerH.DeleteVehicle)

		veh := v1.Group("/vehicles")
		veh.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			veh.GET("/:id/profile", middleware.PermissionMiddleware("customer:view"), vehicleH.GetProfile)
			veh.GET("/:id/history", middleware.PermissionMiddleware("customer:view"), vehicleH.GetHistory)
			veh.GET("/:id/timeline", middleware.PermissionMiddleware("customer:view"), vehicleH.GetTimeline)
			veh.GET("/:id/records", middleware.PermissionMiddleware("customer:view"), vehicleH.ListRecords)
			veh.POST("/:id/records", middleware.PermissionMiddleware("customer:update"), vehicleH.CreateRecord)
			veh.GET("/:id/service-events", middleware.PermissionMiddleware("customer:view"), vehicleH.ListServiceEvents)
			veh.POST("/:id/service-events", middleware.PermissionMiddleware("customer:update"), vehicleH.CreateServiceEvent)
			veh.PUT("/:id/intervals", middleware.PermissionMiddleware("customer:update"), vehicleH.UpdateVehicleIntervals)
			veh.GET("/:id/wheel-services", middleware.PermissionMiddleware("customer:view"), vehicleH.ListWheelServices)
			veh.POST("/:id/wheel-services", middleware.PermissionMiddleware("customer:update"), vehicleH.CreateWheelService)
			veh.GET("/:id/tire-options", middleware.PermissionMiddleware("customer:view"), vehicleH.RecentTireOptions)
			veh.GET("/:id/parts", middleware.PermissionMiddleware("customer:view"), vehicleH.ListParts)
			veh.POST("/:id/parts", middleware.PermissionMiddleware("customer:update"), vehicleH.CreatePart)
			veh.GET("/:id/part-status", middleware.PermissionMiddleware("customer:view"), vehicleH.ListPartStatuses)
			veh.PUT("/:id/part-status", middleware.PermissionMiddleware("customer:update"), vehicleH.SetPartStatus)
			veh.GET("/:id/photos", middleware.PermissionMiddleware("customer:view"), vehicleH.ListGalleryPhotos)
			veh.POST("/:id/photos", middleware.PermissionMiddleware("customer:update"), vehicleH.AddGalleryPhoto)
			veh.POST("/:id/share-link", middleware.PermissionMiddleware("customer:update"), vehicleH.EnsureShareLink)
			veh.DELETE("/:id/share-link", middleware.PermissionMiddleware("customer:update"), vehicleH.RevokeShareLink)
		}
		// Public, token-authorized customer report — the only unauthenticated
		// vehicle route; the random share token is the authorization.
		v1.GET("/public/vehicle-report/:token", vehicleH.PublicReport)
		v1.DELETE("/vehicle-records/:record_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), vehicleH.DeleteRecord)
		v1.PUT("/vehicle-records/:record_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), vehicleH.UpdateRecord)
		v1.POST("/vehicle-records/:record_id/photos", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), vehicleH.AddPhoto)
		v1.DELETE("/vehicle-record-photos/:photo_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), vehicleH.DeletePhoto)
		v1.DELETE("/vehicle-service-events/:event_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), vehicleH.DeleteServiceEvent)
		v1.PUT("/vehicle-service-events/:event_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), vehicleH.UpdateServiceEvent)
		v1.DELETE("/vehicle-wheel-services/:service_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), vehicleH.DeleteWheelService)
		v1.PUT("/vehicle-wheel-services/:service_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), vehicleH.UpdateWheelService)
		v1.POST("/vehicle-wheel-services/:service_id/photos", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), vehicleH.AddWheelServicePhoto)
		v1.DELETE("/vehicle-parts/:part_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), vehicleH.DeletePart)
		v1.PUT("/vehicle-parts/:part_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), vehicleH.UpdatePart)
		v1.PUT("/vehicle-photos/:photo_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), vehicleH.UpdateGalleryPhoto)
		v1.DELETE("/vehicle-photos/:photo_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("customer:update"), vehicleH.DeleteGalleryPhoto)

		// Optional batch-scan traceability (mechanic scans a batch QR before
		// fitting a part). Low-privilege install:scan permission — a mechanic
		// device needs none of the sale/inventory permissions. Purely additive.
		bi := v1.Group("/batch-installs")
		bi.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		bi.Use(middleware.PermissionMiddleware("install:scan"))
		{
			bi.GET("/resolve", batchinstallH.Resolve)
			bi.POST("", batchinstallH.Record)
			bi.GET("/open-jobs", batchinstallH.OpenJobs)
			bi.GET("/mechanics", batchinstallH.Mechanics)
		}

		svcReminders := v1.Group("/service-reminders")
		svcReminders.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			svcReminders.GET("/due", middleware.PermissionMiddleware("customer:view"), vehicleH.ListDueForService)
			svcReminders.GET("/settings", middleware.PermissionMiddleware("settings:view"), vehicleH.GetIntervalSettings)
			svcReminders.PUT("/settings", middleware.PermissionMiddleware("settings:update"), vehicleH.UpdateIntervalSettings)
		}

		sj := v1.Group("/service-jobs")
		sj.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			sj.GET("", middleware.PermissionMiddleware("service:view"), servicejobH.List)
			sj.GET("/:id", middleware.PermissionMiddleware("service:view"), servicejobH.Get)
			sj.POST("", middleware.PermissionMiddleware("service:create"), servicejobH.Create)
			sj.PUT("/:id", middleware.PermissionMiddleware("service:update"), servicejobH.Update)
			sj.DELETE("/:id", middleware.PermissionMiddleware("service:delete"), servicejobH.Delete)
			sj.POST("/:id/items", middleware.PermissionMiddleware("service:update"), servicejobH.AddItem)
			sj.POST("/:id/complete", middleware.PermissionMiddleware("service:update"), servicejobH.Complete)
			sj.POST("/:id/approve-quote", middleware.PermissionMiddleware("service:update"), servicejobH.ApproveQuote)
			sj.POST("/:id/invoice", middleware.PermissionMiddleware("invoice:create"), invoiceH.CreateFromJob)
		}

		v1.DELETE("/service-jobs/items/:item_id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("service:update"), servicejobH.RemoveItem)

		invc := v1.Group("/invoices")
		invc.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			invc.GET("", middleware.PermissionMiddleware("invoice:view"), invoiceH.List)
			invc.GET("/:id", middleware.PermissionMiddleware("invoice:view"), invoiceH.Get)
			invc.POST("", middleware.PermissionMiddleware("invoice:create"), invoiceH.Create)
			invc.PUT("/:id", middleware.PermissionMiddleware("invoice:update"), invoiceH.Update)
			invc.POST("/:id/void", middleware.PermissionMiddleware("invoice:void"), invoiceH.Void)
			invc.POST("/:id/items", middleware.PermissionMiddleware("invoice:update"), invoiceH.AddItem)
			invc.PUT("/:id/items/:item_id", middleware.PermissionMiddleware("invoice:update"), invoiceH.UpdateItem)
			invc.DELETE("/:id/items/:item_id", middleware.PermissionMiddleware("invoice:update"), invoiceH.RemoveItem)
			invc.POST("/:id/payments", middleware.PermissionMiddleware("invoice:update"), invoiceH.RecordPayment)
			invc.PUT("/:id/payments/:payment_id", middleware.PermissionMiddleware("invoice:update"), invoiceH.UpdatePayment)
			invc.DELETE("/:id/payments/:payment_id", middleware.PermissionMiddleware("invoice:update"), invoiceH.DeletePayment)
			invc.POST("/:id/payments/:payment_id/proof", middleware.PermissionMiddleware("invoice:update"), invoiceH.UploadPaymentProof)
			invc.GET("/:id/payments", middleware.PermissionMiddleware("invoice:view"), invoiceH.ListPayments)
			invc.GET("/:id/returns", middleware.PermissionMiddleware("invoice:view"), returnsH.ListForInvoice)
		}

		v1.POST("/returns", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("invoice:update"), returnsH.Create)
		v1.DELETE("/returns/:id", middleware.AuthMiddleware(cfg.JWTSecret, pool), middleware.PermissionMiddleware("invoice:update"), returnsH.Undo)

		v1.GET("/search", middleware.AuthMiddleware(cfg.JWTSecret, pool), searchH.Search)

		exports := v1.Group("/exports")
		exports.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			exports.GET("/invoices", middleware.PermissionMiddleware("invoice:view"), exportH.ExportInvoices)
			exports.GET("/customers", middleware.PermissionMiddleware("customer:view"), exportH.ExportCustomers)
			exports.GET("/products", middleware.PermissionMiddleware("inventory:view"), exportH.ExportProducts)
		}

		upd := v1.Group("/updates")
		upd.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			upd.POST("/deploy", middleware.PermissionMiddleware("settings:update"), updateH.Deploy)
			upd.GET("/status", middleware.PermissionMiddleware("settings:view"), updateH.Status)
		}

		set := v1.Group("/settings")
		set.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			set.GET("", middleware.PermissionMiddleware("settings:view"), settingsH.GetSettings)
			set.PUT("", middleware.PermissionMiddleware("settings:update"), settingsH.UpdateSetting)
			set.GET("/exchange-rate", settingsH.GetExchangeRate)
			set.PUT("/exchange-rate", middleware.PermissionMiddleware("settings:update"), settingsH.UpdateExchangeRate)
			set.GET("/backup/latest", middleware.PermissionMiddleware("settings:view"), exportH.DownloadLatestBackup)
		}

		bks := v1.Group("/backup-schedules")
		bks.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			bks.GET("", middleware.PermissionMiddleware("settings:view"), backupH.List)
			bks.POST("", middleware.PermissionMiddleware("settings:update"), backupH.Create)
			bks.PUT("/:id", middleware.PermissionMiddleware("settings:update"), backupH.Update)
			bks.DELETE("/:id", middleware.PermissionMiddleware("settings:update"), backupH.Delete)
			bks.POST("/:id/run", middleware.PermissionMiddleware("settings:update"), backupH.RunNow)
			bks.GET("/:id/latest", middleware.PermissionMiddleware("settings:view"), backupH.DownloadLatest)
		}

		dash := v1.Group("/dashboard")
		dash.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		dash.Use(middleware.PermissionMiddleware("report:view"))
		{
			dash.GET("/summary", dashboardH.Summary)
			dash.GET("/daily-revenue", dashboardH.DailyRevenue)
			dash.GET("/day-close", dashboardH.DayClose)
			dash.GET("/profit", dashboardH.Profit)
		}

		ana := v1.Group("/analytics")
		ana.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		ana.Use(middleware.PermissionMiddleware("report:view"))
		{
			ana.GET("/sales", analyticsH.Sales)
			ana.GET("/receivables", analyticsH.Receivables)
			ana.GET("/inventory", analyticsH.Inventory)
			ana.GET("/customers", analyticsH.Customers)
			ana.GET("/technicians", analyticsH.Technicians)
			ana.GET("/pnl", analyticsH.PnL)
		}

		exp := v1.Group("/expenses")
		exp.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		exp.Use(middleware.PermissionMiddleware("report:view"))
		{
			exp.GET("", expenseH.List)
			exp.POST("", expenseH.Create)
			exp.PUT("/:id", expenseH.Update)
			exp.DELETE("/:id", expenseH.Delete)
		}

		aud := v1.Group("/audit-logs")
		aud.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		aud.Use(middleware.PermissionMiddleware("report:view"))
		{
			aud.GET("", auditH.List)
		}

		cash := v1.Group("/cash-shifts")
		cash.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		cash.Use(middleware.PermissionMiddleware("invoice:create"))
		{
			cash.GET("", cashshiftH.List)
			cash.GET("/current", cashshiftH.Current)
			cash.POST("/open", cashshiftH.Open)
			cash.POST("/close", cashshiftH.Close)
		}

		sup := v1.Group("/suppliers")
		sup.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			sup.GET("", middleware.PermissionMiddleware("inventory:view"), supplierH.List)
			sup.GET("/:id", middleware.PermissionMiddleware("inventory:view"), supplierH.Get)
			sup.GET("/:id/purchases", middleware.PermissionMiddleware("inventory:view"), supplierH.Purchases)
			sup.POST("", middleware.PermissionMiddleware("inventory:update"), supplierH.Create)
			sup.PUT("/:id", middleware.PermissionMiddleware("inventory:update"), supplierH.Update)
			sup.DELETE("/:id", middleware.PermissionMiddleware("inventory:update"), supplierH.Delete)
			sup.POST("/:id/pay", middleware.PermissionMiddleware("inventory:update"), supplierH.Pay)
		}

		dep := v1.Group("/deposits")
		dep.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			dep.GET("", middleware.PermissionMiddleware("invoice:view"), depositH.List)
			dep.POST("", middleware.PermissionMiddleware("invoice:create"), depositH.Create)
			dep.PUT("/:id", middleware.PermissionMiddleware("invoice:update"), depositH.Update)
			dep.POST("/:id/apply", middleware.PermissionMiddleware("invoice:update"), depositH.Apply)
			dep.POST("/:id/refund", middleware.PermissionMiddleware("invoice:update"), depositH.Refund)
		}

		tg := v1.Group("/telegram")
		tg.Use(middleware.AuthMiddleware(cfg.JWTSecret, pool))
		{
			tg.GET("/channels", middleware.PermissionMiddleware("settings:view"), telegramH.GetChannels)
			tg.PUT("/channels", middleware.PermissionMiddleware("settings:update"), telegramH.SaveChannels)
			tg.GET("/routes", middleware.PermissionMiddleware("settings:view"), telegramH.GetRoutes)
			tg.PUT("/routes", middleware.PermissionMiddleware("settings:update"), telegramH.SaveRoutes)
			tg.POST("/test-send", middleware.PermissionMiddleware("settings:update"), telegramH.TestSend)
			tg.POST("/trigger", middleware.PermissionMiddleware("settings:update"), telegramH.Trigger)
			// On-demand document forward (invoice / vehicle report PDF). Gated on
			// invoice:view so counter staff — not just settings admins — can send.
			tg.POST("/send-document", middleware.PermissionMiddleware("invoice:view"), telegramH.SendDocument)
		}
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info().Str("port", cfg.Port).Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server failed")
		}
	}()

	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()
	go telegramSvc.RunDeliveryLoop(bgCtx)
	go telegramSvc.RunScheduler(bgCtx)
	go backupSvc.Run(bgCtx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal().Err(err).Msg("Server shutdown failed")
	}
	logger.Info().Msg("Server stopped")
}
