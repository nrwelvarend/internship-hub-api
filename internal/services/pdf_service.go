package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/jung-kurt/gofpdf/v2"
)

type PDFService struct {
	UploadDir string
}

func NewPDFService(uploadDir string) *PDFService {
	// Ensure directory exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}
	return &PDFService{UploadDir: uploadDir}
}

func (s *PDFService) GenerateCompletionLetter(result *models.InternshipResult) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "SURAT KETERANGAN SELESAI MAGANG", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Content
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 10, fmt.Sprintf("Menerangkan bahwa:\n\nNama: %s\nUniversitas: %s\nProgram Studi: %s\n\nTelah menyelesaikan Program Magang di %s pada posisi %s dengan capaian evaluasi akhir sebagai berikut:",
		result.User.Name, result.User.University, result.User.Major, result.Application.Vacancy.UnitKerja.Name, result.Application.Vacancy.Title), "", "L", false)

	pdf.Ln(5)

	// Scores table
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(100, 10, "Kategori Penilaian", "1", 0, "C", false, 0, "")
	pdf.CellFormat(40, 10, "Nilai", "1", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(100, 10, "Kehadiran", "1", 0, "L", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", result.AttendanceScore), "1", 1, "C", false, 0, "")

	pdf.CellFormat(100, 10, "Kinerja", "1", 0, "L", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", result.PerformanceScore), "1", 1, "C", false, 0, "")

	pdf.CellFormat(100, 10, "Laporan", "1", 0, "L", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", result.ReportScore), "1", 1, "C", false, 0, "")

	pdf.CellFormat(100, 10, "Kedisiplinan", "1", 0, "L", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", result.DisciplineScore), "1", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(100, 10, "Nilai Akhir", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", result.FinalScore), "1", 1, "C", false, 0, "")

	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 10)
	pdf.MultiCell(0, 5, "Catatan Reviewer: "+result.ReviewNotes, "", "L", false)

	pdf.Ln(20)
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 10, fmt.Sprintf("Dikeluarkan pada: %s", time.Now().Format("02 January 2006")), "", 1, "R", false, 0, "")
	pdf.Ln(15)
	pdf.CellFormat(0, 10, "( Administrator )", "", 1, "R", false, 0, "")

	filename := fmt.Sprintf("completion_%s.pdf", result.ApplicationID.String())
	path := filepath.Join(s.UploadDir, filename)
	err := pdf.OutputFileAndClose(path)
	return filename, err
}

func (s *PDFService) GenerateCertificate(result *models.InternshipResult) (string, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()

	// Luxury Border
	pdf.SetLineWidth(2)
	pdf.Rect(10, 10, 277, 190, "D")
	pdf.SetLineWidth(0.5)
	pdf.Rect(12, 12, 273, 186, "D")

	// Title
	pdf.SetFont("Arial", "B", 36)
	pdf.Ln(20)
	pdf.CellFormat(0, 20, "SERTIFIKAT MAGANG", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 18)
	pdf.CellFormat(0, 15, "Diberikan Kepada:", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(0, 51, 102) // Dark Blue
	pdf.CellFormat(0, 20, result.User.Name, "", 1, "C", false, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 16)
	pdf.MultiCell(0, 10, fmt.Sprintf("Atas dedikasi, kontribusi, dan kinerja selama mengikuti Program Magang\ndi %s sebagai %s.",
		result.Application.Vacancy.UnitKerja.Name, result.Application.Vacancy.Title), "", "C", false)

	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(0, 10, fmt.Sprintf("PREDIKAT: %s", s.getPredicate(result.FinalScore)), "", 1, "C", false, 0, "")

	pdf.Ln(20)
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 10, "Dikeluarkan oleh Internship Hub Global", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 10, time.Now().Format("02 January 2006"), "", 1, "C", false, 0, "")

	filename := fmt.Sprintf("certificate_%s.pdf", result.ApplicationID.String())
	path := filepath.Join(s.UploadDir, filename)
	err := pdf.OutputFileAndClose(path)
	return filename, err
}

func (s *PDFService) getPredicate(score float64) string {
	if score >= 85 {
		return "SANGAT BAIK"
	} else if score >= 75 {
		return "BAIK"
	} else if score >= 60 {
		return "CUKUP"
	}
	return "KURANG"
}
