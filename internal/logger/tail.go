package logger

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Time    time.Time
	Level   string
	UserID  int64
	Message string
	Raw     string
}

var (
	entryRegex = regexp.MustCompile(`^\[(INPUT|OUTPUT|DEBUG|ERROR)\]  \[(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\] (User (\d+): )?(.*)$`)
	dayFileReg = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})\.log$`)
)

func ParseDay(workspace string, date time.Time) ([]Entry, error) {
	dateStr := date.Format("2006-01-02")
	path := filepath.Join(workspace, ".logs", dateStr+".log")
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no log file for %s at %s", dateStr, path)
		}
		return nil, err
	}
	defer f.Close()
	return parseFile(f)
}

func TailLastN(workspace string, n int, onlyDate time.Time) ([]Entry, error) {
	if n <= 0 {
		return []Entry{}, nil
	}

	var entries []Entry

	if !onlyDate.IsZero() {
		parsed, err := ParseDay(workspace, onlyDate)
		if err != nil {
			return nil, err
		}
		entries = parsed
	} else {
		logsDir := filepath.Join(workspace, ".logs")
		files, err := os.ReadDir(logsDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return []Entry{}, nil
			}
			return nil, err
		}

		type datedFile struct {
			date time.Time
			path string
		}
		var matched []datedFile
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			m := dayFileReg.FindStringSubmatch(f.Name())
			if m == nil {
				continue
			}
			d, perr := time.ParseInLocation("2006-01-02", m[1], time.Local)
			if perr != nil {
				continue
			}
			matched = append(matched, datedFile{date: d, path: filepath.Join(logsDir, f.Name())})
		}

		sort.Slice(matched, func(i, j int) bool {
			return matched[i].date.After(matched[j].date)
		})

		for _, df := range matched {
			f, err := os.Open(df.path)
			if err != nil {
				continue
			}
			parsed, parseErr := parseFile(f)
			f.Close()
			if parseErr != nil {
				continue
			}
			for i := len(parsed) - 1; i >= 0; i-- {
				entries = append(entries, parsed[i])
				if len(entries) >= n {
					break
				}
			}
			if len(entries) >= n {
				break
			}
		}
	}

	if len(entries) > n {
		entries = entries[len(entries)-n:]
	}

	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries, nil
}

func parseFile(f *os.File) ([]Entry, error) {
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)

	var entries []Entry
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}

		if m := entryRegex.FindStringSubmatch(line); m != nil {
			ts, _ := time.ParseInLocation("2006-01-02 15:04:05", m[2], time.Local)
			var userID int64
			if m[4] != "" {
				userID, _ = strconv.ParseInt(m[4], 10, 64)
			}
			entries = append(entries, Entry{
				Time:    ts,
				Level:   m[1],
				UserID:  userID,
				Message: m[5],
				Raw:     line,
			})
			continue
		}

		if len(entries) > 0 {
			last := &entries[len(entries)-1]
			if last.Raw == "" {
				last.Raw = line
			} else {
				last.Raw += "\n" + line
			}
			if last.Message == "" {
				last.Message = line
			} else {
				last.Message += "\n" + line
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return entries, err
	}
	return entries, nil
}
