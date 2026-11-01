//go:build windows

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/alexbrainman/printer"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/kardianos/service"
	"gopkg.in/natefinch/lumberjack.v2"
)

type PrintPayload struct {
	PrinterName string `json:"printer_name"`
	RawData     string `json:"raw_data"`
	// Nanti bisa ditambahkan field lain sesuai kebutuhan struk
}

// program mengimplementasikan service.Interface milik kardianos/service
type program struct {
	app       *fiber.App
	logWriter io.Writer
}

func (p *program) Start(s service.Service) error {
	// 1. Setup Fiber App (dibuat di sini agar Stop tidak race dengan goroutine)
	p.app = fiber.New()

	// Start tidak boleh blocking, server dijalankan di goroutine
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	log.Println("Sinyal terminasi diterima. Menutup Print Agent...")
	if p.app != nil {
		return p.app.Shutdown()
	}
	return nil
}

func (p *program) run() {
	// 2. Middleware
	p.app.Use(logger.New(logger.Config{
		Stream: p.logWriter, // Log hit API FE juga masuk ke file
	}))

	// Wajib buka CORS agar frontend online (React) bisa hit ke localhost
	p.app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // Izinkan semua origin untuk development
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		// Chrome Private Network Access: halaman HTTPS publik yang hit localhost
		AllowPrivateNetwork: true,
	}))

	// 3. Routing
	p.app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "ERP Printer Agent is running",
		})
	})

	p.app.Post("/print", func(c fiber.Ctx) error {
		var payload PrintPayload

		if err := c.Bind().Body(&payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request payload",
			})
		}

		if payload.PrinterName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "printer_name wajib diisi",
			})
		}

		// Buka koneksi ke printer Windows
		log.Printf("Membuka koneksi printer: %s...\n", payload.PrinterName)
		pr, err := printer.Open(payload.PrinterName)
		if err != nil {
			log.Printf("Gagal membuka printer %s: %v\n", payload.PrinterName, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal membuka printer: " + err.Error(),
			})
		}
		defer func() {
			log.Printf("Menutup koneksi printer: %s\n", payload.PrinterName)
			pr.Close()
		}()

		err = pr.StartRawDocument("ERP Print Job")
		if err != nil {
			log.Printf("Gagal memulai print job %s: %v\n", payload.PrinterName, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal memulai print job: " + err.Error(),
			})
		}

		defer func() {
			_ = pr.EndDocument()
		}()

		// Kirim raw data (ESCPOS)
		_, err = pr.Write([]byte(payload.RawData))
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

	p.app.Get("/print-test", func(c fiber.Ctx) error {
		log.Println("Tesss... Menerima perintah test print dari Frontend!")

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Test print berhasil diterima oleh agen lokal!",
			"data": "\x1b\x40" + // Initialize printer
				"\x1b\x61\x01" + // Align Center
				"================================================\n" +
				"                 TEST PRINT ERP                 \n" +
				"================================================\n" +
				"\x1b\x61\x00" + // Align Left
				"Printer Model : POS80AB\n" +
				"Paper Width   : 80mm\n" +
				"Command       : ESC/POS\n\n" +
				"Koneksi dari web ke printer berjalan lancar.\n\n\n\n" +
				"\x1d\x56\x00", // Cut paper
		})
	})

	// 4. Start Server (jalan di port 9876 agar tidak bentrok dengan backend utama 8080 atau php-fpm)
	port := os.Getenv("PRINTER_PORT")
	if port == "" {
		port = "9876" // Default port baru
	}

	// Bind hanya ke loopback agar device lain di LAN tidak bisa kirim print job
	log.Printf("Starting Local Print Agent on 127.0.0.1:%s...\n", port)
	if err := p.app.Listen("127.0.0.1:"+port, fiber.ListenConfig{
		DisableStartupMessage: !service.Interactive(),
	}); err != nil {
		log.Printf("Print Agent berhenti: %v\n", err)
	}

	log.Println("Print Agent berhasil ditutup dengan aman.")
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

	// Saat jalan sebagai service tidak ada console, jadi stdout tidak valid.
	// io.MultiWriter berhenti di writer pertama yang error, maka log hanya ke file.
	var logWriter io.Writer = fileLogger
	if service.Interactive() {
		logWriter = io.MultiWriter(os.Stdout, fileLogger)
	}
	log.SetOutput(logWriter)
	log.SetFlags(log.Ldate | log.Ltime) // Format: 2026/06/27 10:00:00

	svcConfig := &service.Config{
		Name:        "ERPPrinterAgent",
		DisplayName: "ERP Printer Agent",
		Description: "Local print agent untuk mencetak struk ERP ke printer thermal (ESC/POS).",
		Option: service.KeyValue{
			"StartType":              "automatic",
			"OnFailure":              "restart",
			"OnFailureDelayDuration": "5s",
		},
	}

	prg := &program{logWriter: logWriter}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatalf("Gagal membuat service: %v", err)
	}

	// Perintah kontrol: printer-agent.exe install|uninstall|start|stop|restart
	if len(os.Args) > 1 {
		action := os.Args[1]
		if err := service.Control(s, action); err != nil {
			fmt.Printf("Gagal menjalankan perintah %q: %v\nPerintah valid: %q\n", action, err, service.ControlAction)
			os.Exit(1)
		}
		fmt.Printf("Perintah %q berhasil dijalankan.\n", action)
		return
	}

	log.Println("==================================================")
	log.Println("Memulai Print Agent ERP...")
	log.Printf("Log aplikasi disimpan di: %s\n", logFilePath)

	// Run blocking: dipanggil oleh Windows Service Manager, atau mode console saat dijalankan manual
	if err := s.Run(); err != nil {
		log.Printf("Print Agent error: %v\n", err)
	}
}
