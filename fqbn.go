package fqbn

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	boardFileName string = "boards.txt"
)

type boardInfo struct {
	name string
	vids []string
	pids []string
}

var boards = map[string]*boardInfo{}

// NameFor : get the fully qualified name for a board with
// a given pid and vid
func NameFor(pid string, vid string) (string, error) {
	if len(boards) == 0 {
		return "", errors.New("There are no cached board details, call LoadBoardInfoFrom first")
	}
	for k, v := range boards {
		if contains(v.pids, pid) && contains(v.vids, vid) {
			return k, nil
		}
	}
	return "", errors.New("There is no name that matches that pid/vid combination")
}

// LoadBoardInfoFrom : Search for board definitions starting at a root, and load them
func LoadBoardInfoFrom(hardwareDirPath string) error {
	err := findBoardFile(hardwareDirPath)
	for _, v := range boards {
		fmt.Println(v)
	}
	return err
}

// oneach walk of the filepath, check if the filename is boards.text
// if it is, try to parse it.
func walk(path string, f os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
		return err
	}
	if !f.IsDir() && f.Name() == boardFileName {
		err := parseForBoardInfo(path)
		return err
	}
	return nil
}

func findBoardFile(rootPath string) error {
	err := filepath.Walk(rootPath, walk)
	if err != nil {
		return err
	}
	return nil
}

func contains(slice []string, element string) bool {
	sort.Strings(slice)
	i := sort.SearchStrings(slice, element)
	return i < len(slice) && slice[i] == element
}

func parseForBoardInfo(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	board := new(boardInfo)
	rootName := getRootName(path)

	//scan every line in the file to extract pid, vid, and board name values
	for scanner.Scan() {
		line := scanner.Text()

		boardName := strings.Split(line, ".")[0]
		fullyQualifiedBoardName := rootName + ":" + boardName

		if isVidDefinition(line) {
			fmt.Println("line " + line + " is vid def")
			vid := line[len(line)-6 : len(line)]
			fmt.Println("found vid: ", vid)
			if board.name == fullyQualifiedBoardName {
				board.vids = append(board.vids, vid)
			} else {
				boards[board.name] = board
				board = new(boardInfo)
				board.name = fullyQualifiedBoardName
				board.vids = append(board.vids, vid)
			}

		} else if isPidDefinition(line) {
			fmt.Println("line " + line + " is pid def")
			pid := line[len(line)-6 : len(line)]
			fmt.Println("found pid: ", pid)
			if board.name == fullyQualifiedBoardName {
				board.pids = append(board.pids, pid)
			} else {
				boards[board.name] = board
				board = new(boardInfo)
				board.name = fullyQualifiedBoardName
				board.pids = append(board.pids, pid)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func getRootName(path string) string {
	precedingDirs, _ := filepath.Split(path)
	pathElements := strings.Split(precedingDirs, string(os.PathSeparator))
	pathElements = pathElements[1 : len(pathElements)-1]
	var sb strings.Builder
	for _, v := range pathElements {
		sb.WriteString(":" + v)
	}
	rootName := sb.String()[1:sb.Len()] // take a string formatted like arduino:avr
	return rootName
}

func isVidDefinition(line string) bool {
	isVid, _ := regexp.MatchString("[a-zA-Z]+\\.(vid)\\.[0-9]=0x[[:xdigit:]]{4}", line)
	return isVid
}

func isPidDefinition(line string) bool {
	isPid, _ := regexp.MatchString("[a-zA-Z]+\\.(pid)\\.[0-9]=0x[[:xdigit:]]{4}", line)
	return isPid
}
