package main

import (
	"bufio"

	"fmt"
	ui "github.com/gizak/termui"
	"log"
	"os"
	"strings"
)

func highlightLine(line string) string {
	if output, found := highlightKw(line, "(Domain: ", ")"); found {
		return output
	} else if output, found = highlightKw(line, "(workgroup: ", ")"); found {
		return output
	} else if output, found = highlightKw(line, "ssl-cert: Subject:", ""); found {
		return output
	} else if output, found = highlightKw(line, "Subject Alternative Name: DNS:", ""); found {
		return output
	} else if output, found = highlightKw(line, "Host: ", ";"); found {
		return output
	} else if output, found = highlightKw(line, "bind.version: ", ""); found {
		return output
	} else if output, found = highlightKw(line, "http-server-header: ", ""); found {
		return output
	} else if output, found = highlightKw(line, "Issuer: ", ""); found {
		return output
	} else if output, found = highlightKw(line, "Computer name: ", ""); found {
		return output
	} else if output, found = highlightKw(line, "NetBIOS computer name: ", ""); found {
		return output
	} else if output, found = highlightKw(line, "Domain name: ", ""); found {
		return output
	} else if output, found = highlightKw(line, "Forest name: ", ""); found {
		return output
	} else if output, found = highlightKw(line, "FQDN: ", ""); found {
		return output
	} else if output, found = highlightKw(line, "Workgroup:", ""); found {
		return output
	} else if output, found = highlightKw(line, "(RID:", ")"); found {
		return output
	} else if output, found = highlightKw(line, "Anonymous access: READ", ""); found {
		return output
	} else if output, found = highlightKw(line, "Potentially risky methods: ", ""); found {
		return output
	} else if output, found = highlightKw(line, "ms-sql-info:", ""); found {
		return output
	}
	//if strings.Contains(line, "(Domain: ") {
	//  tmp := strings.Replace(line, "(Domain: ", "[(Domain: ", -1)
	//  tmp = strings.Replace(tmp, ")", ")](bg-yellow)", -1)
	//  return tmp
	//}

	return line
}

func highlightKw(line, kwStart, kwEnd string) (output string, found bool) {
	if strings.Contains(line, kwStart) {
		output = strings.Replace(line, kwStart, "["+kwStart, 1)
		if kwEnd == "" {
			output += "](fg-black,bg-yellow)"
		} else {
			output = strings.Replace(output, kwEnd, kwEnd+"](fg-black,bg-yellow)", 1)
		}
		return output, true
	}
	return line, false
}

func updateContentNmap(par *ui.Par, parser Parser, file *os.File) {
	reader := bufio.NewReader(file)
	currentLoc := 0
	contentLoc := new(ContentLoc)

	if len(parser.Loc) == 0 {
		contentLoc = &ContentLoc{0, -1}
	} else {
		log.Printf("currentSelectIP: %v", currentSelectIP)
		contentLoc = parser.Loc[parser.IPList[currentSelectIP]]
	}

	windowHeight := ui.TermHeight() - 2
	lineCnt := 0
	selectLineCnt := currentSelectLine
	lineNoCnt := currentSelectLine

	for {
		line, _, err := reader.ReadLine()
		if err != nil || lineCnt == windowHeight {
			break
		}
		lineStr := string(line)
		currentLoc += len(lineStr)

		if currentLoc >= contentLoc.Start && (contentLoc.End == -1 || currentLoc < contentLoc.End) {
			if selectLineCnt > 0 {
				selectLineCnt--
				continue
			}
			lineNoCnt++
			lineNo := fmt.Sprintf("%3d|", lineNoCnt)
			par.Text += lineNo + highlightLine(lineStr) + "\n"
			lineCnt++
		}
	}
}

func updateContentRaw(par *ui.Par, file *os.File) {
	reader := bufio.NewReader(file)
	currentLoc := 0
	selectLineCnt := currentSelectLine
	lineNoCnt := currentSelectLine

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lineStr := string(line)
		currentLoc += len(lineStr)
		if selectLineCnt > 0 {
			selectLineCnt--
			continue
		}

		lineNoCnt++
		lineNo := fmt.Sprintf("%3d|", lineNoCnt)
		par.Text += lineNo + lineStr + "\n"

	}
}
