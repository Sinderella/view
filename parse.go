package main

import (
	"bufio"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
  "strings"
)

type Parser struct {
	FilePath  string
	IPList    []string
	Loc       map[string]*ContentLoc
	Selecting int
}

type ContentLoc struct {
	Start int
	End   int
}

func ParseFile(filePath string) *Parser {
	switch fileExt := path.Ext(filePath); fileExt {
	case ".nmap":
		parser := parseNmap(filePath)
		return parser
	default:
		parser := parseRaw(filePath)
		return parser
	}
	return nil
}

func parseNmap(filePath string) *Parser {
	parser := new(Parser)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	re := regexp.MustCompile("^Nmap scan report for (.*)")
	// reip := regexp.MustCompile("([0-9]{1,3}\\.){3}[0-9]{1,3}(\\/([0-9]|[1-2][0-9]|3[0-2]))?")
	reader := bufio.NewReader(file)
	parser.FilePath = filePath
	parser.Loc = make(map[string]*ContentLoc)
	currentLoc := 0
	extractedIP := []string{""}
	lastIP := ""
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lineStr := string(line)
		currentLoc += len(lineStr)
		extractedContent := re.FindAllString(lineStr, -1)

		if len(extractedContent) == 0 {
			continue
		}
    
    log.Printf("%v", extractedContent[0])

		for _, content := range extractedContent {
			ip := strings.TrimPrefix(content, "Nmap scan report for ")
			extractedIP[0] = ip
		}
		parser.addIP(extractedIP)

		if len(extractedIP) == 0 {
			continue
		}

		if lastIP == "" {
			parser.Loc[extractedIP[0]] = &ContentLoc{currentLoc, -1}
			lastIP = extractedIP[0]
		} else if lastIP != extractedIP[0] {
			parser.Loc[lastIP].End = currentLoc
			parser.Loc[extractedIP[0]] = &ContentLoc{currentLoc, -1}
			lastIP = extractedIP[0]
		}
	}

	sort.Strings(parser.IPList)

	return parser
}

func parseRaw(filePath string) *Parser {
	parser := new(Parser)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	parser.FilePath = filePath

	re := regexp.MustCompile("([0-9]{1,3}\\.){3}[0-9]{1,3}(\\/([0-9]|[1-2][0-9]|3[0-2]))?")
	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lineStr := string(line)
		extractedIP := re.FindAllString(lineStr, -1)
		parser.addIP(extractedIP)
	}
	return parser
}

func (parser *Parser) addIP(ipList []string) {
	duplicated := false
	for _, newIP := range ipList {
		duplicated = false
		for _, existingIP := range parser.IPList {
			if newIP == existingIP {
				duplicated = true
				break
			}
		}
		if !duplicated {
			parser.IPList = append(parser.IPList, newIP)
		}
	}
}
