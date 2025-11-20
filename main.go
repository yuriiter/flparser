package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

const (
	defaultSkills          = "7,9,13,31,68,137,305,323,335,500,598,613,673,759,913,1031,1087,1088,1936,2376"
	defaultClientCountries = "ca,au,no,de,se,ch,gb,us,at,fr,jp,ae,es,lu,ie,nl,be,fi,it,sg,kr,hk,is,nz"
)

type Project struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Budget      string `json:"budget"`
	AverageBid  string `json:"average_bid"`
	BidsCount   string `json:"bids_count"`
	TimeLeft    string `json:"time_left"`
	Description string `json:"description"`
}

type OutputData struct {
	Parameters map[string]string `json:"parameters"`
	Projects   []Project         `json:"projects"`
}

var (
	pTypes          string
	clientCountries []string
	fixedPriceMin   int
	fixedPriceMax   int
	hourlyRateMin   int
	hourlyRateMax   int
	skills          string
	sortOption      string
	queryText       string
	pageNumber      int
	outputFile      string
	outputExt       string
)

var rootCmd = &cobra.Command{
	Use:   "flparser",
	Short: "Scrape projects from Freelancer.com",
	Long: `A CLI tool to parse projects from Freelancer.com based on specific criteria 
and export them to Markdown, CSV, or JSON.`,
	Run: func(cmd *cobra.Command, args []string) {
		runScraper()
	},
}

func main() {
	rootCmd.Flags().StringVar(&pTypes, "types", "hourly,fixed", "Project types: 'hourly,fixed', 'hourly', or 'fixed'")
	rootCmd.Flags().StringSliceVar(&clientCountries, "clientCountries", strings.Split(defaultClientCountries, ","), "Comma separated client country codes")

	rootCmd.Flags().IntVar(&fixedPriceMin, "fixedMin", 0, "Minimum fixed price")
	rootCmd.Flags().IntVar(&fixedPriceMax, "fixedMax", 0, "Maximum fixed price")
	rootCmd.Flags().IntVar(&hourlyRateMin, "hourlyMin", 0, "Minimum hourly rate")
	rootCmd.Flags().IntVar(&hourlyRateMax, "hourlyMax", 0, "Maximum hourly rate")

	rootCmd.Flags().StringVar(&skills, "skills", defaultSkills, "Skill IDs comma separated, or 'all'")
	rootCmd.Flags().StringVar(&sortOption, "sort", "latest", "Sort: oldest, lowestPrice, highestPrice, fewestBids, mostBids")

	rootCmd.Flags().StringVar(&queryText, "q", "", "Search query text")
	rootCmd.Flags().IntVar(&pageNumber, "page", 1, "Page number")

	rootCmd.Flags().StringVarP(&outputFile, "output", "O", "", "Output filename (e.g. results.json)")
	rootCmd.Flags().StringVarP(&outputExt, "extension", "X", "", "Output extension if -O is not set (md, csv, json)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runScraper() {
	// 1. Build URL
	targetURL, paramsMap := buildURL()
	fmt.Print("Fetching Freelancer.com...\n")

	projects, err := scrapeFreelancer(targetURL)
	if err != nil {
		log.Fatalf("Error scraping: %v", err)
	}
	fmt.Printf("Found %d projects.\n", len(projects))

	handleOutput(projects, paramsMap)
}

func buildURL() (string, map[string]string) {
	baseURL := "https://www.freelancer.com/search/projects"
	u, _ := url.Parse(baseURL)
	q := u.Query()

	paramsRecord := make(map[string]string)

	// Types
	if pTypes != "" {
		q.Set("types", pTypes)
		paramsRecord["types"] = pTypes
	}

	if len(clientCountries) > 0 {
		val := strings.Join(clientCountries, ",")
		q.Set("clientCountries", val)
		paramsRecord["clientCountries"] = val
	}

	if fixedPriceMin > 0 {
		q.Set("projectFixedPriceMin", strconv.Itoa(fixedPriceMin))
		paramsRecord["projectFixedPriceMin"] = strconv.Itoa(fixedPriceMin)
	}
	if fixedPriceMax > 0 {
		q.Set("projectFixedPriceMax", strconv.Itoa(fixedPriceMax))
		paramsRecord["projectFixedPriceMax"] = strconv.Itoa(fixedPriceMax)
	}
	if hourlyRateMin > 0 {
		q.Set("projectHourlyRateMin", strconv.Itoa(hourlyRateMin))
		paramsRecord["projectHourlyRateMin"] = strconv.Itoa(hourlyRateMin)
	}
	if hourlyRateMax > 0 {
		q.Set("projectHourlyRateMax", strconv.Itoa(hourlyRateMax))
		paramsRecord["projectHourlyRateMax"] = strconv.Itoa(hourlyRateMax)
	}

	if skills != "all" {
		q.Set("projectSkills", skills)
		paramsRecord["projectSkills"] = skills
	} else {
		paramsRecord["projectSkills"] = "all"
	}

	if sortOption != "" && sortOption != "latest" {
		q.Set("projectSort", sortOption)
		paramsRecord["projectSort"] = sortOption
	}

	if queryText != "" {
		q.Set("q", queryText)
		paramsRecord["q"] = queryText
	}

	if pageNumber > 1 {
		q.Set("page", strconv.Itoa(pageNumber))
		paramsRecord["page"] = strconv.Itoa(pageNumber)
	}

	u.RawQuery = q.Encode()
	return u.String(), paramsRecord
}

func scrapeFreelancer(urlStr string) ([]Project, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var projects []Project

	cleanText := func(s string) string {
		s = strings.ReplaceAll(s, "\n", " ")
		s = strings.ReplaceAll(s, "\r", " ")
		for strings.Contains(s, "  ") {
			s = strings.ReplaceAll(s, "  ", " ")
		}
		return strings.TrimSpace(s)
	}

	doc.Find(".JobSearchCard-item").Each(func(i int, s *goquery.Selection) {
		titleNode := s.Find(".JobSearchCard-primary-heading a")
		title := cleanText(titleNode.Text())

		linkHref, exists := s.Find("a.JobSearchCard-ctas-btn").Attr("href")
		if !exists {
			linkHref, _ = titleNode.Attr("href")
		}
		if strings.HasPrefix(linkHref, "/") {
			linkHref = "https://www.freelancer.com" + linkHref
		}

		desc := cleanText(s.Find(".JobSearchCard-primary-description").Text())

		timeLeft := cleanText(s.Find(".JobSearchCard-primary-heading-days").Text())

		priceFull := s.Find(".JobSearchCard-secondary-price").Text()

		budget := cleanText(priceFull)
		budget = strings.ReplaceAll(budget, "Avg Bid", "")
		budget = cleanText(budget)

		bids := cleanText(s.Find(".JobSearchCard-secondary-entry").Text())

		avgBid := budget

		p := Project{
			Title:       title,
			Link:        linkHref,
			Description: desc,
			TimeLeft:    timeLeft,
			Budget:      budget,
			AverageBid:  avgBid,
			BidsCount:   bids,
		}
		projects = append(projects, p)
	})

	return projects, nil
}
func handleOutput(projects []Project, params map[string]string) {
	timestamp := time.Now().Format("15-04-05_02-01-2006")
	baseName := fmt.Sprintf("freelancer.com_%s", timestamp)

	var targetFile string
	var formats []string

	if outputFile != "" {
		targetFile = outputFile
		ext := strings.ToLower(filepath.Ext(outputFile))
		if ext == "" {
			if outputExt != "" {
				formats = []string{outputExt}
				targetFile = outputFile + "." + outputExt
			} else {
				formats = []string{"csv"}
				targetFile = outputFile + ".csv"
			}
		} else {
			formats = []string{ext[1:]}
		}
	} else {
		if outputExt != "" {
			formats = []string{outputExt}
			targetFile = fmt.Sprintf("%s.%s", baseName, outputExt)
		} else {
			formats = []string{"md", "csv"}
			targetFile = baseName
		}
	}

	for _, fmtType := range formats {
		fname := targetFile
		if outputFile == "" && len(formats) > 1 {
			fname = fmt.Sprintf("%s.%s", baseName, fmtType)
		}

		switch strings.ToLower(fmtType) {
		case "json":
			writeJSON(fname, projects, params)
		case "csv":
			writeCSV(fname, projects, params)
		case "md":
			writeMarkdown(fname, projects, params)
		default:
			fmt.Printf("Unknown format: %s\n", fmtType)
		}
	}
}

func writeJSON(filename string, projects []Project, params map[string]string) {
	data := OutputData{
		Parameters: params,
		Projects:   projects,
	}
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Println("Error marshalling JSON:", err)
		return
	}
	if err := os.WriteFile(filename, file, 0644); err != nil {
		log.Println("Error writing JSON file:", err)
		return
	}
	fmt.Println("Generated:", filename)
}

func writeCSV(filename string, projects []Project, params map[string]string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Println("Error creating CSV file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"# Parameters Used:"})
	for k, v := range params {
		writer.Write([]string{"# " + k + ": " + v})
	}

	header := []string{"Title", "Time Left", "Bids", "Price/AvgBid", "Link", "Description"}
	writer.Write(header)

	for _, p := range projects {
		row := []string{
			p.Title,
			p.TimeLeft,
			p.BidsCount,
			p.Budget,
			p.Link,
			strings.ReplaceAll(p.Description, "\n", " "),
		}
		writer.Write(row)
	}
	fmt.Println("Generated:", filename)
}

func writeMarkdown(filename string, projects []Project, params map[string]string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Println("Error creating Markdown file:", err)
		return
	}
	defer file.Close()

	var sb strings.Builder

	sb.WriteString("# Freelancer.com Projects\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format(time.RFC1123)))

	sb.WriteString("### Search Parameters\n")
	sb.WriteString("| Parameter | Value |\n| --- | --- |\n")
	for k, v := range params {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", k, v))
	}
	sb.WriteString("\n---\n\n")

	for _, p := range projects {
		sb.WriteString(fmt.Sprintf("## [%s](%s)\n", strings.TrimSpace(p.Title), p.Link))
		sb.WriteString(fmt.Sprintf("- **Budget/Price:** %s\n", p.Budget))
		sb.WriteString(fmt.Sprintf("- **Bids:** %s\n", p.BidsCount))
		sb.WriteString(fmt.Sprintf("- **Time:** %s\n", p.TimeLeft))
		sb.WriteString(fmt.Sprintf("\n> %s\n\n", p.Description))
		sb.WriteString("---\n")
	}

	file.WriteString(sb.String())
	fmt.Println("Generated:", filename)
}
