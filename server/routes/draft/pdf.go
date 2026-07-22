package draft

import (
	"appeals/db"
	"appeals/utilities"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf/v2"
)

type PartiesMetadata struct {
	Jurisdiction string    `json:"jurisdiction"`
	CaseNumber   string    `json:"caseNumber"`
	Appellant    string    `json:"appellant"`
	Appellee     string    `json:"appellee"`
	CounselList  []Counsel `json:"counselList"`
}

type Counsel struct {
	Name    string `json:"name"`
	Firm    string `json:"firm"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
}

type SectionConfig struct {
	Letter string
	Title  string
}

var sectionsOrder = []SectionConfig{
	{"a", "(a) Identity of Parties and Counsel"},
	{"d", "(d) Statement of the Case"},
	{"e", "(e) Statement Regarding Oral Argument"},
	{"f", "(f) Issues Presented"},
	{"g", "(g) Statement of Facts"},
	{"h", "(h) Summary of the Argument"},
	{"i", "(i) Argument"},
	{"j", "(j) Prayer"},
	{"k", "(k) Appendix"},
}

// toRoman converts an integer to lowercase Roman numerals (for prelim pages)
func toRoman(num int) string {
	if num <= 0 {
		return ""
	}
	val := []int{10, 9, 5, 4, 1}
	syb := []string{"x", "ix", "v", "iv", "i"}
	roman := ""
	for i := 0; i < len(val); i++ {
		for num >= val[i] {
			roman += syb[i]
			num -= val[i]
		}
	}
	return roman
}

// ParseMarkdown splits the markdown into metadata and section contents
func ParseMarkdown(md string) (PartiesMetadata, map[string]string) {
	meta := PartiesMetadata{}
	sections := map[string]string{}

	lines := strings.Split(md, "\n")
	inFrontmatter := false
	frontmatterContent := ""
	currentSection := ""
	var sectionLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "---json" || (trimmed == "---" && !inFrontmatter && i == 0) {
			inFrontmatter = true
			continue
		}
		if trimmed == "---" && inFrontmatter {
			inFrontmatter = false
			json.Unmarshal([]byte(frontmatterContent), &meta)
			continue
		}

		if inFrontmatter {
			frontmatterContent += line + "\n"
			continue
		}

		if strings.HasPrefix(trimmed, "# (") && len(trimmed) > 4 && trimmed[4] == ')' {
			if currentSection != "" {
				sections[currentSection] = strings.TrimSpace(strings.Join(sectionLines, "\n"))
			}
			currentSection = string(trimmed[3])
			sectionLines = []string{}
			continue
		}

		if currentSection != "" {
			sectionLines = append(sectionLines, line)
		}
	}
	if currentSection != "" {
		sections[currentSection] = strings.TrimSpace(strings.Join(sectionLines, "\n"))
	}

	return meta, sections
}

type TextRun struct {
	Text       string
	IsCitation bool
}

// splitParagraphByCitations parses text for [[case citations]] and returns formatted runs
func splitParagraphByCitations(paragraph string) []TextRun {
	var runs []TextRun
	re := regexp.MustCompile(`\[\[(.*?)\]\]`)
	matches := re.FindAllStringSubmatchIndex(paragraph, -1)

	if len(matches) == 0 {
		return []TextRun{{Text: paragraph, IsCitation: false}}
	}

	lastIdx := 0
	for _, match := range matches {
		startIdx := match[0]
		endIdx := match[1]
		contentStart := match[2]
		contentEnd := match[3]

		// Add leading text
		if startIdx > lastIdx {
			runs = append(runs, TextRun{Text: paragraph[lastIdx:startIdx], IsCitation: false})
		}

		// Add citation
		runs = append(runs, TextRun{Text: paragraph[contentStart:contentEnd], IsCitation: true})
		lastIdx = endIdx
	}

	if lastIdx < len(paragraph) {
		runs = append(runs, TextRun{Text: paragraph[lastIdx:], IsCitation: false})
	}

	return runs
}

// GeneratePDFHandler handles GET /api/drafts/{id}/pdf
func GeneratePDFHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := GetUserIdFromRequest(r)
	if err != nil {
		utilities.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 {
			idStr = parts[3]
		}
	}
	draftId, err := strconv.Atoi(idStr)
	if err != nil {
		utilities.SendError(w, http.StatusBadRequest, "Invalid draft ID")
		return
	}

	var d Draft
	err = db.DB.QueryRow("SELECT Id, Title, Markdown, UserId FROM Draft WHERE Id = ? AND UserId = ?", draftId, userId).Scan(&d.Id, &d.Title, &d.Markdown, &d.UserId)
	if err == sql.ErrNoRows {
		utilities.SendError(w, http.StatusNotFound, "Draft not found")
		return
	} else if err != nil {
		utilities.SendError(w, http.StatusInternalServerError, "Error loading draft")
		return
	}

	meta, sections := ParseMarkdown(d.Markdown)

	// Extract all unique citations
	citationRegex := regexp.MustCompile(`\[\[(.*?)\]\]`)
	allCitationsMap := map[string]bool{}
	for _, secText := range sections {
		matches := citationRegex.FindAllStringSubmatch(secText, -1)
		for _, m := range matches {
			if len(m) > 1 {
				citation := strings.TrimSpace(m[1])
				if citation != "" {
					allCitationsMap[citation] = true
				}
			}
		}
	}

	var sortedCitations []string
	for cit := range allCitationsMap {
		sortedCitations = append(sortedCitations, cit)
	}
	sort.Strings(sortedCitations)

	// Calculate Index of Authorities Pages
	ioaPages := int(math.Ceil(float64(len(sortedCitations)) / 25.0))
	if ioaPages == 0 {
		ioaPages = 1
	}

	tocPage := 2
	ioaStartPage := 3
	contentStartPage := 2 + 1 + ioaPages // Cover (1) + TOC (1) + IOA (ioaPages)

	// Initialize PDF
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(25.4, 25.4, 25.4)
	pdf.SetAutoPageBreak(true, 25.4)

	// Footer page number function
	pdf.SetFooterFunc(func() {
		if pdf.PageNo() > 1 {
			pdf.SetY(-15)
			pdf.SetFont("Times", "", 10)
			pageNo := pdf.PageNo()
			if pageNo < contentStartPage {
				// Roman numeral
				roman := toRoman(pageNo - 1)
				pdf.CellFormat(0, 10, roman, "", 0, "C", false, 0, "")
			} else {
				// Arabic page numbering starts at 1 for the content
				pageNumStr := strconv.Itoa(pageNo - contentStartPage + 1)
				pdf.CellFormat(0, 10, pageNumStr, "", 0, "C", false, 0, "")
			}
		}
	})

	// Page 1: Cover Page
	pdf.AddPage()
	pdf.SetFont("Times", "B", 14)
	pdf.Ln(10)
	pdf.CellFormat(0, 8, strings.ToUpper(meta.Jurisdiction), "", 1, "C", false, 0, "")
	pdf.Ln(10)

	if meta.CaseNumber != "" {
		pdf.SetFont("Times", "B", 12)
		pdf.CellFormat(0, 6, "No. "+meta.CaseNumber, "", 1, "C", false, 0, "")
		pdf.Ln(10)
	}

	// Appellant/Appellee Box
	pdf.SetFont("Times", "", 12)
	boxY := pdf.GetY()
	
	// Left column: Appellant & Appellee names
	pdf.SetY(boxY)
	pdf.SetX(25.4)
	pdf.MultiCell(95, 6, meta.Appellant+"\n\n      Appellant,\n\nv.\n\n"+meta.Appellee+"\n\n      Appellee.", "", "L", false)
	
	// Right column: Vertical line block
	endBoxY := pdf.GetY()
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.4)
	pdf.Line(130, boxY - 2, 130, endBoxY + 2)

	pdf.SetY(endBoxY + 15)
	
	// Title
	pdf.SetFont("Times", "B", 16)
	pdf.CellFormat(0, 8, "APPELLANT'S OPENING BRIEF", "", 1, "C", false, 0, "")
	pdf.Ln(25)

	// Counsel list at the bottom
	pdf.SetY(190)
	pdf.SetFont("Times", "B", 12)
	pdf.CellFormat(0, 6, "Respectfully Submitted,", "", 1, "L", false, 0, "")
	pdf.SetFont("Times", "", 11)
	for _, c := range meta.CounselList {
		counselStr := fmt.Sprintf("%s\n%s\n%s\nPhone: %s\nEmail: %s\n", c.Name, c.Firm, c.Address, c.Phone, c.Email)
		pdf.MultiCell(0, 5, counselStr, "", "L", false)
		pdf.Ln(4)
	}

	// Add blank placeholder page for TOC (Page 2)
	pdf.AddPage()

	// Add blank placeholder pages for Index of Authorities (Page 3 to contentStartPage-1)
	for i := 0; i < ioaPages; i++ {
		pdf.AddPage()
	}

	// Render Content Sections starting at contentStartPage
	sectionPages := map[string]int{}
	citationPages := map[string][]int{}

	for _, sec := range sectionsOrder {
		secText := sections[sec.Letter]
		if secText == "" {
			secText = "None."
		}

		// Ensure we start each section on a new page
		pdf.AddPage()
		sectionPages[sec.Letter] = pdf.PageNo()

		// Title
		pdf.SetFont("Times", "B", 12)
		pdf.CellFormat(0, 8, strings.ToUpper(sec.Title), "", 1, "C", false, 0, "")
		pdf.Ln(6)

		// Parse paragraphs
		paragraphs := strings.Split(secText, "\n\n")
		for _, para := range paragraphs {
			para = strings.TrimSpace(para)
			if para == "" {
				continue
			}

			runs := splitParagraphByCitations(para)
			for _, run := range runs {
				if run.IsCitation {
					citText := strings.TrimSpace(run.Text)
					
					// Record citation page
					currPage := pdf.PageNo()
					pagesList := citationPages[citText]
					found := false
					for _, p := range pagesList {
						if p == currPage {
							found = true
							break
						}
					}
					if !found {
						citationPages[citText] = append(pagesList, currPage)
					}

					// Split case name for italicization
					parts := strings.SplitN(citText, ",", 2)
					caseName := parts[0]
					pdf.SetFont("Times", "I", 12)
					pdf.Write(6, caseName)

					if len(parts) > 1 {
						pdf.SetFont("Times", "", 12)
						pdf.Write(6, ","+parts[1])
					}
				} else {
					pdf.SetFont("Times", "", 12)
					pdf.Write(6, run.Text)
				}
			}
			pdf.Ln(10) // paragraph spacing
		}
	}

	// --- Pass 2: Write back Table of Contents ---
	pdf.SetPage(tocPage)
	pdf.SetY(25.4)
	pdf.SetFont("Times", "B", 14)
	pdf.CellFormat(0, 10, "TABLE OF CONTENTS", "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Print TOC Rows
	printTOCRow(pdf, "(a) Identity of Parties and Counsel", formatPageNum(sectionPages["a"], contentStartPage))
	printTOCRow(pdf, "(b) Table of Contents", toRoman(tocPage-1))
	printTOCRow(pdf, "(c) Index of Authorities", toRoman(ioaStartPage-1))
	printTOCRow(pdf, "(d) Statement of the Case", formatPageNum(sectionPages["d"], contentStartPage))
	printTOCRow(pdf, "(e) Any Statement Regarding Oral Argument", formatPageNum(sectionPages["e"], contentStartPage))
	printTOCRow(pdf, "(f) Issues Presented", formatPageNum(sectionPages["f"], contentStartPage))
	printTOCRow(pdf, "(g) Statement of Facts", formatPageNum(sectionPages["g"], contentStartPage))
	printTOCRow(pdf, "(h) Summary of the Argument", formatPageNum(sectionPages["h"], contentStartPage))
	printTOCRow(pdf, "(i) Argument", formatPageNum(sectionPages["i"], contentStartPage))
	printTOCRow(pdf, "(j) Prayer", formatPageNum(sectionPages["j"], contentStartPage))
	printTOCRow(pdf, "(k) Appendix in Civil Cases", formatPageNum(sectionPages["k"], contentStartPage))

	// --- Pass 2: Write back Index of Authorities ---
	currIoaPage := ioaStartPage
	pdf.SetPage(currIoaPage)
	pdf.SetY(25.4)
	pdf.SetFont("Times", "B", 14)
	pdf.CellFormat(0, 10, "INDEX OF AUTHORITIES", "", 1, "C", false, 0, "")
	pdf.Ln(6)

	citationsPerPage := 25
	for idx, cit := range sortedCitations {
		if idx > 0 && idx%citationsPerPage == 0 {
			currIoaPage++
			pdf.SetPage(currIoaPage)
			pdf.SetY(25.4)
			pdf.SetFont("Times", "B", 14)
			pdf.CellFormat(0, 10, "INDEX OF AUTHORITIES (CONTINUED)", "", 1, "C", false, 0, "")
			pdf.Ln(6)
		}

		pagesList := citationPages[cit]
		var formattedPages []string
		for _, p := range pagesList {
			formattedPages = append(formattedPages, formatPageNum(p, contentStartPage))
		}
		pagesStr := strings.Join(formattedPages, ", ")
		if pagesStr == "" {
			pagesStr = "passim"
		}

		printIOARow(pdf, cit, pagesStr)
	}

	// Output PDF to stream
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"brief_%d.pdf\"", draftId))
	
	err = pdf.Output(w)
	if err != nil {
		fmt.Println("Error writing PDF output:", err)
	}
}

func formatPageNum(pageNo int, contentStartPage int) string {
	if pageNo == 0 {
		return "n/a"
	}
	if pageNo < contentStartPage {
		return toRoman(pageNo - 1)
	}
	return strconv.Itoa(pageNo - contentStartPage + 1)
}

func printTOCRow(pdf *gofpdf.Fpdf, title string, pageStr string) {
	pdf.SetFont("Times", "", 12)
	wTitle := pdf.GetStringWidth(title)
	wPage := pdf.GetStringWidth(pageStr)

	availWidth := 165.1
	dotWidth := pdf.GetStringWidth(".")

	numDots := int((availWidth - wTitle - wPage - 6.0) / dotWidth)
	dots := ""
	for i := 0; i < numDots; i++ {
		dots += "."
	}

	pdf.CellFormat(availWidth-wPage, 6, title+" "+dots, "", 0, "L", false, 0, "")
	pdf.CellFormat(wPage, 6, pageStr, "", 1, "R", false, 0, "")
}

func printIOARow(pdf *gofpdf.Fpdf, citation string, pagesStr string) {
	parts := strings.SplitN(citation, ",", 2)
	caseName := parts[0]
	remainder := ""
	if len(parts) > 1 {
		remainder = "," + parts[1]
	}

	pdf.SetFont("Times", "I", 12)
	wCase := pdf.GetStringWidth(caseName)
	pdf.CellFormat(wCase, 6, caseName, "", 0, "L", false, 0, "")

	pdf.SetFont("Times", "", 12)
	wRemainder := pdf.GetStringWidth(remainder)

	availWidth := 165.1
	wPage := pdf.GetStringWidth(pagesStr)

	remainingSpace := availWidth - wCase - wRemainder - wPage - 6.0
	dotWidth := pdf.GetStringWidth(".")
	numDots := int(remainingSpace / dotWidth)
	dots := ""
	for i := 0; i < numDots; i++ {
		dots += "."
	}

	pdf.CellFormat(wRemainder+remainingSpace, 6, remainder+" "+dots, "", 0, "L", false, 0, "")
	pdf.CellFormat(wPage, 6, pagesStr, "", 1, "R", false, 0, "")
}
