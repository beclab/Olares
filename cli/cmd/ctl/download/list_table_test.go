package download

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestClipDisplayCountsTerminalColumns(t *testing.T) {
	cjk := "军训演习：谁不想急头白脸地演一场匪徒！.mp4"
	if runewidth.StringWidth(cjk) <= taskTableNameWidth {
		t.Fatalf("fixture too short to clip: width=%d", runewidth.StringWidth(cjk))
	}
	got := clipDisplay(cjk, taskTableNameWidth)
	if w := runewidth.StringWidth(got); w > taskTableNameWidth {
		t.Fatalf("clipDisplay width %d, want <= %d (%q)", w, taskTableNameWidth, got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("clipped CJK = %q, want trailing ...", got)
	}
	ascii := "Black Myth: Zhong Kui – 15 Minutes Gameplay Showcase.mp4"
	got = clipDisplay(ascii, taskTableNameWidth)
	if w := runewidth.StringWidth(got); w > taskTableNameWidth {
		t.Fatalf("ascii clip width %d, want <= %d (%q)", w, taskTableNameWidth, got)
	}
}

func TestRenderTasksTableAlignsCJKAgainstASCII(t *testing.T) {
	var buf strings.Builder
	err := renderTasksTable(&buf, []DownloadTask{
		{
			ID:               24,
			Status:           "completed",
			DownloadProvider: "yt-dlp",
			Percent:          100,
			FileName:         "军训演习：谁不想急头白脸地演一场匪徒！.mp4",
			URL:              "https://www.bilibili.com/video/BV1qG8h68ES9/",
			App:              "wise",
		},
		{
			ID:               23,
			Status:           "error",
			DownloadProvider: "yt-dlp",
			Percent:          0,
			FileName:         "Black Myth: Zhong Kui – 15 Minutes Gameplay",
			URL:              "https://www.youtube.com/watch?v=oi2QgPH61JM",
			App:              "wise",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("lines = %d, want 3 (header + 2 rows):\n%s", len(lines), buf.String())
	}
	headerSrc := displayIndex(lines[0], "SOURCE")
	if headerSrc < 0 {
		t.Fatalf("header missing SOURCE: %q", lines[0])
	}
	for i, line := range lines[1:] {
		got := displayIndex(line, "https://")
		if got != headerSrc {
			t.Fatalf("row %d SOURCE starts at column %d, header at %d\n%s", i, got, headerSrc, buf.String())
		}
	}
}

func displayIndex(line, needle string) int {
	i := strings.Index(line, needle)
	if i < 0 {
		return -1
	}
	return runewidth.StringWidth(line[:i])
}
