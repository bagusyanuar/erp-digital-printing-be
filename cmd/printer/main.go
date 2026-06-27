//go:build windows

package main

import (
	"log"

	"github.com/alexbrainman/printer"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

type PrintPayload struct {
	PrinterName string `json:"printer_name"`
	RawData     string `json:"raw_data"`
	// Nanti bisa ditambahkan field lain sesuai kebutuhan struk
}

func main() {
	// 1. Setup Fiber App
	app := fiber.New()

	// 2. Middleware
	app.Use(logger.New())

	// Wajib buka CORS agar frontend online (React) bisa hit ke localhost
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // Izinkan semua origin untuk development
		AllowHeaders: []string{"Origin, Content-Type, Accept"},
		AllowMethods: []string{"GET, POST, OPTIONS"},
	}))

	// 3. Routing
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "ERP Printer Agent is running",
		})
	})

	app.Post("/print", func(c fiber.Ctx) error {
		var payload PrintPayload

		if err := c.Bind().Body(&payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request payload",
			})
		}

		// Buka koneksi ke printer Windows
		p, err := printer.Open(payload.PrinterName)
		if err != nil {
			log.Printf("Gagal membuka printer %s: %v\n", payload.PrinterName, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal membuka printer: " + err.Error(),
			})
		}
		defer p.Close()

		// Kirim raw data (ESCPOS)
		_, err = p.Write([]byte(payload.RawData))
		if err != nil {
			log.Printf("Gagal mengirim data ke printer %s: %v\n", payload.PrinterName, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal mengirim data ke printer: " + err.Error(),
			})
		}

		log.Printf("Sukses mengirim perintah print ke printer: %s\n", payload.PrinterName)

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Print job success",
		})
	})

	app.Get("/print-test", func(c fiber.Ctx) error {
		log.Println("Tesss... Menerima perintah test print dari Frontend!")

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Test print berhasil diterima oleh agen lokal!",
			"data": "========================\n" +
				"     TEST PRINT ERP     \n" +
				"========================\n" +
				"Koneksi dari web ke printer berjalan lancar.\n",
		})
	})

	// 4. Start Server (jalan di port 9000 agar tidak bentrok dengan backend utama 8080)
	port := ":9000"
	log.Printf("Starting Local Print Agent on %s...\n", port)
	if err := app.Listen(port); err != nil {
		log.Fatalf("Error starting print agent: %v", err)
	}
}
