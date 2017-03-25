package main

import (
  "bufio"

  ui "github.com/gizak/termui"
  "os"
  "log"
)

func updateContentNmap(par *ui.Par, parser Parser, file *os.File) {
  reader := bufio.NewReader(file)
  currentLoc := 0
  contentLoc := new(ContentLoc)

  if len(parser.Loc) == 0 {
    contentLoc = &ContentLoc{0, -1}
  } else {
    contentLoc = parser.Loc[parser.IPList[0]]
  }

  log.Printf("Display: %v", parser.IPList[0])

  for {
    line, _, err := reader.ReadLine()
    if err != nil {
      break
    }
    lineStr := string(line)
    currentLoc += len(lineStr)
    if currentLoc >= contentLoc.Start && (contentLoc.End == -1 || currentLoc <= contentLoc.End) {
      par.Text += lineStr + "\n"
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
