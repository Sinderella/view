package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	ui "github.com/gizak/termui"
	"github.com/rjeczalik/notify"
)

var currentSelectFile = 0
var firstDisplayFile = 0
var noOfFiles = 0

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

	ui.Body.Align()
	ui.Render(ui.Body)

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		done <- struct{}{}
		ui.StopLoop()
	})
	ui.Loop()
}

func createContentArea(changeFileCh <-chan string) (par *ui.Par) {
	par = ui.NewPar("")
	par.BorderLabel = "Content"
	par.Height = ui.TermHeight()
	par.Border = true
	par.Y = 0

	go monitorUpdateContentArea(changeFileCh, par)
	return par
}

func updateContentArea(par *ui.Par, filePath string) {
	if filePath == "" {
		return
	}
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	par.Text = ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		par.Text += scanner.Text() + "\n"
	}
	ui.Body.Align()
	ui.Render(ui.Body)
}

func monitorUpdateContentArea(changeFileCh <-chan string, par *ui.Par) {
	for {
		select {
		case filePath := <-changeFileCh:
			log.Println("monitorUpdateContentArea: " + filePath)
			updateContentArea(par, filePath)
		}
	}
}

func createIPList(changeFileCh <-chan string) (par *ui.Par, displayLocCh chan string) {
	displayLocCh = make(chan string)

	par = ui.NewPar("")
	par.BorderLabel = "IP List"
	par.Height = ui.TermHeight()
	par.Width = 25
	par.Border = true

	ui.Handle("/sys/kbd/j", func(e ui.Event) {
		log.Println("Selecting: ")
	})

	ui.Handle("/sys/kbd/k", func(e ui.Event) {
		log.Println("Selecting: ")
	})

	go monitorUpdateIPList(changeFileCh, displayLocCh, par)
	return par, displayLocCh
}

func updateIPList(par *ui.Par, filePath string) {
	if filePath == "" {
		return
	}
	ipList := ParseFile(filePath)

	par.Text = ""

	for _, ip := range ipList {
		par.Text += ip + "\n"
	}

	ui.Body.Align()
	ui.Render(ui.Body)
}

func monitorUpdateIPList(changeFileCh <-chan string, displayLocCh chan<- string, par *ui.Par) {
	for {
		select {
		case filePath := <-changeFileCh:
			log.Println("monitorUpdateIPList: " + filePath)
			updateIPList(par, filePath)
			displayLocCh <- filePath
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

	ui.Handle("/sys/kbd/<up>", func(e ui.Event) {
		currentSelectFile--
		if currentSelectFile < 0 {
			currentSelectFile = 0
		}
		selectedFilePath := updateFileList(watchingPath, changeFileCh, ls)
		log.Println("Selecting: " + selectedFilePath)
	})

	ui.Handle("/sys/kbd/<down>", func(e ui.Event) {
		currentSelectFile++
		if currentSelectFile >= noOfFiles {
			currentSelectFile = noOfFiles - 1
		}
		selectedFilePath := updateFileList(watchingPath, changeFileCh, ls)
		log.Println("Selecting: " + selectedFilePath)
	})

	ui.Body.Align()
	ui.Render(ui.Body)

	go monitorUpdateFileList(notifyCh, changeFileCh, watchingPath, ls)

	return ls
}

func updateFileList(watchingPath string, changeFileCh chan<- string, ls *ui.List) (selectedFilePath string) {
	filesName := []string{}

	filesInfo, err := ioutil.ReadDir(watchingPath)
	noOfFiles = len(filesInfo)
	windowList := ui.TermHeight() - 2
	if err != nil {
		log.Panic("Cannot list files: " + err.Error())
	}
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
			fileName = fmt.Sprintf("[%s](bg-green)", fileName)
		}
		filesName = append(filesName, fileName)
	}

	ls.Items = filesName
	ui.Body.Align()
	ui.Render(ls)

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
