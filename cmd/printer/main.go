//go:build windows

package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/alexbrainman/printer"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"gopkg.in/natefinch/lumberjack.v2"
)

type PrintPayload struct {
	PrinterName string `json:"printer_name"`
	RawData     string `json:"raw_data"`
	// Nanti bisa ditambahkan field lain sesuai kebutuhan struk
}

func main() {
	// 0. Setup File Logging (simpan sejajar dengan .exe)
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Gagal mendapatkan path executable: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	logFilePath := filepath.Join(exeDir, "logs", "printer_agent.log")

	fileLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    10, // megabytes sebelum dirotasi
		MaxBackups: 5,  // maksimal simpan 5 file backup log
		MaxAge:     30, // hari
		Compress:   true,
	}
	defer fileLogger.Close()

	// Tulis log ke Terminal AND ke File
	multiWriter := io.MultiWriter(os.Stdout, fileLogger)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime) // Format: 2026/06/27 10:00:00

	log.Println("==================================================")
	log.Println("Memulai Print Agent ERP...")
	log.Printf("Log aplikasi disimpan di: %s\n", logFilePath)

	// 1. Setup Fiber App
	app := fiber.New()

	// 2. Middleware
	app.Use(logger.New(logger.Config{
		Stream: multiWriter, // Log hit API FE juga masuk ke file
	}))

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
		log.Printf("Membuka koneksi printer: %s...\n", payload.PrinterName)
		p, err := printer.Open(payload.PrinterName)
		if err != nil {
			log.Printf("Gagal membuka printer %s: %v\n", payload.PrinterName, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal membuka printer: " + err.Error(),
			})
		}
		defer func() {
			log.Printf("Menutup koneksi printer: %s\n", payload.PrinterName)
			p.Close()
		}()

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
	
	// Graceful Shutdown Setup
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		log.Println("Sinyal terminasi diterima. Menutup Print Agent...")
		_ = app.Shutdown()
	}()

	log.Printf("Starting Local Print Agent on %s...\n", port)
	if err := app.Listen(port); err != nil {
		log.Printf("Print Agent berhenti: %v\n", err)
	}
	
	log.Println("Print Agent berhasil ditutup dengan aman.")
}
