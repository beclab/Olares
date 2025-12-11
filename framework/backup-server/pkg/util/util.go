package util

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

const (
	DateFormat = "2006-01-02 15:04:05"
)

func TrimLineBreak(s string) string {
	return strings.TrimRight(s, "\r\n")
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

func DirSize(path string) (uint64, error) {
	var size int64 = 0
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	var res = uint64(size)

	return res, err
}

func GetDiskFreeSpace(path string) (uint64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}
	free := stat.Bavail * uint64(stat.Bsize)
	return free, nil
}

func Base64encode(s []byte) string {
	return base64.StdEncoding.EncodeToString(s)
}

func Base64decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func MD5(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func ParseToInt64(v string) int64 {
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

// ToJSON returns a json string
func ToJSON(v any) string {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		panic(err)
	}
	return buf.String()
}

// PrettyJSON returns a pretty formated json string
func PrettyJSON(v any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
	return buf.String()
}

// FilePathExists returns a boolean, whether the file or directory is exists
func FilePathExists(name string) bool {
	_, err := os.Stat(name)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// ListContains returns a boolean that v is in items
func ListContains[T comparable](items []T, v T) bool {
	if items == nil {
		return false
	}

	for _, item := range items {
		if v == item {
			return true
		}
	}
	return false
}

func ListMatchContains(items []string, v string) bool {
	if items == nil {
		return false
	}

	for _, item := range items {
		if strings.Contains(v, item) {
			return true
		}
	}
	return false
}

func BytesToMD5Hash(in []byte) string {
	hash := md5.Sum(in)
	return hex.EncodeToString(hash[:])
}

func BytesToSha256Hash(in []byte) string {
	h := sha256.New()
	h.Write(in)

	return fmt.Sprintf("%x", h.Sum(nil))
}

func EncodeStringToBase64(in string) string {
	return base64.StdEncoding.EncodeToString([]byte(in))
}

func DecodeBase64ToString(in string) string {
	ou, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return "'"
	}
	return string(ou)
}

func ParseToCron(frequency, timesOfDay string, dayOfWeek int, dateOfMonth int) (cron string, err error) {
	var (
		hourInt   int
		minuteInt int
	)

	splitTimes := strings.Split(timesOfDay, ":")
	if len(splitTimes) != 2 {
		return "", errors.Errorf("invalid times of day: %q", timesOfDay)
	}
	hour, minute := splitTimes[0], splitTimes[1]

	hourInt, err = strconv.Atoi(hour)
	if err != nil {
		return
	}
	minuteInt, err = strconv.Atoi(minute)
	if err != nil {
		return
	}
	if hourInt < 0 || hourInt > 24 {
		return "", errors.Errorf("invalid hour %q", hour)
	}
	if minuteInt < 0 || minuteInt > 59 {
		return "", errors.Errorf("invalid minute %q", minute)
	}

	minuteHour := minute + " " + hour

	switch frequency {
	case "@hourly":
		cron = fmt.Sprintf("*/%d * * * *", minuteInt)
	case "@daily":
		cron = minuteHour + " * * *"
	case "@weekly":
		cron = minuteHour + " * * " + strconv.Itoa(dayOfWeek-1)
	case "@monthly":
		cron = minuteHour + fmt.Sprintf(" %s * *", strconv.Itoa(dateOfMonth))
	case "@yearly":
		cron = minuteHour + " 1 1 *"
	}

	if cron == "" {
		err = fmt.Errorf("invalid frequency: %q or times of day: %q", frequency, timesOfDay)
	}
	return
}

func ParseTimestampToLocal(value string) (string, error) {
	var v, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		return "", err
	}

	var t = time.UnixMilli(v)

	var _, localoffset = time.Now().Zone()
	var utcTime = t.Add(time.Duration(localoffset) * time.Second)

	return fmt.Sprintf("%.2d:%.2d", utcTime.Hour(), utcTime.Minute()), nil
}

func ParseToNextUnixTime(frequency, timesOfDay string, dayOfWeek int, dateOfMonth int) int64 {
	switch frequency {
	case "@daily":
		return 86400
	case "@weekly":
		return 604800
	case "@monthly":
		return 2592000
	default:
		return 0
	}
}

func GetFirstDayOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	offset := -time.Duration(weekday-1) * 24 * time.Hour

	return t.Add(offset).Truncate(24 * time.Hour)
}

func IsTimestampNearingExpiration(targetTimestamp int64) bool {
	adjustedTimestamp := targetTimestamp - (15 * 1000)
	currentTimestamp := time.Now().UnixMilli()
	return adjustedTimestamp < currentTimestamp
}

func ParseUnixMilliToDate(targetTimestamp int64) string {
	t := time.UnixMilli(targetTimestamp)
	return t.Format(DateFormat)
}

func ParseTimeText(timeStr string) int64 {
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		return 0
	}

	return t.Unix()
}

func GetSuffix(c string, s string) (string, error) {
	var r = strings.Split(c, s)
	if len(r) != 2 {
		return "", fmt.Errorf("get suffix invalid, context: %s", c)
	}
	return r[1], nil
}

func ReplacePathPrefix(s string, prefix1 string, prefix2 string) (result string) {
	pos := strings.Index(s, prefix1)
	if pos >= 0 {
		result = strings.Replace(s, prefix1, "/Files", 1)
		return
	}

	result = strings.Replace(s, prefix2, "/Files.External", 1)
	return
}

func FormatBytes(bytes uint64) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
		TB = 1 << 40
	)

	var result string
	switch {
	case bytes >= TB:
		result = fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		result = fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		result = fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		result = fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		result = fmt.Sprintf("%d Byte", bytes)
	}

	return result
}

func TrimArrayPrefix(array []string, prefix string) []string {
	if len(array) == 0 {
		return []string{}
	}
	var result []string
	for _, v := range array {
		if strings.HasPrefix(v, prefix) {
			result = append(result, strings.TrimPrefix(v, prefix))
		}
	}
	return result
}

func CombineArray(args ...[]string) []string {
	var result []string
	for _, arg := range args {
		result = append(result, arg...)
	}

	return result
}

func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
