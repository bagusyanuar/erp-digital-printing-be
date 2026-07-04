# Rencana Implementasi Export Excel - ERP Digital Printing

Dokumen ini mencakup analisis dan rencana pemilihan library Go untuk kebutuhan export data laporan ke format Excel (.xlsx), terutama untuk Report Selling/Dashboard.

## 🛠 Pilihan Library
Diputuskan untuk menggunakan library **Excelize (v2)**.

### Alasan Pemilihan:
1. **Dukungan Multiple Sheets**: Memungkinkan pemisahan data ringkasan, detail penjualan, dan data pelanggan dalam satu file.
2. **Kaya Fitur Styling**: Mendukung formatting table, alignment, borders, format mata uang (IDR), dan pewarnaan status order/pembayaran.
3. **Kompatibilitas Tinggi**: Menghasilkan file Excel standard yang kompatibel dengan Microsoft Excel, Google Sheets, dan LibreOffice tanpa rusak.

---

## 📋 Skema Output Sheet (Rencana)

Dalam satu file `.xlsx`, data akan dibagi menjadi beberapa sheet berikut:

1. **Sheet 1: Ringkasan Laporan**
   - Ringkasan omset total, jumlah order masuk, status pembayaran (Lunas, DP, Piutang).
   - Bentuk data aggregat/chart sederhana.

2. **Sheet 2: Detail Transaksi**
   - Log order secara terperinci (No Invoice, Tanggal, Pelanggan, Tipe Pelanggan, Total Pembayaran, Status).

3. **Sheet 3: Detail Produk Terjual**
   - Breakdown item pekerjaan (Cetak Banner, Kartu Nama, Laser Cut, dll) beserta volume/ukuran.

---

## 🛠 Contoh Struktur Code Go

```go
package report

import (
    "github.com/xuri/excelize/v2"
)

// GenerateSalesReport menghasilkan file excel stream atau disimpan langsung
func GenerateSalesReport(data any) (*excelize.File, error) {
    f := excelize.NewFile()
    
    // 1. Setup Sheet Ringkasan
    f.SetSheetName("Sheet1", "Ringkasan")
    f.SetCellValue("Ringkasan", "A1", "Ringkasan Laporan Penjualan")
    
    // 2. Setup Sheet Detail
    f.NewSheet("Detail Transaksi")
    f.SetCellValue("Detail Transaksi", "A1", "No Invoice")
    f.SetCellValue("Detail Transaksi", "B1", "Nama Pelanggan")
    
    return f, nil
}
```
