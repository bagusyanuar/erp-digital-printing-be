# Printer Architecture Planning

## Overview
Dokumen ini menjelaskan arsitektur dan alur pencetakan (printing) struk kasir dari sistem Web POS (Point of Sales) ERP Digital Printing ke printer thermal offline yang terhubung ke komputer kasir.

## Architecture Flow (Opsi Terpilih)
Sistem menggunakan pendekatan **Direct Localhost** dimana Frontend (FE) bertindak sebagai eksekutor utama (orchestrator).

**Skema Flow:**
`FE (React/Browser)` -> `BE Web (Cloud)` -> `FE (React/Browser)` -> `Print Agent (.exe di Localhost)` -> `Hardware Printer`

### Detail Langkah:
1. **Simpan/Ambil Data (FE -> BE)**
   Frontend mengirimkan data transaksi (POST) atau mengambil data struk (GET) dari Backend Web (Cloud) untuk memproses validasi dan menyimpan riwayat transaksi.
2. **Format Struk (di FE)**
   Data JSON dari BE Web diterima oleh FE. FE kemudian merangkai dan memformat data tersebut menjadi teks struk mentah. Format ini bisa berupa Plain Text biasa atau menggunakan kode ESC/POS jika butuh styling hardware (seperti text tebal, perataan tengah, pemotong kertas otomatis).
3. **Eksekusi Print (FE -> Local Print Agent)**
   FE mengirimkan (melalui HTTP POST) format teks struk mentah tadi ke `http://localhost:9000/print`. API ini dilayani oleh Print Agent (aplikasi Golang `.exe` yang berjalan di komputer lokal kasir).
4. **Cetak ke Hardware (Local Print Agent -> Printer)**
   Print Agent menerima request dan meneruskannya secara langsung ke driver printer hardware kasir tanpa adanya *delay* jaringan internet.

## Kenapa Memilih Pendekatan Ini?
* **Responsif:** Tidak ada *delay* jaringan saat mencetak karena FE langsung hit ke localhost.
* **Aman & Simple:** Menghindari masalah NAT/Firewall. Backend di Cloud (VPS) tidak perlu menembus router komputer kasir lokal untuk mengirim perintah print.
* **Standar Web POS:** Merupakan cara yang paling umum dan stabil dipakai pada berbagai sistem Web POS modern.

## Konfigurasi Teknis Print Agent
* **OS Target:** Windows (karena package `github.com/alexbrainman/printer` berjalan dengan driver Windows).
* **CORS:** Wajib diaktifkan di sisi Print Agent (`AllowOrigins: "*"`) agar FE (React) tidak diblokir oleh browser saat melakukan request ke localhost.
* **Build Command (dari Mac/Linux ke Windows):**
  ```bash
  GOOS=windows GOARCH=amd64 go build -o printer_agent.exe ./cmd/printer
  ```

## Format Data (FE -> Print Agent)
Data yang dikirimkan dari Frontend ke Local Print Agent (`http://localhost:9000/print`) berupa **JSON Object** dengan format:

```json
{
  "printer_name": "POS58",
  "raw_data": "========================\nTOKO ERP DIGITAL PRINTING\n========================\nTotal: Rp 150.000\nTerima Kasih!\n\n"
}
```

### Detail Field
* **`printer_name`**: Nama printer persis seperti yang terdaftar di Windows kasir (*Control Panel -> Devices and Printers*). Dapat dikonfigurasi secara dinamis dari FE.
* **`raw_data`**: Berisi teks struk yang akan dicetak. Bisa murni berupa teks biasa berbaris (dipisah dengan `\n`), ATAU bisa juga dicampur dengan karakter khusus / perintah HEX **ESC/POS**.

### Integrasi ESC/POS
Jika hardware printer mendukung (seperti auto-cutter, rata tengah, teks tebal), FE bertugas merakit string `raw_data` dengan kombinasi teks dan *Escape sequence*. 
* **Contoh Init Printer:** `\x1B\x40`
* **Contoh Auto-Cutter:** `\x1D\x56\x00`
Print Agent tidak melakukan manipulasi apapun pada data ini, murni meneruskan *(pass-through)* string `raw_data` langsung ke driver printer.
