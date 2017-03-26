package main

import (
  "fmt"
  "io/ioutil"
  "log"
  "os"
  "path"

  ui "github.com/gizak/termui"
  "github.com/rjeczalik/notify"
)

var (
  selectingColumn = 0

  currentSelectFile = 0
  firstDisplayFile  = 0
  noOfFiles         = 0

  currentSelectIP = 0
  firstDisplayIP  = 0
  noOfIPs         = 0

  currentSelectLine = 0
  firstDisplayLine  = 0
  noOfLines         = 0
)

func CreateUI(done chan<- struct{}, notifyCh <-chan notify.EventInfo, watchingPath string) {
  err := ui.Init()
  if err != nil {
    panic(err)
  }
  defer ui.Close()

  changeFileCh := make(chan string)

  iplst, displayLocCh := createIPList(changeFileCh)
  contentArea := createContentArea(displayLocCh)
  ls := createFileList(notifyCh, changeFileCh, watchingPath)

  ui.Body.AddRows(
    ui.NewRow(
      ui.NewCol(2, 0, ls),
      ui.NewCol(1, 0, iplst),
      ui.NewCol(9, 0, contentArea),
    ),
  )

  ui.Handle("/sys/kbd/<left>", func(e ui.Event) {
    selectingColumn--
    if selectingColumn < 0 {
      selectingColumn = 0
    }
    selectedFilePath := updateFileList(watchingPath, changeFileCh, ls)
    log.Println("Selecting: " + selectedFilePath)
  })

  ui.Handle("/sys/kbd/<right>", func(e ui.Event) {
    selectingColumn++
    if selectingColumn > 2 {
      selectingColumn = 2
    }
    selectedFilePath := updateFileList(watchingPath, changeFileCh, ls)
    log.Println("Selecting: " + selectedFilePath)
  })

  ui.Handle("/sys/kbd/<up>", func(e ui.Event) {
    switch selectingColumn {
    case 0:
      currentSelectFile--
      if currentSelectFile < 0 {
        currentSelectFile = 0
      }
      currentSelectIP = 0
      //selectedFilePath := updateFileList(watchingPath, changeFileCh, ls)
      //log.Println("Selecting: " + selectedFilePath)
    case 1:
      currentSelectIP--
      if currentSelectIP < 0 {
        currentSelectIP = 0
      }
      currentSelectLine = 0
      //updateIPList(iplst, displayLocCh, getCurrentFilePath(ls.Items, watchingPath))
    case 2:
      currentSelectLine--
      log.Printf("curLine: %v, noOfLines: %v", currentSelectLine, noOfLines)
      if currentSelectLine < 0 {
        currentSelectLine = 0
      }
    }
    selectedFilePath := updateFileList(watchingPath, changeFileCh, ls)
    log.Println("Selecting: " + selectedFilePath)
    ui.Render(contentArea)
  })

  ui.Handle("/sys/kbd/<down>", func(e ui.Event) {
    switch selectingColumn {
    case 0:
      currentSelectFile++
      if currentSelectFile >= noOfFiles {
        currentSelectFile = noOfFiles - 1
      }
      currentSelectIP = 0
      //selectedFilePath := updateFileList(watchingPath, changeFileCh, ls)
      //log.Println("Selecting: " + selectedFilePath)
    case 1:
      currentSelectIP++
      log.Printf("curIP: %v, noOfIPs: %v", currentSelectIP, noOfIPs)
      if currentSelectIP >= noOfIPs && noOfIPs != 0 {
        currentSelectIP = noOfIPs - 1
      }
      currentSelectLine = 0
      //updateIPList(iplst, displayLocCh, getCurrentFilePath(ls.Items, watchingPath))
    case 2:
      currentSelectLine++
      log.Printf("curLine: %v, noOfLines: %v", currentSelectLine, noOfLines)
      if currentSelectLine >= noOfLines && noOfLines != 0 {
        currentSelectLine = noOfLines - 1
      }
    }
    selectedFilePath := updateFileList(watchingPath, changeFileCh, ls)
    log.Println("Selecting: " + selectedFilePath)
    ui.Body.Align()
    ui.Render(ui.Body)
  })

  ui.Body.Align()
  ui.Render(ui.Body)

  ui.Handle("/sys/kbd/q", func(ui.Event) {
    done <- struct{}{}
    ui.StopLoop()
  })
  ui.Loop()
}

func createContentArea(displayLocCh <-chan Parser) (par *ui.Par) {
  par = ui.NewPar("")
  par.BorderLabel = "Content"
  par.Height = ui.TermHeight()
  par.Border = true
  par.Y = 0

  go monitorUpdateContentArea(displayLocCh, par)
  return par
}

func updateContentArea(par *ui.Par, parser Parser) {
  par.Text = ""
  file, err := os.Open(parser.FilePath)
  if err != nil {
    log.Fatal(err)
  }
  defer file.Close()

  switch ext := path.Ext(parser.FilePath); ext {
  case ".nmap":
    updateContentNmap(par, parser, file)
  default:
    updateContentRaw(par, file)
  }

  ui.Body.Align()
  ui.Render(ui.Body)
}

func monitorUpdateContentArea(displayLocCh <-chan Parser, par *ui.Par) {
  for {
    select {
    case parser := <-displayLocCh:
      log.Printf("monitorUpdateContentArea: loc(%+v)", parser)
      updateContentArea(par, parser)
    }
  }
}

func createIPList(changeFileCh <-chan string) (par *ui.Par, displayLocCh chan Parser) {
  displayLocCh = make(chan Parser)

  par = ui.NewPar("")
  par.BorderLabel = "IP List"
  par.Height = ui.TermHeight()
  par.Width = 25
  par.Border = true

  go monitorUpdateIPList(changeFileCh, displayLocCh, par)
  return par, displayLocCh
}

func updateIPList(par *ui.Par, displayLocCh chan<- Parser, filePath string) {
  if filePath == "" {
    return
  }
  parser := ParseFile(filePath)
  noOfIPs = len(parser.IPList)
  displayLocCh <- *parser

  par.Text = ""

  windowList := ui.TermHeight() - 2
  for idx, ip := range parser.IPList {
    if idx < firstDisplayIP {
      continue
    }
    if firstDisplayIP+windowList == currentSelectIP {
      firstDisplayIP = currentSelectIP - windowList + 1
      continue
    } else if currentSelectIP <= firstDisplayIP {
      firstDisplayIP--
      if firstDisplayIP < 0 {
        firstDisplayIP = 0
      }
    }

    if idx == currentSelectIP {
      if selectingColumn == 1 {
        par.Text += fmt.Sprintf("[%s](bg-green)\n", ip)
      } else {
        par.Text += fmt.Sprintf("[%s](bg-red)\n", ip)
      }
    } else {
      par.Text += ip + "\n"
    }
  }

  ui.Body.Align()
  ui.Render(ui.Body)
}

func monitorUpdateIPList(changeFileCh <-chan string, displayLocCh chan<- Parser, par *ui.Par) {
  for {
    select {
    case filePath := <-changeFileCh:
      log.Println("monitorUpdateIPList: " + filePath)
      updateIPList(par, displayLocCh, filePath)
    }
  }
}

func createFileList(notifyCh <-chan notify.EventInfo, changeFileCh chan<- string, watchingPath string) (ls *ui.List) {
  filesName := []string{}
  ls = ui.NewList()
  ls.Items = filesName
  ls.ItemFgColor = ui.ColorYellow
  ls.BorderLabel = "List"
  ls.Height = ui.TermHeight()
  ls.Width = 25
  ls.Y = 0

  updateFileList(watchingPath, changeFileCh, ls)

  ui.Body.Align()
  ui.Render(ui.Body)

  go monitorUpdateFileList(notifyCh, changeFileCh, watchingPath, ls)

  return ls
}

func filterFiles(fileList []os.FileInfo) (filteredList []os.FileInfo) {
  for _, file := range fileList {
    switch ext := path.Ext(file.Name()); ext {
    case ".xml":
      continue
    case ".gnmap":
      continue
    default:
      filteredList = append(filteredList, file)
    }
  }

  return filteredList
}

func updateFileList(watchingPath string, changeFileCh chan<- string, ls *ui.List) (selectedFilePath string) {
  filesName := []string{}

  filesInfo, err := ioutil.ReadDir(watchingPath)
  if err != nil {
    log.Panic("Cannot list files: " + err.Error())
  }
  filesInfo = filterFiles(filesInfo)
  noOfFiles = len(filesInfo)
  windowList := ui.TermHeight() - 2
  log.Printf("cur: %d", currentSelectFile)
  log.Printf("first: %d", firstDisplayFile)
  for idx, fileInfo := range filesInfo {
    if idx < firstDisplayFile {
      continue
    }
    if firstDisplayFile+windowList == currentSelectFile {
      firstDisplayFile = currentSelectFile - windowList + 1
      continue
    } else if currentSelectFile <= firstDisplayFile {
      firstDisplayFile--
      if firstDisplayFile < 0 {
        firstDisplayFile = 0
      }
    }
    fileName := fileInfo.Name()
    if fileInfo.IsDir() {
      fileName += "/"
    }
    if idx == currentSelectFile {
      selectedFilePath = path.Join(watchingPath, fileName)
      changeFileCh <- selectedFilePath
      if selectingColumn == 0 {
        fileName = fmt.Sprintf("[%s](bg-green)", fileName)
      } else {
        fileName = fmt.Sprintf("[%s](bg-red)", fileName)
      }
    }
    filesName = append(filesName, fileName)
  }

  ls.Items = filesName
  //ui.Body.Align()
  //ui.Render(ls)

  return selectedFilePath
}

func monitorUpdateFileList(notifyCh <-chan notify.EventInfo, changeFileCh chan<- string, watchingPath string, ls *ui.List) {
  for {
    select {
    case <-notifyCh:
      updateFileList(watchingPath, changeFileCh, ls)
    }
  }
}
