package test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"booking-service/internal/app"
	"booking-service/internal/config"
	"booking-service/internal/model"
	"booking-service/internal/router"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const testWebhookSecret = "docs-test-webhook-secret"

var (
	setupOnce   sync.Once
	setupDB     *gorm.DB
	setupSchema string
	setupErr    error
)

func integrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	setupOnce.Do(func() {
		_ = godotenv.Load("../.env", ".env")
		cfg := config.Load()
		cfg.DB = docsTestDBConfig(cfg.DB)

		adminDB, err := gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{})
		if err != nil {
			setupErr = fmt.Errorf("connect PostgreSQL: %w", err)
			return
		}
		sqlDB, err := adminDB.DB()
		if err != nil {
			setupErr = fmt.Errorf("open PostgreSQL connection: %w", err)
			return
		}
		if err := sqlDB.Ping(); err != nil {
			setupErr = fmt.Errorf("ping PostgreSQL: %w", err)
			return
		}

		setupSchema = fmt.Sprintf("docs_test_%d", time.Now().UnixNano())
		if err := adminDB.Exec(`CREATE SCHEMA "` + setupSchema + `"`).Error; err != nil {
			setupErr = fmt.Errorf("create test schema: %w", err)
			return
		}

		scopedDSN := cfg.DB.DSN() + " search_path=" + setupSchema
		setupDB, err = gorm.Open(postgres.Open(scopedDSN), &gorm.Config{})
		if err != nil {
			setupErr = fmt.Errorf("connect test schema: %w", err)
			return
		}
		if err := applyMigrations(setupDB); err != nil {
			setupErr = err
		}
	})

	if setupErr != nil {
		t.Skipf("integration test requires PostgreSQL: %v", setupErr)
	}
	return setupDB
}

func docsTestDBConfig(cfg config.DBConfig) config.DBConfig {
	if value := os.Getenv("TEST_DB_HOST"); value != "" {
		cfg.Host = value
	} else if cfg.Host == "postgres" {
		cfg.Host = "localhost"
	}
	if value := os.Getenv("TEST_DB_PORT"); value != "" {
		cfg.Port = value
	}
	if value := os.Getenv("TEST_DB_USER"); value != "" {
		cfg.User = value
	}
	if value := os.Getenv("TEST_DB_PASSWORD"); value != "" {
		cfg.Password = value
	}
	if value := os.Getenv("TEST_DB_NAME"); value != "" {
		cfg.Name = value
	}
	if value := os.Getenv("TEST_DB_SSLMODE"); value != "" {
		cfg.SSLMode = value
	}
	return cfg
}

func applyMigrations(db *gorm.DB) error {
	files, err := filepath.Glob("../migration/*.up.sql")
	if err != nil {
		return fmt.Errorf("find migrations: %w", err)
	}
	sort.Strings(files)
	for _, file := range files {
		query, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}
		if err := db.Exec(string(query)).Error; err != nil {
			return fmt.Errorf("apply migration %s: %w", file, err)
		}
	}
	return nil
}

type testFixture struct {
	DB       *gorm.DB
	App      *app.Container
	Router   http.Handler
	Event    model.Event
	Category model.TicketCategory
	Stock    model.TicketStock
}

func newFixture(t *testing.T, totalStock int) *testFixture {
	t.Helper()

	db := integrationDB(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	event := model.Event{
		Slug:         "docs-test-" + suffix,
		Name:         "Docs Test Event",
		VenueName:    "Test Venue",
		VenueAddress: "Test Address",
		StartsAt:     time.Now().Add(24 * time.Hour),
		Status:       "published",
	}
	if err := db.Create(&event).Error; err != nil {
		t.Fatalf("create event fixture: %v", err)
	}

	category := model.TicketCategory{
		EventID:       event.ID,
		Name:          "General Admission",
		Price:         100_000,
		MaxPerBooking: 10,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category fixture: %v", err)
	}

	stock := model.TicketStock{
		TicketCategoryID:  category.ID,
		TotalQuantity:     totalStock,
		AvailableQuantity: totalStock,
		Version:           1,
	}
	if err := db.Create(&stock).Error; err != nil {
		t.Fatalf("create stock fixture: %v", err)
	}

	container := app.NewContainer(db, testWebhookSecret)
	return &testFixture{
		DB:       db,
		App:      container,
		Router:   router.New(container, []string{"*"}),
		Event:    event,
		Category: category,
		Stock:    stock,
	}
}

func (f *testFixture) createReadyWaitingRoom(t *testing.T, token string) model.WaitingRoom {
	t.Helper()
	checkoutToken := "checkout-" + token
	expiresAt := time.Now().Add(10 * time.Minute)
	waitingRoom := model.WaitingRoom{
		EventID:          f.Event.ID,
		EventName:        f.Event.Name,
		TicketCategoryID: f.Category.ID,
		QueueToken:       "queue-" + token,
		CheckoutToken:    &checkoutToken,
		Status:           model.WaitingRoomStatusReady,
		ExpiredAt:        &expiresAt,
	}
	if err := f.DB.Create(&waitingRoom).Error; err != nil {
		t.Fatalf("create waiting room fixture: %v", err)
	}
	return waitingRoom
}

func (f *testFixture) createPendingBooking(t *testing.T, quantity int) model.Booking {
	t.Helper()

	guest := model.Guest{
		Name:    "Payment Test Guest",
		Email:   fmt.Sprintf("payment-%d@example.com", time.Now().UnixNano()),
		Phone:   "-",
		Address: "-",
	}
	if err := f.DB.Create(&guest).Error; err != nil {
		t.Fatalf("create guest fixture: %v", err)
	}

	totalPrice := f.Category.Price * int64(quantity)
	booking := model.Booking{
		BookingCode:       fmt.Sprintf("BKG-TEST-%d", time.Now().UnixNano()),
		GuestID:           guest.ID,
		GuestName:         guest.Name,
		GuestEmail:        guest.Email,
		GuestPhone:        guest.Phone,
		GuestAddress:      guest.Address,
		EventID:           f.Event.ID,
		EventSlug:         f.Event.Slug,
		EventName:         f.Event.Name,
		EventVenueName:    f.Event.VenueName,
		EventVenueAddress: f.Event.VenueAddress,
		EventStartsAt:     f.Event.StartsAt,
		Status:            "pending_payment",
		TotalTicket:       quantity,
		TotalPrice:        totalPrice,
		ExpiresAt:         time.Now().Add(15 * time.Minute),
	}
	if err := f.DB.Create(&booking).Error; err != nil {
		t.Fatalf("create booking fixture: %v", err)
	}

	item := model.BookingItem{
		BookingID:          booking.ID,
		TicketCategoryID:   f.Category.ID,
		TicketCategoryName: f.Category.Name,
		Quantity:           quantity,
		UnitPrice:          f.Category.Price,
		SubtotalPrice:      totalPrice,
	}
	if err := f.DB.Create(&item).Error; err != nil {
		t.Fatalf("create booking item fixture: %v", err)
	}

	if err := f.DB.Model(&model.TicketStock{}).
		Where("id = ?", f.Stock.ID).
		Updates(map[string]any{
			"available_quantity": f.Stock.TotalQuantity - quantity,
			"reserved_quantity":  quantity,
		}).Error; err != nil {
		t.Fatalf("reserve stock fixture: %v", err)
	}
	return booking
}

type paymentWebhookResponse struct {
	Success bool `json:"success"`
	Data    struct {
		PaymentTransactionID int64  `json:"payment_transaction_id"`
		TransactionCode      string `json:"transaction_code"`
		BookingID            int64  `json:"booking_id"`
		Duplicate            bool   `json:"duplicate"`
	} `json:"data"`
}

func sendPaymentWebhook(
	t *testing.T,
	handler http.Handler,
	payload map[string]any,
) (int, paymentWebhookResponse) {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payment webhook: %v", err)
	}
	mac := hmac.New(sha256.New, []byte(testWebhookSecret))
	_, _ = mac.Write(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Payment-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	var response paymentWebhookResponse
	if recorder.Body.Len() > 0 {
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode payment webhook response %q: %v", recorder.Body.String(), err)
		}
	}
	return recorder.Code, response
}
