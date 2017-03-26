package main

import (
  "bufio"

  ui "github.com/gizak/termui"
  "os"
  "log"
  "strings"
)

func highlightLine(line string) (string) {
  if output, found := highlightKw(line, "(Domain: ", ")"); found {
    return output
  } else if output, found = highlightKw(line, "(workgroup: ", ")"); found {
    return output
  } else if output, found = highlightKw(line, "| ssl-cert: Subject:", ""); found {
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
    output = strings.Replace(line, kwStart, "["+kwStart, -1)
    if kwEnd == "" {
      output += "](bg-yellow)"
    } else {
      output = strings.Replace(output, kwEnd, kwEnd+"](bg-yellow)", -1)
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

  for {
    line, _, err := reader.ReadLine()
    if err != nil || lineCnt == windowHeight {
      break
    }
    lineStr := string(line)
    currentLoc += len(lineStr)
    if currentLoc >= contentLoc.Start && (contentLoc.End == -1 || currentLoc <= contentLoc.End) {
      par.Text += highlightLine(lineStr) + "\n"
      lineCnt++
    }
  }
}

func updateContentRaw(par *ui.Par, file *os.File) {
  reader := bufio.NewReader(file)
  currentLoc := 0

  for {
    line, _, err := reader.ReadLine()
    if err != nil {
      break
    }
    lineStr := string(line)
    currentLoc += len(lineStr)
    par.Text += lineStr + "\n"

  }
}
